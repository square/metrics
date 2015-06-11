"use strict";

console.log("Hello, world!");

var module = angular.module("main",[]);

module.controller("mainCtrl", function($scope, $http) {
	$scope.queryResult = "";
	$scope.queryText = "";
	$scope.onSubmitQuery = function() {
		$http.get('/query', {params:{query: $scope.queryText}}).
		  success(function(data, status, headers, config) {
		    $scope.queryResult = data;
		  }).
		  error(function(data, status, headers, config) {
		    $scope.queryResult = data;
		  });
	};
	$scope.$watch("queryResult", resultUpdate);
});

var chart;
var canvas;
var ctx;

function onload() {
	canvas = document.getElementById("graphCanvas");
	ctx = canvas.getContext("2d");
}

function resultUpdate(value) {
	if (chart) {
		chart.destroy();
		chart = null;
	}
	if (value && value.name == "select") {
		var datasets = [];
		for (var i = 0; i < value.body.length; i++) {
			var list = value.body[i];
			for (var j = 0; j < list.Series.length; j++) {
				var series = list.Series[j];
				var seriesData = {
					data:[],
					label: "D" + datasets.length,
		            strokeColor: "rgba(151,187,205,1)",
				};
				for (var t = 0; t < series.Values.length; t++) {
					seriesData.data[t] = series.Values[t];
				}
				datasets.push(seriesData);
			}
		}
		for (var i = 0; i < datasets.length; i++) {
			datasets[i].pointColor = datasets[i].strokeColor = "hsl(" + Math.floor(360 * i / datasets.length) + ",50%,65%)";
		}
		console.log(ctx);
		var labels = [];

		for (var i = 0; i < value.body[0].Series[0].Values.length; i++) {
			labels[i] = i + "X";
		}

		chart = new Chart(ctx).Line({datasets:datasets,labels:labels}, {datasetFill:false, bezierCurve: false, pointDot:false});
	}
}