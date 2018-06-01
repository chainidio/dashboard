angular.module('chainid.app').component('endpointSelector', {
  templateUrl: 'app/chainid/components/endpoint-selector/endpointSelector.html',
  controller: 'EndpointSelectorController',
  bindings: {
    'endpoints': '<',
    'groups': '<',
    'selectEndpoint': '<'
  }
});
