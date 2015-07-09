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

google.load("visualization", "1.0", {"packages":["corechart", "timeline"]});

module.config(function($locationProvider) {
  $locationProvider.html5Mode(true);
});

module.factory("$google", function($rootScope) {
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

// Autocompletion setup (depends on $http to perform request to /token).
module.run(function($http) {
  if (!document.getElementById("query-input")) {
    return;
  }
  var autocom = new Autocom(document.getElementById("query-input"));
  var keywords = ["describe", "select", "from", "to", "resolution", "where", "all", "metrics", "sample", "by", "now"];
  autocom.options = keywords;
  autocom.prefixPattern = "`[a-zA-Z_][a-zA-Z._-]*`?|[a-zA-Z_][a-zA-Z._-]*";
  autocom.continuePattern = "[a-zA-Z_`.-]";
  autocom.tooltipX = 0;
  autocom.tooltipY = 20;
  autocom.config.skipWord = 0.05; // make it (5x) cheaper to skip letters in a candidate word
  autocom.config.skipWordEnd = 0.01; // add a small cost to skipping ends of words, which benefits shorter candidates
  $http.get("/token").success(function(data, status, headers, config) {
    if (!data.success || !data.body) {
      return;
    }
    if (data.body.functions) {
      autocom.options = autocom.options.concat( data.body.functions );
    }
    if (data.body.metrics) {
      autocom.options = autocom.options.concat( data.body.metrics.map(function(name) {
        if (name.indexOf("-") >= 0) {
          return "`" + name + "`";
        }
        return name;
      }));
    }
  });
});

// A counter is an object with .inc() and .dec() methods,
// as well as .pos() and .zero() predicates.
function Counter() {
  var count = 0;
  this.inc = function() {
    count++;
  };
  this.dec = function() {
    count--;
  };
  this.pos = function() {
    return count > 0;
  };
  this.zero = function() {
    return count === 0;
  };
};
// A singleton Counter for launched queries.
module.factory("$launchedQueries", function() {
  return new Counter();
});
// A singleton Counter for waiting charts
module.factory("$chartWaiting", function() {
  return new Counter();
});

// A ticket booth will give you a ticket with .next()
// The ticket will be .valid() until another ticket has been asked for.
function TicketBooth() {
  var count = 0;
  function Ticket(n) {
    this.valid = function() {
      return n === count;
    }
  }
  this.next = function() {
    return new Ticket(++count);
  };
}
// A singleton ticketbooth for query counting
module.factory("$queryTicketBooth", function() {
  return new TicketBooth();
});

// chart-related functions
function clearChart(chart) {
  if (chart.chart !== null) {
    chart.chart.clearChart();
    chart.chart = null;
  }
};

module.factory("$asyncRender", function($chartWaiting, $rootScope) {
  return function(chart, dataTable, options) {
    google.visualization.events.addListener(chart.chart, "ready", function() {
      $rootScope.$apply(function() { $chartWaiting.dec(); });
    });
    setTimeout(function(){chart.chart.draw(dataTable, options)}, 1);
  };
});

module.factory("$launchRequest", function($google, $http, $receive, $queryTicketBooth, $launchedQueries, $q) {
  return function(params) {
    var resultPromise = $q.defer(); // Will be resolved with received value.
    var ticket = $queryTicketBooth.next();
    $launchedQueries.inc();
    var request = $http.get("/query", {
      params:params
    }).success(function(data, status, headers, config) {
      $launchedQueries.dec();
      if (ticket.valid()) {
        resultPromise.resolve(data);
        $google(function() { $receive(data) });
      }
    }).error(function(data, status, headers, config) {
      $launchedQueries.dec();
      if (ticket.valid()) {
        resultPromise.resolve(data);
        $google(function() { $receive(data) });
      }
    });
    return resultPromise.promise;
  };
});

module.factory("$receiveSelect", function(
      $asyncRender,
      $chartWaiting,
      $inputModel,
      $location,
      $mainChart
  ) {
  return function(object) {
    if (!(object && object.name == "select" && object.body && object.body.length && object.body[0].series && object.body[0].series.length && object.body[0].timerange)) {
      // invalid data.
      return;
    }
    $chartWaiting.inc();
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
        var cell = series[i].values[t];
        if (cell === null) {
          row.push(NaN);
        } else {
          row.push(cell);
        }
      }
      table.push(row);
    }
    var dataTable = google.visualization.arrayToDataTable(table);
    var options = {
      legend:    {position: "bottom"},
      title:     $location.search()["title"],
      chartArea: {left: "5%", width:"90%", top: "5%", height: "85%"}
    }
    if ($inputModel.renderType === "line") {
      $mainChart.chart = new google.visualization.LineChart($mainChart.dom);
    } else if ($inputModel.renderType === "area") {
      options.isStacked = true;
      $mainChart.chart = new google.visualization.AreaChart($mainChart.dom);
    }
    $asyncRender($mainChart, dataTable, options);
  };
});

module.factory("$profilingEnabled", function($inputModel) {
  return {
    hasProfiling: function(data) {
      return $inputModel.profile && data && data.profile;
    }
  };
});

