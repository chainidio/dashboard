angular.module('chainid.app').component('teamsDatatable', {
  templateUrl: 'app/chainid/components/datatables/teams-datatable/teamsDatatable.html',
  controller: 'GenericDatatableController',
  bindings: {
    title: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    showTextFilter: '<',
    removeAction: '<'
  }
});
