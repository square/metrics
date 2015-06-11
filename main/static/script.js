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
	// Perform this callback once the package has loaded:
	google.setOnLoadCallback(drawChart);
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

	var startTime = object.body[0].timerange.start;
	var endTime = object.body[0].timerange.end;
	var resolution = object.body[0].timerange.resolution;

	for (var t = 0; t < series[0].values.length; t++) {
		var row = [new Date(startTime + resolution * t)];
		for (var i = 0; i < series.length; i++) {
			row.push(series[i].values[t] || 0);
		}
		table.push(row);
	}
	console.log(table);
	var trueTable = google.visualization.arrayToDataTable(table);
	var options = {
		title: "Select Result",
		legend: {position: "bottom"}
	}

	var chart = new google.visualization.LineChart(document.getElementById('chart_div'));
	chart.draw(trueTable, options);
}

/*
function resultUpdate(value) {
	if (chart) {
		chart.destroy();
		chart = null;
	}
	if (value && value.name == "select" && value.body && value.body.length && value.body[0].series && value.body[0].series.length) {
		var datasets = [];
		for (var i = 0; i < value.body.length; i++) {
			var list = value.body[i];
			for (var j = 0; j < list.series.length; j++) {
				var series = list.series[j];
				var seriesData = {
					data:[],
					label: "D" + datasets.length,
		            strokeColor: "rgba(151,187,205,1)",
				};
				for (var t = 0; t < series.values.length; t++) {
					seriesData.data[t] = series.values[t];
				}
				datasets.push(seriesData);
			}
		}
		for (var i = 0; i < datasets.length; i++) {
			datasets[i].pointColor = datasets[i].strokeColor = "hsl(" + Math.floor(360 * i / datasets.length) + ",50%,65%)";
		}

		var labels = [];
		for (var i = 0; i < value.body[0].series[0].values.length; i++) {
			labels[i] = i + "X";
		}

		chart = new Chart(ctx).Line({datasets:datasets,labels:labels}, {datasetFill:false, bezierCurve: false, pointDot:false});
	}
}

*/