package handler

import (
	"encoding/json"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/chainid-io/dashboard"
	httperror "github.com/chainid-io/dashboard/http/error"
	"github.com/chainid-io/dashboard/http/proxy"
	"github.com/chainid-io/dashboard/http/security"

	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// ExtensionHandler represents an HTTP API handler for managing Settings.
type ExtensionHandler struct {
	*mux.Router
	Logger          *log.Logger
	EndpointService chainid.EndpointService
	ProxyManager    *proxy.Manager
}

// NewExtensionHandler returns a new instance of ExtensionHandler.
func NewExtensionHandler(bouncer *security.RequestBouncer) *ExtensionHandler {
	h := &ExtensionHandler{
		Router: mux.NewRouter(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	h.Handle("/{endpointId}/extensions",
		bouncer.AuthenticatedAccess(http.HandlerFunc(h.handlePostExtensions))).Methods(http.MethodPost)
	h.Handle("/{endpointId}/extensions/{extensionType}",
		bouncer.AuthenticatedAccess(http.HandlerFunc(h.handleDeleteExtensions))).Methods(http.MethodDelete)
	return h
}

type (
	postExtensionRequest struct {
		Type int    `valid:"required"`
		URL  string `valid:"required"`
	}
)

func (handler *ExtensionHandler) handlePostExtensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["endpointId"])
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}
	endpointID := chainid.EndpointID(id)

	endpoint, err := handler.EndpointService.Endpoint(endpointID)
	if err == chainid.ErrEndpointNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	var req postExtensionRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidJSON, http.StatusBadRequest, handler.Logger)
		return
	}

	_, err = govalidator.ValidateStruct(req)
	if err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidRequestFormat, http.StatusBadRequest, handler.Logger)
		return
	}

	extensionType := chainid.EndpointExtensionType(req.Type)

	var extension *chainid.EndpointExtension

	for _, ext := range endpoint.Extensions {
		if ext.Type == extensionType {
			extension = &ext
		}
	}

	if extension != nil {
		extension.URL = req.URL
	} else {
		extension = &chainid.EndpointExtension{
			Type: extensionType,
			URL:  req.URL,
		}
		endpoint.Extensions = append(endpoint.Extensions, *extension)
	}

	err = handler.EndpointService.UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	encodeJSON(w, extension, handler.Logger)
}

func (handler *ExtensionHandler) handleDeleteExtensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["endpointId"])
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}
	endpointID := chainid.EndpointID(id)

	endpoint, err := handler.EndpointService.Endpoint(endpointID)
	if err == chainid.ErrEndpointNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	extType, err := strconv.Atoi(vars["extensionType"])
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}
	extensionType := chainid.EndpointExtensionType(extType)

	for idx, ext := range endpoint.Extensions {
		if ext.Type == extensionType {
			endpoint.Extensions = append(endpoint.Extensions[:idx], endpoint.Extensions[idx+1:]...)
		}
	}

	err = handler.EndpointService.UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}
}
