angular.module('chainid.app').component('endpointsDatatable', {
  templateUrl: 'app/chainid/components/datatables/endpoints-datatable/endpointsDatatable.html',
  controller: 'GenericDatatableController',
  bindings: {
    title: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    showTextFilter: '<',
    endpointManagement: '<',
    accessManagement: '<',
    removeAction: '<'
  }
});
