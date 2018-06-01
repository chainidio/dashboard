angular.module('chainid.app').component('groupsDatatable', {
  templateUrl: 'app/chainid/components/datatables/groups-datatable/groupsDatatable.html',
  controller: 'GenericDatatableController',
  bindings: {
    title: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    showTextFilter: '<',
    accessManagement: '<',
    removeAction: '<'
  }
});
