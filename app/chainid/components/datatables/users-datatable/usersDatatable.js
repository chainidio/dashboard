angular.module('chainid.app').component('usersDatatable', {
  templateUrl: 'app/chainid/components/datatables/users-datatable/usersDatatable.html',
  controller: 'GenericDatatableController',
  bindings: {
    title: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    showTextFilter: '<',
    removeAction: '<',
    authenticationMethod: '<'
  }
});
