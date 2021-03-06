package handler

import (
	"bytes"
	"strings"

	"github.com/chainid-io/dashboard"
	"github.com/chainid-io/dashboard/crypto"
	"github.com/chainid-io/dashboard/http/client"
	httperror "github.com/chainid-io/dashboard/http/error"
	"github.com/chainid-io/dashboard/http/proxy"
	"github.com/chainid-io/dashboard/http/security"

	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

// EndpointHandler represents an HTTP API handler for managing Docker endpoints.
type EndpointHandler struct {
	*mux.Router
	Logger                      *log.Logger
	authorizeEndpointManagement bool
	EndpointService             chainid.EndpointService
	EndpointGroupService        chainid.EndpointGroupService
	FileService                 chainid.FileService
	ProxyManager                *proxy.Manager
}

const (
	// ErrEndpointManagementDisabled is an error raised when trying to access the endpoints management endpoints
	// when the server has been started with the --external-endpoints flag
	ErrEndpointManagementDisabled = chainid.Error("Endpoint management is disabled")
)

// NewEndpointHandler returns a new instance of EndpointHandler.
func NewEndpointHandler(bouncer *security.RequestBouncer, authorizeEndpointManagement bool) *EndpointHandler {
	h := &EndpointHandler{
		Router: mux.NewRouter(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
		authorizeEndpointManagement: authorizeEndpointManagement,
	}
	h.Handle("/endpoints",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handlePostEndpoints))).Methods(http.MethodPost)
	h.Handle("/endpoints",
		bouncer.RestrictedAccess(http.HandlerFunc(h.handleGetEndpoints))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handleGetEndpoint))).Methods(http.MethodGet)
	h.Handle("/endpoints/{id}",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handlePutEndpoint))).Methods(http.MethodPut)
	h.Handle("/endpoints/{id}/access",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handlePutEndpointAccess))).Methods(http.MethodPut)
	h.Handle("/endpoints/{id}",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handleDeleteEndpoint))).Methods(http.MethodDelete)

	return h
}

type (
	putEndpointAccessRequest struct {
		AuthorizedUsers []int `valid:"-"`
		AuthorizedTeams []int `valid:"-"`
	}

	putEndpointsRequest struct {
		Name                string `valid:"-"`
		URL                 string `valid:"-"`
		PublicURL           string `valid:"-"`
		GroupID             int    `valid:"-"`
		TLS                 bool   `valid:"-"`
		TLSSkipVerify       bool   `valid:"-"`
		TLSSkipClientVerify bool   `valid:"-"`
	}

	postEndpointPayload struct {
		name                      string
		url                       string
		endpointType              int
		publicURL                 string
		groupID                   int
		useTLS                    bool
		skipTLSServerVerification bool
		skipTLSClientVerification bool
		caCert                    []byte
		cert                      []byte
		key                       []byte
		azureApplicationID        string
		azureTenantID             string
		azureAuthenticationKey    string
	}
)

// handleGetEndpoints handles GET requests on /endpoints
func (handler *EndpointHandler) handleGetEndpoints(w http.ResponseWriter, r *http.Request) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	endpoints, err := handler.EndpointService.Endpoints()
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	groups, err := handler.EndpointGroupService.EndpointGroups()
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	filteredEndpoints, err := security.FilterEndpoints(endpoints, groups, securityContext)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	for i := range filteredEndpoints {
		filteredEndpoints[i].AzureCredentials = chainid.AzureCredentials{}
	}

	encodeJSON(w, filteredEndpoints, handler.Logger)
}

