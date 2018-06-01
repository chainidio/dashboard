angular.module('chainid.app').component('registriesDatatable', {
  templateUrl: 'app/chainid/components/datatables/registries-datatable/registriesDatatable.html',
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
