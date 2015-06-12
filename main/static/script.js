"use strict";

var module = angular.module("main",[]);

var chartsReady = false;

var debounce = false;

module.controller("mainCtrl", function($scope, $http) {
	$scope.queryResult = "";
	$scope.queryText = "";
	$scope.onSubmitQuery = function() {
		debounce = true;
		window.location.hash = "#" + encodeURIComponent($scope.queryText);
		$http.get('/query', {params:{query: $scope.queryText}}).
		  success(function(data, status, headers, config) {
		    $scope.queryResult = data;
		  }).
		  error(function(data, status, headers, config) {
		    $scope.queryResult = data;
		  });
	};

	$scope.$watch("queryResult", resultUpdate);

	function readHash() {
		if (debounce) {
			debounce = false;
			return;
		}
		var urlQuery = window.location.hash
		if (urlQuery != "") {
			// Drop the leading '#'
			urlQuery = urlQuery.substring(1);
			// The remainder is the query to perform.
			// Decode it (if neccesarry).
			$scope.queryText = decodeURIComponent(urlQuery);
			$scope.onSubmitQuery();
		}
	}

	readHash(); // Read the current hash (if present).

	// Whenever the user changes the hash (such as pasting in a new URL), call readHash() again.
	window.addEventListener("hashchange", readHash, false);
});

var chart;


function onload() {
	google.load('visualization', '1.0', {'packages':['corechart']});
	google.setOnLoadCallback(function(){ chartsReady = true; });
}

function dateFromIndex(index, timerange) {
	return new Date(timerange.start + timerange.resolution * index);
}



function resultUpdate(object) {
	if (!chartsReady) {
		// Ask again in 30ms.
		setTimeout(function(){resultUpdate(object);}, 10);
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

	chart = new google.visualization.LineChart(document.getElementById('chart-div'));
	setTimeout(function(){chart.draw(dataTable, options)}, 1);
}