func (handler *EndpointHandler) createAzureEndpoint(payload *postEndpointPayload) (*chainid.Endpoint, error) {
	credentials := chainid.AzureCredentials{
		ApplicationID:     payload.azureApplicationID,
		TenantID:          payload.azureTenantID,
		AuthenticationKey: payload.azureAuthenticationKey,
	}

	httpClient := client.NewHTTPClient()
	_, err := httpClient.ExecuteAzureAuthenticationRequest(&credentials)
	if err != nil {
		return nil, err
	}

	endpoint := &chainid.Endpoint{
		Name:             payload.name,
		URL:              payload.url,
		Type:             chainid.AzureEnvironment,
		GroupID:          chainid.EndpointGroupID(payload.groupID),
		PublicURL:        payload.publicURL,
		AuthorizedUsers:  []chainid.UserID{},
		AuthorizedTeams:  []chainid.TeamID{},
		Extensions:       []chainid.EndpointExtension{},
		AzureCredentials: credentials,
	}

	err = handler.EndpointService.CreateEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *EndpointHandler) createTLSSecuredEndpoint(payload *postEndpointPayload) (*chainid.Endpoint, error) {
	tlsConfig, err := crypto.CreateTLSConfigurationFromBytes(payload.caCert, payload.cert, payload.key, payload.skipTLSClientVerification, payload.skipTLSServerVerification)
	if err != nil {
		return nil, err
	}

	agentOnDockerEnvironment, err := client.ExecutePingOperation(payload.url, tlsConfig)
	if err != nil {
		return nil, err
	}

	endpointType := chainid.DockerEnvironment
	if agentOnDockerEnvironment {
		endpointType = chainid.AgentOnDockerEnvironment
	}

	endpoint := &chainid.Endpoint{
		Name:      payload.name,
		URL:       payload.url,
		Type:      endpointType,
		GroupID:   chainid.EndpointGroupID(payload.groupID),
		PublicURL: payload.publicURL,
		TLSConfig: chainid.TLSConfiguration{
			TLS:           payload.useTLS,
			TLSSkipVerify: payload.skipTLSServerVerification,
		},
		AuthorizedUsers: []chainid.UserID{},
		AuthorizedTeams: []chainid.TeamID{},
		Extensions:      []chainid.EndpointExtension{},
	}

	err = handler.EndpointService.CreateEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	folder := strconv.Itoa(int(endpoint.ID))

	if !payload.skipTLSServerVerification {
		r := bytes.NewReader(payload.caCert)
		// TODO: review the API exposed by the FileService to store
		// a file from a byte slice and return the path to the stored file instead
		// of using multiple legacy calls (StoreTLSFile, GetPathForTLSFile) here.
		err = handler.FileService.StoreTLSFile(folder, chainid.TLSFileCA, r)
		if err != nil {
			handler.EndpointService.DeleteEndpoint(endpoint.ID)
			return nil, err
		}
		caCertPath, _ := handler.FileService.GetPathForTLSFile(folder, chainid.TLSFileCA)
		endpoint.TLSConfig.TLSCACertPath = caCertPath
	}

	if !payload.skipTLSClientVerification {
		r := bytes.NewReader(payload.cert)
		err = handler.FileService.StoreTLSFile(folder, chainid.TLSFileCert, r)
		if err != nil {
			handler.EndpointService.DeleteEndpoint(endpoint.ID)
			return nil, err
		}
		certPath, _ := handler.FileService.GetPathForTLSFile(folder, chainid.TLSFileCert)
		endpoint.TLSConfig.TLSCertPath = certPath

		r = bytes.NewReader(payload.key)
		err = handler.FileService.StoreTLSFile(folder, chainid.TLSFileKey, r)
		if err != nil {
			handler.EndpointService.DeleteEndpoint(endpoint.ID)
			return nil, err
		}
		keyPath, _ := handler.FileService.GetPathForTLSFile(folder, chainid.TLSFileKey)
		endpoint.TLSConfig.TLSKeyPath = keyPath
	}

	err = handler.EndpointService.UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *EndpointHandler) createUnsecuredEndpoint(payload *postEndpointPayload) (*chainid.Endpoint, error) {
	endpointType := chainid.DockerEnvironment

	if !strings.HasPrefix(payload.url, "unix://") {
		agentOnDockerEnvironment, err := client.ExecutePingOperation(payload.url, nil)
		if err != nil {
			return nil, err
		}
		if agentOnDockerEnvironment {
			endpointType = chainid.AgentOnDockerEnvironment
		}
	}

	endpoint := &chainid.Endpoint{
		Name:      payload.name,
		URL:       payload.url,
		Type:      endpointType,
		GroupID:   chainid.EndpointGroupID(payload.groupID),
		PublicURL: payload.publicURL,
		TLSConfig: chainid.TLSConfiguration{
			TLS: false,
		},
		AuthorizedUsers: []chainid.UserID{},
		AuthorizedTeams: []chainid.TeamID{},
		Extensions:      []chainid.EndpointExtension{},
	}

	err := handler.EndpointService.CreateEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (handler *EndpointHandler) createEndpoint(payload *postEndpointPayload) (*chainid.Endpoint, error) {
	if chainid.EndpointType(payload.endpointType) == chainid.AzureEnvironment {
		return handler.createAzureEndpoint(payload)
	}

	if payload.useTLS {
		return handler.createTLSSecuredEndpoint(payload)
	}
	return handler.createUnsecuredEndpoint(payload)
}

func convertPostEndpointRequestToPayload(r *http.Request) (*postEndpointPayload, error) {
	payload := &postEndpointPayload{}
	payload.name = r.FormValue("Name")

	endpointType := r.FormValue("EndpointType")

	if payload.name == "" || endpointType == "" {
		return nil, ErrInvalidRequestFormat
	}

	parsedType, err := strconv.Atoi(endpointType)
	if err != nil {
		return nil, err
	}

	payload.url = r.FormValue("URL")
	payload.endpointType = parsedType

	if chainid.EndpointType(payload.endpointType) != chainid.AzureEnvironment && payload.url == "" {
		return nil, ErrInvalidRequestFormat
	}

	payload.publicURL = r.FormValue("PublicURL")

	if chainid.EndpointType(payload.endpointType) == chainid.AzureEnvironment {
		payload.azureApplicationID = r.FormValue("AzureApplicationID")
		payload.azureTenantID = r.FormValue("AzureTenantID")
		payload.azureAuthenticationKey = r.FormValue("AzureAuthenticationKey")

		if payload.azureApplicationID == "" || payload.azureTenantID == "" || payload.azureAuthenticationKey == "" {
			return nil, ErrInvalidRequestFormat
		}
	}

	rawGroupID := r.FormValue("GroupID")
	if rawGroupID == "" {
		payload.groupID = 1
	} else {
		groupID, err := strconv.Atoi(rawGroupID)
		if err != nil {
			return nil, err
		}
		payload.groupID = groupID
	}

	payload.useTLS = r.FormValue("TLS") == "true"

	if payload.useTLS {
		payload.skipTLSServerVerification = r.FormValue("TLSSkipVerify") == "true"
		payload.skipTLSClientVerification = r.FormValue("TLSSkipClientVerify") == "true"

		if !payload.skipTLSServerVerification {
			caCert, err := getUploadedFileContent(r, "TLSCACertFile")
			if err != nil {
				return nil, err
			}
			payload.caCert = caCert
		}

		if !payload.skipTLSClientVerification {
			cert, err := getUploadedFileContent(r, "TLSCertFile")
			if err != nil {
				return nil, err
			}
			payload.cert = cert
			key, err := getUploadedFileContent(r, "TLSKeyFile")
			if err != nil {
				return nil, err
			}
			payload.key = key
		}
	}

	return payload, nil
}

// handlePostEndpoints handles POST requests on /endpoints
func (handler *EndpointHandler) handlePostEndpoints(w http.ResponseWriter, r *http.Request) {
	if !handler.authorizeEndpointManagement {
		httperror.WriteErrorResponse(w, ErrEndpointManagementDisabled, http.StatusServiceUnavailable, handler.Logger)
		return
	}

	payload, err := convertPostEndpointRequestToPayload(r)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	endpoint, err := handler.createEndpoint(payload)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	encodeJSON(w, &endpoint, handler.Logger)
}

// handleGetEndpoint handles GET requests on /endpoints/:id
func (handler *EndpointHandler) handleGetEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	endpointID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	endpoint, err := handler.EndpointService.Endpoint(chainid.EndpointID(endpointID))
	if err == chainid.ErrEndpointNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	endpoint.AzureCredentials = chainid.AzureCredentials{}

	encodeJSON(w, endpoint, handler.Logger)
}