module.factory("$receiveProfile", function($profilingEnabled, $chartWaiting, $asyncRender, $timelineChart) {
  return function(object) {
    if (!$profilingEnabled.hasProfiling(object)) {
      return
    }
    $chartWaiting.inc();
    var dataTable = new google.visualization.DataTable();
    var options = {
      chartArea: {left: "5%", width:"90%", height: "500px"},
      avoidOverlappingGridLines: false
    };
    $timelineChart.chart = new google.visualization.Timeline($timelineChart.dom);
    dataTable.addColumn({ type: 'string', id: 'Name' });
    dataTable.addColumn({ type: 'number', id: 'Start' });
    dataTable.addColumn({ type: 'number', id: 'End' });
    var minValue = Number.POSITIVE_INFINITY;
    for (var i = 0; i < object.profile.length; i++) {
      var profile = object.profile[i];
      minValue = Math.min(profile.start, minValue);
      minValue = Math.min(profile.finish, minValue);
    };
    function normalize(value) {
      return value - minValue;
    };
    for (var i = 0; i < object.profile.length; i++) {
      var profile = object.profile[i];
      var row = [ profile.name , normalize(profile.start), normalize(profile.finish) ];
      dataTable.addRows([row]);
    }
    $asyncRender($timelineChart, dataTable, options);
  };
});

module.factory("$mainChart", function() {
  return {
    dom:   document.getElementById("chart-div"),
    chart: null
  };
});

module.factory("$timelineChart", function() {
  return {
    dom:   document.getElementById("timeline-div"),
    chart: null
  };
});

module.factory("$receive", function($mainChart, $timelineChart, $receiveSelect, $receiveProfile) {
  return function(object) {
    clearChart($mainChart);
    clearChart($timelineChart);
    $receiveSelect(object);
    $receiveProfile(object);
  };
});

module.factory("$inputModel", function() {
  return {
    profile: false,
    query: "",
    renderType: "line"
  };
});

module.controller("mainCtrl", function(
  $location,
  $scope,
  $launchedQueries,
  $chartWaiting,
  $launchRequest,
  $inputModel,
  $profilingEnabled
  ) {

  $scope.queryHistory = [];

  // Store the $inputModel so that the view can change it through inputs.
  $scope.inputModel = $inputModel;

  $scope.embedLink = "";

  $scope.queryResult = "";

  function updateEmbed() {
    var url = $location.absUrl();
    var queryAt = url.indexOf("?");
    $scope.embedLink = $location.protocol() + "://" + $location.host() + ":" + $location.port()
      + "/embed" + url.substring(queryAt);
  }

  // Triggers when the button is clicked.
  $scope.onSubmitQuery = function() {
    $inputModel.query = document.getElementById("query-input").value;
    $location.search("query", $inputModel.query);
    $location.search("renderType", $inputModel.renderType);
    $location.search("profile", $inputModel.profile.toString());
  };

  $scope.$on("$locationChangeSuccess", function() {
    // this triggers at least once (in the beginning).
    var queries = $location.search();
    $inputModel.query = queries["query"] || "";
    $inputModel.renderType = queries["renderType"] || "line";
    $inputModel.profile = queries["profile"] === "true";
    // Add the query to the history, if it hasn't been seen before and it's non-empty
    var trimmedQuery = $inputModel.query.trim();
    if (trimmedQuery !== "" && $scope.queryHistory.indexOf(trimmedQuery) === -1) {
      $scope.queryHistory.push(trimmedQuery);
    }
    if (trimmedQuery) {
      $launchRequest({
        profile: $inputModel.profile,
        query:   $inputModel.query
      }).then(function(data) {
        $scope.queryResult = data;
      });
    }
  });

  $scope.historySelect = function(query) {
    $inputModel.query = query;
  }

  // true if the output should be tabular.
  $scope.isTabular = function() {
    return ["describe all", "describe metrics", "describe"].indexOf($scope.queryResult.name) >= 0;
  };

  $scope.hasProfiling = $profilingEnabled.hasProfiling; // So that the HTML view can check for profiling

  $scope.screenState = function() {
    if ($launchedQueries.pos()) {
      return "loading";
    } else if ($launchedQueries.zero() && $chartWaiting.pos()) {
      return "rendering";
    } else if ($scope.queryResult && !$scope.queryResult.success) {
      return "error";
    } else {
      return "rendered";
    }
  };
  updateEmbed()
});

module.controller("embedCtrl", function($location, $scope, $launchedQueries, $chartWaiting, $launchRequest, $inputModel, $profilingEnabled){
  $scope.queryResult = "";
  $scope.screenState = function() {
    if ($launchedQueries.pos()) {
      return "loading";
    } else if ($launchedQueries.zero() && $chartWaiting.pos()) {
      return "rendering";
    } else if ($scope.queryResult && !$scope.queryResult.success) {
      return "error";
    } else {
      return "rendered";
    }
  };
  $scope.hasProfiling = $profilingEnabled.hasProfiling;

  var queries = $location.search();
  $inputModel.profile = false;
  $inputModel.query = "";
  $inputModel.renderType = queries["renderType"] || "line";
  $launchRequest({
    profile: false,
    query:   queries["query"] || ""
  }).then(function(data) {
    $scope.queryResult = data;
  });

  var url = $location.absUrl();
  var at = url.indexOf("?");
  $scope.metricsURL = $location.protocol() + "://" + $location.host() + ":" + $location.port()
    + "/ui" + url.substring(at);
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
