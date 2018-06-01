angular.module('chainid.app', [])
.config(['$stateRegistryProvider', function ($stateRegistryProvider) {
  'use strict';

  var root = {
    name: 'root',
    abstract: true,
    resolve: {
      requiresLogin: ['StateManager', function (StateManager) {
        var applicationState = StateManager.getState();
        return applicationState.application.authentication;
      }]
    },
    views: {
      'sidebar@': {
        templateUrl: 'app/chainid/views/sidebar/sidebar.html',
        controller: 'SidebarController'
      }
    }
  };

  var chainid = {
    name: 'chainid',
    parent: 'root',
    abstract: true
  };

  var about = {
    name: 'chainid.about',
    url: '/about',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/about/about.html'
      }
    }
  };

  var account = {
    name: 'chainid.account',
    url: '/account',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/account/account.html',
        controller: 'AccountController'
      }
    }
  };

  var authentication = {
    name: 'chainid.auth',
    url: '/auth',
    params: {
      logout: false,
      error: ''
    },
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/auth/auth.html',
        controller: 'AuthenticationController'
      },
      'sidebar@': {}
    },
    data: {
      requiresLogin: false
    }
  };

  var init = {
    name: 'chainid.init',
    abstract: true,
    url: '/init',
    data: {
      requiresLogin: false
    },
    views: {
      'sidebar@': {}
    }
  };

  var initEndpoint = {
    name: 'chainid.init.endpoint',
    url: '/endpoint',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/init/endpoint/initEndpoint.html',
        controller: 'InitEndpointController'
      }
    }
  };

  var initAdmin = {
    name: 'chainid.init.admin',
    url: '/admin',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/init/admin/initAdmin.html',
        controller: 'InitAdminController'
      }
    }
  };

  var endpoints = {
    name: 'chainid.endpoints',
    url: '/endpoints',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/endpoints/endpoints.html',
        controller: 'EndpointsController'
      }
    }
  };

  var endpoint = {
    name: 'chainid.endpoints.endpoint',
    url: '/:id',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/endpoints/edit/endpoint.html',
        controller: 'EndpointController'
      }
    }
  };

  var endpointCreation  = {
    name: 'chainid.endpoints.new',
    url: '/new',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/endpoints/create/createendpoint.html',
        controller: 'CreateEndpointController'
      }
    }
  };

  var endpointAccess = {
    name: 'chainid.endpoints.endpoint.access',
    url: '/access',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/endpoints/access/endpointAccess.html',
        controller: 'EndpointAccessController'
      }
    }
  };

  var groups = {
    name: 'chainid.groups',
    url: '/groups',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/groups/groups.html',
        controller: 'GroupsController'
      }
    }
  };

  var group = {
    name: 'chainid.groups.group',
    url: '/:id',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/groups/edit/group.html',
        controller: 'GroupController'
      }
    }
  };

  var groupCreation = {
    name: 'chainid.groups.new',
    url: '/new',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/groups/create/creategroup.html',
        controller: 'CreateGroupController'
      }
    }
  };

  var groupAccess = {
    name: 'chainid.groups.group.access',
    url: '/access',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/groups/access/groupAccess.html',
        controller: 'GroupAccessController'
      }
    }
  };

  var registries = {
    name: 'chainid.registries',
    url: '/registries',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/registries/registries.html',
        controller: 'RegistriesController'
      }
    }
  };

  var registry = {
    name: 'chainid.registries.registry',
    url: '/:id',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/registries/edit/registry.html',
        controller: 'RegistryController'
      }
    }
  };

  var registryCreation  = {
    name: 'chainid.registries.new',
    url: '/new',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/registries/create/createregistry.html',
        controller: 'CreateRegistryController'
      }
    }
  };

  var registryAccess = {
    name: 'chainid.registries.registry.access',
    url: '/access',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/registries/access/registryAccess.html',
        controller: 'RegistryAccessController'
      }
    }
  };

  var settings = {
    name: 'chainid.settings',
    url: '/settings',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/settings/settings.html',
        controller: 'SettingsController'
      }
    }
  };

  var settingsAuthentication = {
    name: 'chainid.settings.authentication',
    url: '/auth',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/settings/authentication/settingsAuthentication.html',
        controller: 'SettingsAuthenticationController'
      }
    }
  };

  var support = {
    name: 'chainid.support',
    url: '/support',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/support/support.html'
      }
    }
  };

  var users = {
    name: 'chainid.users',
    url: '/users',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/users/users.html',
        controller: 'UsersController'
      }
    }
  };

  var user = {
    name: 'chainid.users.user',
    url: '/:id',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/users/edit/user.html',
        controller: 'UserController'
      }
    }
  };

  var teams = {
    name: 'chainid.teams',
    url: '/teams',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/teams/teams.html',
        controller: 'TeamsController'
      }
    }
  };

  var team = {
    name: 'chainid.teams.team',
    url: '/:id',
    views: {
      'content@': {
        templateUrl: 'app/chainid/views/teams/edit/team.html',
        controller: 'TeamController'
      }
    }
  };

  $stateRegistryProvider.register(root);
  $stateRegistryProvider.register(chainid);
  $stateRegistryProvider.register(about);
  $stateRegistryProvider.register(account);
  $stateRegistryProvider.register(authentication);
  $stateRegistryProvider.register(init);
  $stateRegistryProvider.register(initEndpoint);
  $stateRegistryProvider.register(initAdmin);
  $stateRegistryProvider.register(endpoints);
  $stateRegistryProvider.register(endpoint);
  $stateRegistryProvider.register(endpointAccess);
  $stateRegistryProvider.register(endpointCreation);
  $stateRegistryProvider.register(groups);
  $stateRegistryProvider.register(group);
  $stateRegistryProvider.register(groupAccess);
  $stateRegistryProvider.register(groupCreation);
  $stateRegistryProvider.register(registries);
  $stateRegistryProvider.register(registry);
  $stateRegistryProvider.register(registryAccess);
  $stateRegistryProvider.register(registryCreation);
  $stateRegistryProvider.register(settings);
  $stateRegistryProvider.register(settingsAuthentication);
  $stateRegistryProvider.register(support);
  $stateRegistryProvider.register(users);
  $stateRegistryProvider.register(user);
  $stateRegistryProvider.register(teams);
  $stateRegistryProvider.register(team);
}]);
