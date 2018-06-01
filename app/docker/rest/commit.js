angular.module('chainid.docker')
.factory('Commit', ['$resource', 'API_ENDPOINT_ENDPOINTS', 'EndpointProvider', function CommitFactory($resource, API_ENDPOINT_ENDPOINTS, EndpointProvider) {
  'use strict';
  return $resource(API_ENDPOINT_ENDPOINTS + '/:endpointId/docker/commit', {
    endpointId: EndpointProvider.endpointID
  },
  {
    commitContainer: {method: 'POST', params: {container: '@id', repo: '@repo', tag: '@tag'}, ignoreLoadingBar: true}
  });
}]);