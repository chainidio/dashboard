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

// DockerHandler represents an HTTP API handler for proxying requests to the Docker API.
type DockerHandler struct {
	*mux.Router
	Logger                *log.Logger
	EndpointService       chainid.EndpointService
	EndpointGroupService  chainid.EndpointGroupService
	TeamMembershipService chainid.TeamMembershipService
	ProxyManager          *proxy.Manager
}

// NewDockerHandler returns a new instance of DockerHandler.
func NewDockerHandler(bouncer *security.RequestBouncer) *DockerHandler {
	h := &DockerHandler{
		Router: mux.NewRouter(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	h.PathPrefix("/{id}/docker").Handler(
		bouncer.AuthenticatedAccess(http.HandlerFunc(h.proxyRequestsToDockerAPI)))
	return h
}

func (handler *DockerHandler) proxyRequestsToDockerAPI(w http.ResponseWriter, r *http.Request) {
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

	memberships, err := handler.TeamMembershipService.TeamMembershipsByUserID(tokenData.ID)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	if tokenData.Role != chainid.AdministratorRole {
		group, err := handler.EndpointGroupService.EndpointGroup(endpoint.GroupID)
		if err != nil {
			httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
			return
		}

		if !security.AuthorizedEndpointAccess(endpoint, group, tokenData.ID, memberships) {
			httperror.WriteErrorResponse(w, chainid.ErrEndpointAccessDenied, http.StatusForbidden, handler.Logger)
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

	http.StripPrefix("/"+id+"/docker", proxy).ServeHTTP(w, r)
}
