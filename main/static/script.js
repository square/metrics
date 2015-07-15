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

module.factory("$windowSize", function($window) {
  return {
    height:  $window.innerHeight,
    width:   $window.innerWidth,
    version: 0 // updated whenever width or height is updated, so this object can be watched.
  }
});

module.directive("googleChart", function($chartWaiting, $timeout, $windowSize) {
  return {
    restrict: "E",
    template: "<div style='width:100%;height:100px'> <pre>{{ data }} {{ option }}</pre></div>",
    scope: {
      chartType: "&",
      data:      "&",
      option:    "&"
    },
    link: function(scope, element, attrs) {
      var chart = null;
      scope.$watch("option()", function(newValue) {
        render();
      }, true);
      scope.$watch("data()", function(newValue) {
        render();
      });
      scope.$watch(function() { return $windowSize.version }, function(newValue) {
        render();
      });
      scope.$watch("chartType()", function(newValue) {
        if (chart !== null) {
          chart.clearChart();
          chart = null;
        }
        if (newValue === "line") {
          chart = new google.visualization.LineChart(element[0]);
        } else if (newValue === "area") {
          chart = new google.visualization.AreaChart(element[0]);
        } else if (newValue === "timeline") {
          chart = new google.visualization.Timeline(element[0]);
        }
        render();
      });

      function render() {
        $timeout(function(){
          // TODO - add this somewhere.
          google.visualization.events.addListener(chart, "ready", function() {
            scope.$apply(function() { $chartWaiting.dec(); });
          });
          var data = scope.data();
          var option = scope.option();
          if (data && option) {
            $chartWaiting.inc();
            chart.draw(data, option);
          }
        }, 1);
      }
    }
  };
});

module.run(function($window, $timeout, $windowSize) {
  var DELAY_MS = 100;
  var counter = 0;
  angular.element($window).bind("resize", function() {
    counter++;
    var currentCounter = counter; // capture the current value via the closure.
    $timeout(function() {
      if (currentCounter  == counter) {
        // console.log('new size:', $window.innerWidth, $window.innerHeight);
        $windowSize.height = $window.innerHeight;
        $windowSize.width = $window.innerWidth;
        $windowSize.version++;
      }
    }, DELAY_MS);
  })
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

module.factory("$launchRequest", function($google, $http, $queryTicketBooth, $launchedQueries, $q) {
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
      } else {
        resultPromise.reject(null);
      }
    }).error(function(data, status, headers, config) {
      $launchedQueries.dec();
      if (ticket.valid()) {
        resultPromise.resolve(data);
      } else {
        resultPromise.reject(null);
      }
    });
    return resultPromise.promise;
  };
});

module.controller("commonCtrl", function(
  $chartWaiting,
  $launchedQueries,
  $location,
  $scope
  ){
  $scope.inputModel = {
    profile: false,
    query: "",
    renderType: "line"
  };
  $scope.hasProfiling = hasProfiling;
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
  $scope.setQueryResult = function(queryResult) {
    var options = {
      legend:    {position: "bottom"},
      title:     $location.search()["title"],
      chartArea: {left: "5%", width:"90%", top: "5%", height: "85%"}
    }
    if ($scope.inputModel.renderType === "area") {
      options.isStacked = true;
    }
    $scope.queryResult =   queryResult;
    $scope.selectOptions = options;
    $scope.selectResult =  convertSelectResponse(queryResult);
    $scope.profileResult = convertProfileResponse(queryResult);
  };
});

module.controller("mainCtrl", function(
  $chartWaiting,
  $controller,
  $google,
  $launchRequest,
  $launchedQueries,
  $location,
  $scope
  ) {
  $controller("commonCtrl", {$scope: $scope});
  $scope.queryHistory = [];
  $scope.embedLink = "";
  $scope.queryResult = "";

  function updateEmbed() {
    var url = $location.absUrl();
    var queryAt = url.indexOf("?");
    if (queryAt !== -1) {
      $scope.embedLink = $location.protocol() + "://" + $location.host() + ":" + $location.port() + "/embed" + url.substring(queryAt);
    } else {
      $scope.embedLink = "";
    }
  }

  // Triggers when the button is clicked.
  $scope.onSubmitQuery = function() {
    $scope.inputModel.query = document.getElementById("query-input").value;
    $location.search("query", $scope.inputModel.query);
    $location.search("renderType", $scope.inputModel.renderType);
    $location.search("profile", $scope.inputModel.profile.toString());
  };

  $scope.$on("$locationChangeSuccess", function() {
    // this triggers at least once (in the beginning).
    var queries = $location.search();
    $scope.inputModel.query = queries["query"] || "";
    $scope.inputModel.renderType = queries["renderType"] || "line";
    $scope.inputModel.profile = queries["profile"] === "true";
    // Add the query to the history, if it hasn't been seen before and it's non-empty
    var trimmedQuery = $scope.inputModel.query.trim();
    if (trimmedQuery !== "" && $scope.queryHistory.indexOf(trimmedQuery) === -1) {
      $scope.queryHistory.push(trimmedQuery);
    }
    if (trimmedQuery) {
      $launchRequest({
        profile: $scope.inputModel.profile,
        query:   $scope.inputModel.query
      }).then(function(data) {
        $scope.setQueryResult(data);
      });
    }
  });

  $scope.historySelect = function(query) {
    $scope.inputModel.query = query;
  }

  // true if the output should be tabular.
  $scope.isTabular = function() {
    return ["describe all", "describe metrics", "describe"].indexOf($scope.queryResult.name) >= 0;
  };
  updateEmbed()
});

module.controller("embedCtrl", function(
  $chartWaiting,
  $controller,
  $launchRequest,
  $launchedQueries,
  $location,
  $google,
  $scope
  ){
  $controller("commonCtrl", {$scope: $scope});
  $scope.queryResult =  null;

  var queries = $location.search();
  // Store the $inputModel so that the view can change it through inputs.
  $scope.inputModel.profile = false;
  $scope.inputModel.query = "";
  $scope.inputModel.renderType = queries["renderType"] || "line";
  $launchRequest({
    profile: false,
    query:   queries["query"] || ""
  }).then(function(data) {
    $scope.setQueryResult(data);
  });

  var url = $location.absUrl();
  var at = url.indexOf("?");
  $scope.metricsURL = $location.protocol() + "://" + $location.host() + ":" + $location.port()
    + "/ui" + url.substring(at);
  debugger;
});

// utility functions
function convertProfileResponse(object) {
  if (!(object && object.profile)) {
    return null
  }
  var dataTable = new google.visualization.DataTable();
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
  return dataTable;
}

function convertSelectResponse(object) {
  if (!(object && object.name == "select" &&
        object.body &&
        object.body.length &&
        object.body[0].series &&
        object.body[0].series.length &&
        object.body[0].timerange)) {
    // invalid data.
    return null;
  }
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
  return google.visualization.arrayToDataTable(table);
}


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

function hasProfiling(data) {
  return data && data.profile;
}

