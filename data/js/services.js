var lovebeatServices = angular.module('lovebeatServices', ['ngResource']);

lovebeatServices.factory('Service', ['$resource',
  function($resource){
      return $resource('/api/services/:serviceId?view=:viewId', {serviceId: '@name'}, {
	get: {method:'GET'},
	query: {method:'GET', isArray:true},
	trigger: {method:'POST'}
    });
  }]);

lovebeatServices.factory('View', ['$resource',
  function($resource){
    return $resource('/api/views/?', {}, {
      query: {method:'GET', isArray:true}
    });
  }]);
