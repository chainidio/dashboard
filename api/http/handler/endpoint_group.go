package handler

import (
	"github.com/chainid-io/dashboard"
	httperror "github.com/chainid-io/dashboard/http/error"
	"github.com/chainid-io/dashboard/http/security"

	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

// EndpointGroupHandler represents an HTTP API handler for managing endpoint groups.
type EndpointGroupHandler struct {
	*mux.Router
	Logger               *log.Logger
	EndpointService      chainid.EndpointService
	EndpointGroupService chainid.EndpointGroupService
}

// NewEndpointGroupHandler returns a new instance of EndpointGroupHandler.
func NewEndpointGroupHandler(bouncer *security.RequestBouncer) *EndpointGroupHandler {
	h := &EndpointGroupHandler{
		Router: mux.NewRouter(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	h.Handle("/endpoint_groups",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handlePostEndpointGroups))).Methods(http.MethodPost)
	h.Handle("/endpoint_groups",
		bouncer.RestrictedAccess(http.HandlerFunc(h.handleGetEndpointGroups))).Methods(http.MethodGet)
	h.Handle("/endpoint_groups/{id}",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handleGetEndpointGroup))).Methods(http.MethodGet)
	h.Handle("/endpoint_groups/{id}",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handlePutEndpointGroup))).Methods(http.MethodPut)
	h.Handle("/endpoint_groups/{id}/access",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handlePutEndpointGroupAccess))).Methods(http.MethodPut)
	h.Handle("/endpoint_groups/{id}",
		bouncer.AdministratorAccess(http.HandlerFunc(h.handleDeleteEndpointGroup))).Methods(http.MethodDelete)

	return h
}

type (
	postEndpointGroupsResponse struct {
		ID int `json:"Id"`
	}

	postEndpointGroupsRequest struct {
		Name                string                 `valid:"required"`
		Description         string                 `valid:"-"`
		Labels              []chainid.Pair       `valid:""`
		AssociatedEndpoints []chainid.EndpointID `valid:""`
	}

	putEndpointGroupAccessRequest struct {
		AuthorizedUsers []int `valid:"-"`
		AuthorizedTeams []int `valid:"-"`
	}

	putEndpointGroupsRequest struct {
		Name                string                 `valid:"-"`
		Description         string                 `valid:"-"`
		Labels              []chainid.Pair       `valid:""`
		AssociatedEndpoints []chainid.EndpointID `valid:""`
	}
)

// handleGetEndpointGroups handles GET requests on /endpoint_groups
func (handler *EndpointGroupHandler) handleGetEndpointGroups(w http.ResponseWriter, r *http.Request) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	endpointGroups, err := handler.EndpointGroupService.EndpointGroups()
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	filteredEndpointGroups, err := security.FilterEndpointGroups(endpointGroups, securityContext)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	encodeJSON(w, filteredEndpointGroups, handler.Logger)
}

// handlePostEndpointGroups handles POST requests on /endpoint_groups
func (handler *EndpointGroupHandler) handlePostEndpointGroups(w http.ResponseWriter, r *http.Request) {
	var req postEndpointGroupsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidJSON, http.StatusBadRequest, handler.Logger)
		return
	}

	_, err := govalidator.ValidateStruct(req)
	if err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidRequestFormat, http.StatusBadRequest, handler.Logger)
		return
	}

	endpointGroup := &chainid.EndpointGroup{
		Name:            req.Name,
		Description:     req.Description,
		Labels:          req.Labels,
		AuthorizedUsers: []chainid.UserID{},
		AuthorizedTeams: []chainid.TeamID{},
	}

	err = handler.EndpointGroupService.CreateEndpointGroup(endpointGroup)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	endpoints, err := handler.EndpointService.Endpoints()
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	for _, endpoint := range endpoints {
		if endpoint.GroupID == chainid.EndpointGroupID(1) {
			err = handler.checkForGroupAssignment(endpoint, endpointGroup.ID, req.AssociatedEndpoints)
			if err != nil {
				httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
				return
			}
		}
	}

	encodeJSON(w, &postEndpointGroupsResponse{ID: int(endpointGroup.ID)}, handler.Logger)
}

// handleGetEndpointGroup handles GET requests on /endpoint_groups/:id
func (handler *EndpointGroupHandler) handleGetEndpointGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	endpointGroupID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	endpointGroup, err := handler.EndpointGroupService.EndpointGroup(chainid.EndpointGroupID(endpointGroupID))
	if err == chainid.ErrEndpointGroupNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	encodeJSON(w, endpointGroup, handler.Logger)
}

