// Copyright 2015 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
"use strict";

var module = angular.module("main",[]);

google.load("visualization", "1.0", {"packages":["corechart"]});

module.config(function($locationProvider) {
  $locationProvider.html5Mode(true);
});

module.service("$google", function($rootScope) {
  // abstraction over the async loading of google libraries.
  // registered functions are either invoked immediately (if the library finished loading).
  // or queued in an array.
  var googleFunctions = [];
  var googleLoaded = false;
  google.setOnLoadCallback(function(){
    $rootScope.$apply(function() {
      googleLoaded = true;
      googleFunctions.map(function(googleFunction) {
        googleFunction();
      });
      googleFunctions.length = 0; // clear the array.
    });
  });

  return function(func) {
    if (googleLoaded) {
      func();
    } else {
      googleFunctions.push(func);
    }
  }
});

module.controller("mainCtrl", function(
  $google,
  $http,
  $location,
  $q,
  $scope
  ) {
  var queryCounter = 0; // ever-incrementing counter of queries - used to detect out-of-order queries.
  $scope.queryResult = "";
  $scope.inputModel = {
    query: "",
    renderType: "line"
  };

  // Triggers when the button is clicked.
  $scope.onSubmitQuery = function() {
    $location.search("query", $scope.inputModel.query)
    $location.search("renderType", $scope.inputModel.renderType)
  };

  // true if the output should be tabular.
  $scope.isTabular = function() {
    return ["describe all", "describe metrics", "describe"].indexOf($scope.queryResult.name) >= 0;
  };
  $scope.screenState = function() {
    if ($scope.launchedQuery > 0) {
      return "loading";
    } else if ($scope.launchedQuery === 0 && $scope.chartWaiting > 0) {
      return "rendering";
    } else if ($scope.queryResult && !$scope.queryResult.success) {
      return "error";
    } else {
      return "rendered";
    }
  };

  $scope.chartWaiting = 0;
  $scope.launchedQuery = 0;

  function launchRequest(query) {
    queryCounter++;
    var currentQueryCounter = queryCounter; // store it in the closure.
    $scope.launchedQuery++;
    var request = $http.get("/query", {
      params:{query: query}
    }).success(function(data, status, headers, config) {
      $scope.launchedQuery--;
      $scope.queryResult = data;
      if (currentQueryCounter === queryCounter) {
        $google(function() { receive(data) });
      }
    }).error(function(data, status, headers, config) {
      $scope.launchedQuery--;
      $scope.queryResult = data;
      if (currentQueryCounter === queryCounter) {
        $google(function() { receive(data) });
      }
    });
  }

  var chart = null;
  function receive(object) {
    if (chart !== null) {
      chart.clearChart();
      chart = null;
    }
    if (!(object && object.name == "select" && object.body && object.body.length && object.body[0].series && object.body[0].series.length && object.body[0].timerange)) {
      // invalid data.
      return;
    }
    $scope.chartWaiting++;
    var series = [];
    var labels = ["Time"];
    var table = [labels];
    for (var i = 0; i < object.body.length; i++) {
      // Each of these is a list of series
      var serieslist = object.body[i];
      for (var j = 0; j < serieslist.series.length; j++) {
        var s = object.body[i].series[j];
        series.push(s);
        labels.push(makeLabel(serieslist, s));
      }
    }
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
      legend: {position: "bottom"},
      chartArea: {left: "5%", width:"90%", height: "300px"}
    }
    var chartDiv = document.getElementById("chart-div")
    if ($scope.inputModel.renderType === "line") {
      chart = new google.visualization.LineChart(chartDiv);
    } else if ($scope.inputModel.renderType === "area") {
      options.isStacked = true;
      chart = new google.visualization.AreaChart(chartDiv);
    }
    google.visualization.events.addListener(chart, "ready", function() {
      $scope.$apply(function($scope) { $scope.chartWaiting--; });
    });
    setTimeout(function(){chart.draw(dataTable, options)}, 1);
  }

  $scope.$on("$locationChangeSuccess", function() {
    // this triggers at least once (in the beginning).
    var queries = $location.search();
    $scope.inputModel.query = queries["query"] || "";
    $scope.inputModel.renderType = queries["renderType"] || "line";
    if ($scope.inputModel.query) {
      launchRequest($scope.inputModel.query);
    }
  });
});

function dateFromIndex(index, timerange) {
  return new Date(timerange.start + timerange.resolution * index);
}

function makeLabel(serieslist, series) {
  var label = serieslist.name + " ";
  for (var key in series.tagset) {
    label += key + ":" + series.tagset[key] + " "; 
  }
  return label;
}
