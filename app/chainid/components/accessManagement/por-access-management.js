angular.module('chainid.app').component('porAccessManagement', {
  templateUrl: 'app/chainid/components/accessManagement/porAccessManagement.html',
  controller: 'porAccessManagementController',
  bindings: {
    accessControlledEntity: '<',
    inheritFrom: '<',
    entityType: '@',
    updateAccess: '&'
  }
});
