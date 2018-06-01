angular.module('chainid.app').component('codeEditor', {
  templateUrl: 'app/chainid/components/code-editor/codeEditor.html',
  controller: 'CodeEditorController',
  bindings: {
    identifier: '@',
    placeholder: '@',
    yml: '<',
    readOnly: '<',
    onChange: '<',
    value: '<'
  }
});
