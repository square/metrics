"use strict";

var module = angular.module("main",[]);

var chartsReady = false;

var debounce = false;

module.controller("mainCtrl", function($scope, $http, $q) {
	$scope.queryResult = "";
	$scope.queryText = "";
	$scope.launchedQuery = null;
	// Triggers when the button is clicked.
	$scope.onSubmitQuery = function() {
		debounce = true;
		window.location.hash = "#" + encodeURIComponent($scope.queryText);
		launchRequest($scope.queryText);
	};

	function readHash() {
		if (debounce) {
			debounce = false;
			return;
		}
		var urlQuery = window.location.hash
		if (urlQuery != "" && urlQuery != "#") {
			// Drop the leading '#'
			urlQuery = urlQuery.substring(1);
			// The remainder is the query to perform.
			// Decode it (if neccesarry).
			$scope.queryText = decodeURIComponent(urlQuery);
			launchRequest($scope.queryText);
		}
	}

	function cancelQuery() {
		if ($scope.launchedQuery === null) {
			return;
		}
		$scope.launchedQuery.canceler.resolve();
		$scope.launchedQuery = null;
	}

	function launchRequest(query) {
		cancelQuery();
		var canceler = $q.defer();
		var request = $http.get('/query', {
			params:{query: query},
			timeout:canceler // This allows the query to be canceled.
		}).success(function(data, status, headers, config) {
			$scope.queryResult = data;
			receive(data);
		}).error(function(data, status, headers, config) {
			$scope.queryResult = data;
			receive(data);
		});
		// Save the launched query so that it can be canceled later.
		$scope.launchedQuery = {request: request, canceler:canceler};

	}

	$scope.chartWaiting = false;

	function receive(object) {
		$scope.launchedQuery = null;
		if (!chartsReady) {
			setTimeout(function(){receive(object);}, 10);
			return;
		}
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
				row.push(series[i].values[t] || NaN);
			}
			table.push(row);
		}
		var dataTable = google.visualization.arrayToDataTable(table);
		var options = {
			title: "Select Result",
			legend: {position: "bottom"},
			chartArea: {left: "5%", width:"90%"}
		}
		$scope.chartWaiting = true;
		$scope.chart = new google.visualization.LineChart(document.getElementById('chart-div'));
		google.visualization.events.addListener($scope.chart, 'ready', function() {
			$scope.chartWaiting = false;
			$scope.$apply(); // Updating to tell Angular that chartWaiting has changed.
		});
		setTimeout(function(){$scope.chart.draw(dataTable, options)}, 1);
	}

	readHash(); // Read the current hash (if present).

	// Whenever the user changes the hash (such as pasting in a new URL), call readHash() again.
	window.addEventListener("hashchange", readHash, false);
});

function onload() {
	google.load('visualization', '1.0', {'packages':['corechart']});
	google.setOnLoadCallback(function(){ chartsReady = true; });
}

function dateFromIndex(index, timerange) {
	return new Date(timerange.start + timerange.resolution * index);
}

onload();

