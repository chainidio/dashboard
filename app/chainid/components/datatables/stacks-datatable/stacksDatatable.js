angular.module('chainid.app').component('stacksDatatable', {
  templateUrl: 'app/chainid/components/datatables/stacks-datatable/stacksDatatable.html',
  controller: 'StacksDatatableController',
  bindings: {
    title: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    showTextFilter: '<',
    showOwnershipColumn: '<',
    removeAction: '<',
    displayExternalStacks: '<'
  }
});
