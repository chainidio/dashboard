angular.module('chainid.app').component('slider', {
  templateUrl: 'app/chainid/components/slider/slider.html',
  controller: 'SliderController',
  bindings: {
    model: '=',
    onChange: '&',
    floor: '<',
    ceil: '<',
    step: '<',
    precision: '<'
  }
});
