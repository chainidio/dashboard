package security

import (
	"github.com/chainid-io/dashboard"
	httperror "github.com/chainid-io/dashboard/http/error"

	"net/http"
	"strings"
)

type (
	// RequestBouncer represents an entity that manages API request accesses
	RequestBouncer struct {
		jwtService            chainid.JWTService
		userService           chainid.UserService
		teamMembershipService chainid.TeamMembershipService
		authDisabled          bool
	}

	// RestrictedRequestContext is a data structure containing information
	// used in RestrictedAccess
	RestrictedRequestContext struct {
		IsAdmin         bool
		IsTeamLeader    bool
		UserID          chainid.UserID
		UserMemberships []chainid.TeamMembership
	}
)

// NewRequestBouncer initializes a new RequestBouncer
func NewRequestBouncer(jwtService chainid.JWTService, userService chainid.UserService, teamMembershipService chainid.TeamMembershipService, authDisabled bool) *RequestBouncer {
	return &RequestBouncer{
		jwtService:            jwtService,
		userService:           userService,
		teamMembershipService: teamMembershipService,
		authDisabled:          authDisabled,
	}
}

// PublicAccess defines a security check for public endpoints.
// No authentication is required to access these endpoints.
func (bouncer *RequestBouncer) PublicAccess(h http.Handler) http.Handler {
	h = mwSecureHeaders(h)
	return h
}

// AuthenticatedAccess defines a security check for private endpoints.
// Authentication is required to access these endpoints.
func (bouncer *RequestBouncer) AuthenticatedAccess(h http.Handler) http.Handler {
	h = bouncer.mwCheckAuthentication(h)
	h = mwSecureHeaders(h)
	return h
}

// RestrictedAccess defines a security check for restricted endpoints.
// Authentication is required to access these endpoints.
// The request context will be enhanced with a RestrictedRequestContext object
// that might be used later to authorize/filter access to resources.
func (bouncer *RequestBouncer) RestrictedAccess(h http.Handler) http.Handler {
	h = bouncer.mwUpgradeToRestrictedRequest(h)
	h = bouncer.AuthenticatedAccess(h)
	return h
}

// AdministratorAccess defines a chain of middleware for restricted endpoints.
// Authentication as well as administrator role are required to access these endpoints.
func (bouncer *RequestBouncer) AdministratorAccess(h http.Handler) http.Handler {
	h = mwCheckAdministratorRole(h)
	h = bouncer.AuthenticatedAccess(h)
	return h
}

// mwSecureHeaders provides secure headers middleware for handlers.
func mwSecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// mwUpgradeToRestrictedRequest will enhance the current request with
// a new RestrictedRequestContext object.
func (bouncer *RequestBouncer) mwUpgradeToRestrictedRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenData, err := RetrieveTokenData(r)
		if err != nil {
			httperror.WriteErrorResponse(w, chainid.ErrResourceAccessDenied, http.StatusForbidden, nil)
			return
		}

		requestContext, err := bouncer.newRestrictedContextRequest(tokenData.ID, tokenData.Role)
		if err != nil {
			httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, nil)
			return
		}

		ctx := storeRestrictedRequestContext(r, requestContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// mwCheckAdministratorRole check the role of the user associated to the request
func mwCheckAdministratorRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenData, err := RetrieveTokenData(r)
		if err != nil || tokenData.Role != chainid.AdministratorRole {
			httperror.WriteErrorResponse(w, chainid.ErrResourceAccessDenied, http.StatusForbidden, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// mwCheckAuthentication provides Authentication middleware for handlers
func (bouncer *RequestBouncer) mwCheckAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenData *chainid.TokenData
		if !bouncer.authDisabled {
			var token string

			// Get token from the Authorization header
			tokens, ok := r.Header["Authorization"]
			if ok && len(tokens) >= 1 {
				token = tokens[0]
				token = strings.TrimPrefix(token, "Bearer ")
			}

			if token == "" {
				httperror.WriteErrorResponse(w, chainid.ErrUnauthorized, http.StatusUnauthorized, nil)
				return
			}

			var err error
			tokenData, err = bouncer.jwtService.ParseAndVerifyToken(token)
			if err != nil {
				httperror.WriteErrorResponse(w, err, http.StatusUnauthorized, nil)
				return
			}

			_, err = bouncer.userService.User(tokenData.ID)
			if err != nil && err == chainid.ErrUserNotFound {
				httperror.WriteErrorResponse(w, chainid.ErrUnauthorized, http.StatusUnauthorized, nil)
				return
			} else if err != nil {
				httperror.WriteErrorResponse(w, err, http.StatusInternalServerError, nil)
				return
			}
		} else {
			tokenData = &chainid.TokenData{
				Role: chainid.AdministratorRole,
			}
		}

		ctx := storeTokenData(r, tokenData)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

func (bouncer *RequestBouncer) newRestrictedContextRequest(userID chainid.UserID, userRole chainid.UserRole) (*RestrictedRequestContext, error) {
	requestContext := &RestrictedRequestContext{
		IsAdmin: true,
		UserID:  userID,
	}

	if userRole != chainid.AdministratorRole {
		requestContext.IsAdmin = false
		memberships, err := bouncer.teamMembershipService.TeamMembershipsByUserID(userID)
		if err != nil {
			return nil, err
		}

		isTeamLeader := false
		for _, membership := range memberships {
			if membership.Role == chainid.TeamLeader {
				isTeamLeader = true
			}
		}

		requestContext.IsTeamLeader = isTeamLeader
		requestContext.UserMemberships = memberships
	}

	return requestContext, nil
}
