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
  var keywords = ["describe", "select", "from", "to", "resolution", "where", "all", "metrics", "sample", "by"];
  autocom.options = keywords;
  autocom.prefixPattern = "`[a-zA-Z][a-zA-Z.-]*`?|[a-zA-Z][a-zA-Z.-]*";
  autocom.tooltipX = 0;
  autocom.tooltipY = 20;
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
})

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

module.factory("$asyncRender", function($chartWaiting) {
  return function($scope, chart, dataTable, options) {
    google.visualization.events.addListener(chart.chart, "ready", function() {
      $scope.$apply(function() { $chartWaiting.dec(); });
    });
    setTimeout(function(){chart.chart.draw(dataTable, options)}, 1);
  };
});

module.factory("$launchRequest", function($google, $http, $receive, $queryTicketBooth, $launchedQueries) {
  return function($scope, params) {
    var ticket = $queryTicketBooth.next();
    $launchedQueries.inc();
    var request = $http.get("/query", {
      params:params
    }).success(function(data, status, headers, config) {
      $launchedQueries.dec();
      $scope.queryResult = data;
      if (ticket.valid()) {
        $google(function() { $receive($scope, data) });
      }
    }).error(function(data, status, headers, config) {
      $launchedQueries.dec();
      $scope.queryResult = data;
      if (ticket.valid()) {
        $google(function() { $receive($scope, data) });
      }
    });
  };
});

module.factory("$receiveSelect", function($chartWaiting, $asyncRender) {
  return function($scope, object, chart) {
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
        row.push(series[i].values[t] || NaN);
      }
      table.push(row);
    }
    var dataTable = google.visualization.arrayToDataTable(table);
    var options = {
      legend: {position: "bottom"},
      chartArea: {left: "5%", width:"90%", top: "5%", height: "90%"}
    }
    if ($scope.inputModel.renderType === "line") {
      chart.chart = new google.visualization.LineChart(chart.dom);
    } else if ($scope.inputModel.renderType === "area") {
      options.isStacked = true;
      chart.chart = new google.visualization.AreaChart(chart.dom);
    }
    $asyncRender($scope, chart, dataTable, options);
  };
});

module.factory("$receiveProfile", function($chartWaiting, $asyncRender) {
  return function($scope, object, chart) {
    if (!$scope.hasProfiling()) {
      return
    }
    console.log('increase');
    $chartWaiting.inc();
    var dataTable = new google.visualization.DataTable();
    var options = {
      chartArea: {left: "5%", width:"90%", height: "500px"},
      avoidOverlappingGridLines: false
    };
    chart.chart = new google.visualization.Timeline(chart.dom);
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
    $asyncRender($scope, chart, dataTable, options);
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
  return function($scope, object) {
    clearChart($mainChart);
    clearChart($timelineChart);
    $receiveSelect($scope, object, $mainChart);
    $receiveProfile($scope, object, $timelineChart);
  };
});

module.controller("mainCtrl", function(
  $location,
  $scope,
  $launchedQueries,
  $chartWaiting,
  $launchRequest
  ) {

  $scope.embedLink = "";

  $scope.queryResult = "";
  $scope.inputModel = {
    profile: false,
    query: "",
    renderType: "line"
  };

  // Triggers when the button is clicked.
  $scope.onSubmitQuery = function() {
    // TODO - unhack this.
    $scope.inputModel.query = document.getElementById("query-input").value;
    $location.search("query", $scope.inputModel.query);
    $location.search("renderType", $scope.inputModel.renderType);
    $location.search("profile", $scope.inputModel.profile.toString());

    var url = $location.absUrl();
    var queryAt = url.indexOf("?");
    $scope.embedLink = url.substring(0, queryAt) + "embed.html" + url.substring(queryAt);
  };

  $scope.$on("$locationChangeSuccess", function() {
    // this triggers at least once (in the beginning).
    var queries = $location.search();
    $scope.inputModel.query = queries["query"] || "";
    $scope.inputModel.renderType = queries["renderType"] || "line";
    $scope.inputModel.profile = queries["profile"] === "true";
    if ($scope.inputModel.query) {
      $launchRequest($scope, {
        profile: $scope.inputModel.profile,
        query:   $scope.inputModel.query
      });
    }
  });

  // true if the output should be tabular.
  $scope.isTabular = function() {
    return ["describe all", "describe metrics", "describe"].indexOf($scope.queryResult.name) >= 0;
  };
  $scope.hasProfiling = function() {
    return !!($scope.queryResult && $scope.queryResult.profile);
  };

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
  var url = $location.absUrl();
  var queryAt = url.indexOf("?");
  $scope.embedLink = url.substring(0, queryAt) + "embed.html" + url.substring(queryAt);
  console.log($scope.embedLink);
});

module.controller("embedCtrl", function($location, $scope, $launchedQueries, $chartWaiting, $launchRequest){
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
  $scope.hasProfiling = function() {
    return false;
  };

  var queries = $location.search();
  $scope.inputModel = { renderType: "line" };
  $scope.inputModel.renderType = queries["renderType"] || "line";
  $launchRequest($scope, {
    profile: false,
    query:   queries["query"] || ""
  });

  var url = $location.absUrl();
  var embedString = "embed.html";
  var at = url.indexOf(embedString);
  $scope.metricsURL = url.substring(0, at) + url.substring(at + embedString.length);
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
