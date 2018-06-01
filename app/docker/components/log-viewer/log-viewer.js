angular.module('chainid.docker').component('logViewer', {
  templateUrl: 'app/docker/components/log-viewer/logViewer.html',
  controller: 'LogViewerController',
  bindings: {
    data: '=',
    displayTimestamps: '=',
    logCollectionChange: '<',
    lineCount: '='
  }
});
