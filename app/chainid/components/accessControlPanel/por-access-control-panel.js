angular.module('chainid.app').component('porAccessControlPanel', {
  templateUrl: 'app/chainid/components/accessControlPanel/porAccessControlPanel.html',
  controller: 'porAccessControlPanelController',
  bindings: {
    // The component will use this identifier when updating the resource control object.
    resourceId: '<',
    // The component will display information about this resource control object.
    resourceControl: '=',
    // This component is usually displayed inside a resource-details view.
    // This variable specifies the type of the associated resource.
    // Accepted values: 'container', 'service' or 'volume'.
    resourceType: '<'
  }
});
