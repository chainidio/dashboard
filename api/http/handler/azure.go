package handler

import (
	"strconv"

	"github.com/chainid-io/dashboard"
	httperror "github.com/chainid-io/dashboard/http/error"
	"github.com/chainid-io/dashboard/http/proxy"
	"github.com/chainid-io/dashboard/http/security"

	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// AzureHandler represents an HTTP API handler for proxying requests to the Azure API.
type AzureHandler struct {
	*mux.Router
	Logger                *log.Logger
	EndpointService       chainid.EndpointService
	EndpointGroupService  chainid.EndpointGroupService
	TeamMembershipService chainid.TeamMembershipService
	ProxyManager          *proxy.Manager
}

// NewAzureHandler returns a new instance of AzureHandler.
func NewAzureHandler(bouncer *security.RequestBouncer) *AzureHandler {
	h := &AzureHandler{
		Router: mux.NewRouter(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	h.PathPrefix("/{id}/azure").Handler(
		bouncer.AuthenticatedAccess(http.HandlerFunc(h.proxyRequestsToAzureAPI)))
	return h
}

func (handler *AzureHandler) checkEndpointAccess(endpoint *chainid.Endpoint, userID chainid.UserID) error {
	memberships, err := handler.TeamMembershipService.TeamMembershipsByUserID(userID)
	if err != nil {
		return err
	}

	group, err := handler.EndpointGroupService.EndpointGroup(endpoint.GroupID)
	if err != nil {
		return err
	}

	if !security.AuthorizedEndpointAccess(endpoint, group, userID, memberships) {
		return chainid.ErrEndpointAccessDenied
	}

	return nil
}

func (handler *AzureHandler) proxyRequestsToAzureAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	parsedID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	endpointID := chainid.EndpointID(parsedID)
	endpoint, err := handler.EndpointService.Endpoint(endpointID)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	if tokenData.Role != chainid.AdministratorRole {
		err = handler.checkEndpointAccess(endpoint, tokenData.ID)
		if err != nil && err == chainid.ErrEndpointAccessDenied {
			httperror.WriteErrorResponse(w, err, http.StatusForbidden, handler.Logger)
			return
		} else if err != nil {
			httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
			return
		}
	}

	var proxy http.Handler
	proxy = handler.ProxyManager.GetProxy(string(endpointID))
	if proxy == nil {
		proxy, err = handler.ProxyManager.CreateAndRegisterProxy(endpoint)
		if err != nil {
			httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
			return
		}
	}

	http.StripPrefix("/"+id+"/azure", proxy).ServeHTTP(w, r)
}
