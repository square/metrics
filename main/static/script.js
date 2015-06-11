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

function onload() {
	google.load('visualization', '1.0', {'packages':['corechart']});
}

function dateFromIndex(index, timerange) {
	return new Date(timerange.start + timerange.resolution * index);
}

function resultUpdate(object) {
	if (!(object && object.name == "select" && object.body && object.body.length && object.body[0].series && object.body[0].series.length && object.body[0].timerange)) {
		return;
	}
	var series = [];
	for (var i = 0; i < object.body.length; i++) {
		// Each of these is a list of series
		for (var j = 0; j < object.body[i].series.length; j++) {
			series.push(object.body[i].series[j]);
		}
	}
	var labels = ["Time"];
	for (var i = 0; i < series.length; i++) {
		var label = "";
		for (var k in series[i].tagset) {
			label += k + ": " + series[i].tagset[k] + " "; 
		}
		labels.push(label);
	}
	var table = [labels];
	// Next, add each row.

	var timerange = object.body[0].timerange;
	for (var t = 0; t < series[0].values.length; t++) {
		var row = [dateFromIndex(t, timerange)];
		for (var i = 0; i < series.length; i++) {
			row.push(series[i].values[t] || 0);
		}
		table.push(row);
	}
	var dataTable = google.visualization.arrayToDataTable(table);
	var options = {
		title: "Select Result",
		legend: {position: "bottom"},
		chartArea: {left: "5%", width:"90%"}
	}

	var chart = new google.visualization.LineChart(document.getElementById('chart-div'));
	setTimeout(function(){chart.draw(dataTable, options)}, 1);
}
