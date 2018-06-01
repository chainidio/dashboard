package security

import "github.com/chainid-io/dashboard"

// FilterUserTeams filters teams based on user role.
// non-administrator users only have access to team they are member of.
func FilterUserTeams(teams []chainid.Team, context *RestrictedRequestContext) []chainid.Team {
	filteredTeams := teams

	if !context.IsAdmin {
		filteredTeams = make([]chainid.Team, 0)
		for _, membership := range context.UserMemberships {
			for _, team := range teams {
				if team.ID == membership.TeamID {
					filteredTeams = append(filteredTeams, team)
					break
				}
			}
		}
	}

	return filteredTeams
}

// FilterLeaderTeams filters teams based on user role.
// Team leaders only have access to team they lead.
func FilterLeaderTeams(teams []chainid.Team, context *RestrictedRequestContext) []chainid.Team {
	filteredTeams := teams

	if context.IsTeamLeader {
		filteredTeams = make([]chainid.Team, 0)
		for _, membership := range context.UserMemberships {
			for _, team := range teams {
				if team.ID == membership.TeamID && membership.Role == chainid.TeamLeader {
					filteredTeams = append(filteredTeams, team)
					break
				}
			}
		}
	}

	return filteredTeams
}

// FilterUsers filters users based on user role.
// Non-administrator users only have access to non-administrator users.
func FilterUsers(users []chainid.User, context *RestrictedRequestContext) []chainid.User {
	filteredUsers := users

	if !context.IsAdmin {
		filteredUsers = make([]chainid.User, 0)

		for _, user := range users {
			if user.Role != chainid.AdministratorRole {
				filteredUsers = append(filteredUsers, user)
			}
		}
	}

	return filteredUsers
}

// FilterRegistries filters registries based on user role and team memberships.
// Non administrator users only have access to authorized registries.
func FilterRegistries(registries []chainid.Registry, context *RestrictedRequestContext) ([]chainid.Registry, error) {

	filteredRegistries := registries
	if !context.IsAdmin {
		filteredRegistries = make([]chainid.Registry, 0)

		for _, registry := range registries {
			if AuthorizedRegistryAccess(&registry, context.UserID, context.UserMemberships) {
				filteredRegistries = append(filteredRegistries, registry)
			}
		}
	}

	return filteredRegistries, nil
}

// FilterEndpoints filters endpoints based on user role and team memberships.
// Non administrator users only have access to authorized endpoints (can be inherited via endoint groups).
func FilterEndpoints(endpoints []chainid.Endpoint, groups []chainid.EndpointGroup, context *RestrictedRequestContext) ([]chainid.Endpoint, error) {
	filteredEndpoints := endpoints

	if !context.IsAdmin {
		filteredEndpoints = make([]chainid.Endpoint, 0)

		for _, endpoint := range endpoints {
			endpointGroup := getAssociatedGroup(&endpoint, groups)

			if AuthorizedEndpointAccess(&endpoint, endpointGroup, context.UserID, context.UserMemberships) {
				filteredEndpoints = append(filteredEndpoints, endpoint)
			}
		}
	}

	return filteredEndpoints, nil
}

// FilterEndpointGroups filters endpoint groups based on user role and team memberships.
// Non administrator users only have access to authorized endpoint groups.
func FilterEndpointGroups(endpointGroups []chainid.EndpointGroup, context *RestrictedRequestContext) ([]chainid.EndpointGroup, error) {
	filteredEndpointGroups := endpointGroups

	if !context.IsAdmin {
		filteredEndpointGroups = make([]chainid.EndpointGroup, 0)

		for _, group := range endpointGroups {
			if AuthorizedEndpointGroupAccess(&group, context.UserID, context.UserMemberships) {
				filteredEndpointGroups = append(filteredEndpointGroups, group)
			}
		}
	}

	return filteredEndpointGroups, nil
}

func getAssociatedGroup(endpoint *chainid.Endpoint, groups []chainid.EndpointGroup) *chainid.EndpointGroup {
	for _, group := range groups {
		if group.ID == endpoint.GroupID {
			return &group
		}
	}
	return nil
}