// handlePutEndpointGroupAccess handles PUT requests on /endpoint_groups/:id/access
func (handler *EndpointGroupHandler) handlePutEndpointGroupAccess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	endpointGroupID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	var req putEndpointGroupAccessRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidJSON, http.StatusBadRequest, handler.Logger)
		return
	}

	_, err = govalidator.ValidateStruct(req)
	if err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidRequestFormat, http.StatusBadRequest, handler.Logger)
		return
	}

	endpointGroup, err := handler.EndpointGroupService.EndpointGroup(chainid.EndpointGroupID(endpointGroupID))
	if err == chainid.ErrEndpointGroupNotFound {
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
		endpointGroup.AuthorizedUsers = authorizedUserIDs
	}

	if req.AuthorizedTeams != nil {
		authorizedTeamIDs := []chainid.TeamID{}
		for _, value := range req.AuthorizedTeams {
			authorizedTeamIDs = append(authorizedTeamIDs, chainid.TeamID(value))
		}
		endpointGroup.AuthorizedTeams = authorizedTeamIDs
	}

	err = handler.EndpointGroupService.UpdateEndpointGroup(endpointGroup.ID, endpointGroup)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}
}

// handlePutEndpointGroup handles PUT requests on /endpoint_groups/:id
func (handler *EndpointGroupHandler) handlePutEndpointGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	endpointGroupID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	var req putEndpointGroupsRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidJSON, http.StatusBadRequest, handler.Logger)
		return
	}

	_, err = govalidator.ValidateStruct(req)
	if err != nil {
		httperror.WriteErrorResponse(w, ErrInvalidRequestFormat, http.StatusBadRequest, handler.Logger)
		return
	}

	groupID := chainid.EndpointGroupID(endpointGroupID)
	endpointGroup, err := handler.EndpointGroupService.EndpointGroup(groupID)
	if err == chainid.ErrEndpointGroupNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	if req.Name != "" {
		endpointGroup.Name = req.Name
	}

	if req.Description != "" {
		endpointGroup.Description = req.Description
	}

	endpointGroup.Labels = req.Labels

	err = handler.EndpointGroupService.UpdateEndpointGroup(endpointGroup.ID, endpointGroup)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	endpoints, err := handler.EndpointService.Endpoints()
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	for _, endpoint := range endpoints {
		err = handler.updateEndpointGroup(endpoint, groupID, req.AssociatedEndpoints)
		if err != nil {
			httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
			return
		}
	}
}

func (handler *EndpointGroupHandler) updateEndpointGroup(endpoint chainid.Endpoint, groupID chainid.EndpointGroupID, associatedEndpoints []chainid.EndpointID) error {
	if endpoint.GroupID == groupID {
		return handler.checkForGroupUnassignment(endpoint, associatedEndpoints)
	} else if endpoint.GroupID == chainid.EndpointGroupID(1) {
		return handler.checkForGroupAssignment(endpoint, groupID, associatedEndpoints)
	}
	return nil
}

func (handler *EndpointGroupHandler) checkForGroupUnassignment(endpoint chainid.Endpoint, associatedEndpoints []chainid.EndpointID) error {
	for _, id := range associatedEndpoints {
		if id == endpoint.ID {
			return nil
		}
	}

	endpoint.GroupID = chainid.EndpointGroupID(1)
	return handler.EndpointService.UpdateEndpoint(endpoint.ID, &endpoint)
}

func (handler *EndpointGroupHandler) checkForGroupAssignment(endpoint chainid.Endpoint, groupID chainid.EndpointGroupID, associatedEndpoints []chainid.EndpointID) error {
	for _, id := range associatedEndpoints {

		if id == endpoint.ID {
			endpoint.GroupID = groupID
			return handler.EndpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		}
	}
	return nil
}

// handleDeleteEndpointGroup handles DELETE requests on /endpoint_groups/:id
func (handler *EndpointGroupHandler) handleDeleteEndpointGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	endpointGroupID, err := strconv.Atoi(id)
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusBadRequest, handler.Logger)
		return
	}

	if endpointGroupID == 1 {
		httperror.WriteErrorResponse(w, chainid.ErrCannotRemoveDefaultGroup, http.StatusForbidden, handler.Logger)
		return
	}

	groupID := chainid.EndpointGroupID(endpointGroupID)
	_, err = handler.EndpointGroupService.EndpointGroup(groupID)
	if err == chainid.ErrEndpointGroupNotFound {
		httperror.WriteErrorResponse(w, err, http.StatusNotFound, handler.Logger)
		return
	} else if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	err = handler.EndpointGroupService.DeleteEndpointGroup(chainid.EndpointGroupID(endpointGroupID))
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	endpoints, err := handler.EndpointService.Endpoints()
	if err != nil {
		httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
		return
	}

	for _, endpoint := range endpoints {
		if endpoint.GroupID == groupID {
			endpoint.GroupID = chainid.EndpointGroupID(1)
			err = handler.EndpointService.UpdateEndpoint(endpoint.ID, &endpoint)
			if err != nil {
				httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, handler.Logger)
				return
			}
		}
	}
}
