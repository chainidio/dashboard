angular.module('chainid.app').component('stackServicesDatatable', {
  templateUrl: 'app/chainid/components/datatables/stack-services-datatable/stackServicesDatatable.html',
  controller: 'GenericDatatableController',
  bindings: {
    title: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    nodes: '<',
    scaleAction: '<',
    publicUrl: '<',
    showTextFilter: '<'
  }
});