// handlePutEndpointAccess handles PUT requests on /endpoints/:id/access
func (handler *EndpointHandler) handlePutEndpointAccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	endpointID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	var req putEndpointAccessRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidJSON, http.StatusBadRequest, handler.Logger)
		return
	}

	_, err = govalidator.ValidateStruct(req)
	if err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidRequestFormat, http.StatusBadRequest, handler.Logger)
		return
	}

	endpoint, err := handler.EndpointService.Endpoint(chainid.EndpointID(endpointID))
	if err == chainid.ErrEndpointNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	if req.AuthorizedUsers != nil {
		authorizedUserIDs := []chainid.UserID{}
		for _, value := range req.AuthorizedUsers {
			authorizedUserIDs = append(authorizedUserIDs, chainid.UserID(value))
		}
		endpoint.AuthorizedUsers = authorizedUserIDs
	}

	if req.AuthorizedTeams != nil {
		authorizedTeamIDs := []chainid.TeamID{}
		for _, value := range req.AuthorizedTeams {
			authorizedTeamIDs = append(authorizedTeamIDs, chainid.TeamID(value))
		}
		endpoint.AuthorizedTeams = authorizedTeamIDs
	}

	err = handler.EndpointService.UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}
}

// handlePutEndpoint handles PUT requests on /endpoints/:id
func (handler *EndpointHandler) handlePutEndpoint(w http.ResponseWriter, r *http.Request) {
	if !handler.authorizeEndpointManagement {
		httperror.WriteErrorResponse(w, ErrEndpointManagementDisabled, http.StatusServiceUnavailable, handler.Logger)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	endpointID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	var req putEndpointsRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidJSON, http.StatusBadRequest, handler.Logger)
		return
	}

	_, err = govalidator.ValidateStruct(req)
	if err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidRequestFormat, http.StatusBadRequest, handler.Logger)
		return
	}

	endpoint, err := handler.EndpointService.Endpoint(chainid.EndpointID(endpointID))
	if err == chainid.ErrEndpointNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	if req.Name != "" {
		endpoint.Name = req.Name
	}

	if req.URL != "" {
		endpoint.URL = req.URL
	}

	if req.PublicURL != "" {
		endpoint.PublicURL = req.PublicURL
	}

	if req.GroupID != 0 {
		endpoint.GroupID = chainid.EndpointGroupID(req.GroupID)
	}

	folder := strconv.Itoa(int(endpoint.ID))
	if req.TLS {
		endpoint.TLSConfig.TLS = true
		endpoint.TLSConfig.TLSSkipVerify = req.TLSSkipVerify
		if !req.TLSSkipVerify {
			caCertPath, _ := handler.FileService.GetPathForTLSFile(folder, chainid.TLSFileCA)
			endpoint.TLSConfig.TLSCACertPath = caCertPath
		} else {
			endpoint.TLSConfig.TLSCACertPath = ""
			handler.FileService.DeleteTLSFile(folder, chainid.TLSFileCA)
		}

		if !req.TLSSkipClientVerify {
			certPath, _ := handler.FileService.GetPathForTLSFile(folder, chainid.TLSFileCert)
			endpoint.TLSConfig.TLSCertPath = certPath
			keyPath, _ := handler.FileService.GetPathForTLSFile(folder, chainid.TLSFileKey)
			endpoint.TLSConfig.TLSKeyPath = keyPath
		} else {
			endpoint.TLSConfig.TLSCertPath = ""
			handler.FileService.DeleteTLSFile(folder, chainid.TLSFileCert)
			endpoint.TLSConfig.TLSKeyPath = ""
			handler.FileService.DeleteTLSFile(folder, chainid.TLSFileKey)
		}
	} else {
		endpoint.TLSConfig.TLS = false
		endpoint.TLSConfig.TLSSkipVerify = false
		endpoint.TLSConfig.TLSCACertPath = ""
		endpoint.TLSConfig.TLSCertPath = ""
		endpoint.TLSConfig.TLSKeyPath = ""
		err = handler.FileService.DeleteTLSFiles(folder)
		if err != nil {
			httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
			return
		}
	}

	_, err = handler.ProxyManager.CreateAndRegisterProxy(endpoint)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	err = handler.EndpointService.UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}
}

// handleDeleteEndpoint handles DELETE requests on /endpoints/:id
func (handler *EndpointHandler) handleDeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	if !handler.authorizeEndpointManagement {
		httperror.WriteErrorResponse(w, ErrEndpointManagementDisabled, http.StatusServiceUnavailable, handler.Logger)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	endpointID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	endpoint, err := handler.EndpointService.Endpoint(chainid.EndpointID(endpointID))

	if err == chainid.ErrEndpointNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	handler.ProxyManager.DeleteProxy(string(endpointID))
	handler.ProxyManager.DeleteExtensionProxies(string(endpointID))

	err = handler.EndpointService.DeleteEndpoint(chainid.EndpointID(endpointID))
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	if endpoint.TLSConfig.TLS {
		err = handler.FileService.DeleteTLSFiles(id)
		if err != nil {
			httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
			return
		}
	}
}
