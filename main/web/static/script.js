// Copyright 2015 - 2016 Square Inc.
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

/*jshint -W117 */
"use strict";

var MAX_RENDERED = 200;

google.load("visualization", "1.0", {
  "packages": ["corechart", "timeline"]
});

var module = angular.module("main", ["ngMaterial", "ngMessages", "ngSanitize", "LocalStorageModule"])
  .config(function ($mdThemingProvider) {
    $mdThemingProvider.theme('default')
      .primaryPalette('green');
  });

module.config(function ($locationProvider) {
  $locationProvider.html5Mode(true);
});

module.config(function (localStorageServiceProvider) {
  localStorageServiceProvider.setPrefix('mqe');
});

module.config(['$compileProvider', function ($compileProvider) {
  $compileProvider.debugInfoEnabled(false);
}]);

module.factory("$windowSize", function ($window) {
  return {
    height: $window.innerHeight,
    width: $window.innerWidth,
    version: 0 // updated whenever width or height is updated, so this object can be watched.
  };
});

// enable autocomplete on query input text area.
// TODO: reduce amount of magic and add tests
module.directive('queryInput', function ($http) {
  return {
    restrict: 'A',
    replace: 'false',
    link: function (scope, elem, attrs) {
      var autocom = new Autocom(elem[0]);
      var keywords = [
        "all", "by", "collapse", "describe", "from", "group", "match", "metrics",
        "now", "resolution", "sample", "select", "to", "where"
      ];
      var latterKeywords = [
        "from", "match", "now", "resolution", "sample", "by", "to",
      ];
      autocom.options = keywords.slice(0); // note: copies the array
      autocom.prefixPattern = "`[a-zA-Z_][a-zA-Z._-]*`?|[a-zA-Z_][a-zA-Z._-]*";
      autocom.continuePattern = "[a-zA-Z_`.-]";
      autocom.tooltipX = 0;
      autocom.tooltipY = 20;
      autocom.config.skipWord = 0.05; // make it (5x) cheaper to skip letters in a candidate word
      autocom.config.skipWordEnd = 0.01; // add a small cost to skipping ends of words, which benefits shorter candidates
      autocom.activeRegion = function (beforeText, afterText, candidate) {
        if (beforeText.match(/\s(where|from|to|resolution|sample)\s/)) { // Note: only works 99% of the time.
          return latterKeywords.indexOf(candidate) >= 0;
        }
        return true;
      };
      $http.get("/token").success(function (data, status, headers, config) {
        if (!data.success || !data.body) {
          return;
        }
        if (data.body.functions) {
          autocom.options = autocom.options.concat(data.body.functions);
        }
        var tolerate = [];
        if (data.body.metrics) {
          autocom.options = autocom.options.concat(data.body.metrics.map(function (name) {
            if (name.indexOf("-") >= 0) {
              return "`" + name + "`";
            } else {
              tolerate.push("`" + name + "`");
            }
            return name;
          }));
        }
        autocom.tolerate = tolerate;
      });
    }
  };
});

module.directive("googleChart", function ($chartWaiting, $timeout, $windowSize) {
  return {
    restrict: "E",
    template: "<div style='width:100%;height:100px'></div>",
    scope: {
      chartType: "&",
      data: "&",
      option: "&"
    },
    link: function (scope, element, attrs) {
      var chart = null;
      scope.$watch("option()", function (newValue) {
        render();
      }, true);
      scope.$watch("data()", function (newValue) {
        render();
      });
      scope.$watch(function () {
        return $windowSize.version;
      }, function (newValue) {
        render();
      });
      scope.$watch("chartType()", function (newValue) {
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

      function getUnits(value) {
        if (typeof value !== "string") {
          return null;
        }
        var result = value.match(/^([0-9.-]+)(%|px)$/);
        if (result === null) {
          return null;
        } else {
          return {
            value: parseFloat(result[1]),
            units: result[2]
          };
        }
      }

      function unitless(value) {
        return getUnits(value).value;
      }

      function fixUnits(value, total) {
        var match = getUnits(value);
        if (match === null) {
          return null;
        }
        switch (match.units) {
        case "px":
          return (match.value / total * 100) + "%";
        case "%":
          return match.value + "%";
        }
        throw "not accessible";
      }

      function deepCopy(thing) {
        if (typeof thing != "object") {
          return thing;
        }
        if (thing instanceof Array) {
          return thing;
        }
        var copy = {};
        for (var i in thing) {
          copy[i] = deepCopy(thing[i]);
        }
        return copy;
      }

      function render() {
        $timeout(function () {
          var data = scope.data();
          var option = scope.option();
          if (data && option) {
            $chartWaiting.inc();
            google.visualization.events.addListener(chart, "ready", function () {
              scope.$apply(function () {
                $chartWaiting.dec();
              });
            });
            var elementStyle = getComputedStyle(element[0]);
            var totalWidth = unitless(elementStyle.width) * 1;
            var totalHeight = unitless(elementStyle.height) * 1;
            option = deepCopy(option);
            if (option && option.chartArea) {
              var area = option.chartArea;
              var left = fixUnits(area.left, totalWidth);
              var top = fixUnits(area.top, totalHeight);
              var right = fixUnits(area.right, totalWidth);
              var bottom = fixUnits(area.bottom, totalHeight);
              var width = fixUnits(area.width, totalWidth);
              var height = fixUnits(area.height, totalHeight);
              if (right !== undefined) {
                width = (100 - unitless(left) - unitless(right)) + "%";
              }
              if (bottom !== undefined) {
                height = (100 - unitless(top) - unitless(bottom)) + "%";
              }
              option.chartArea = {
                left: left,
                top: top,
                width: width,
                height: height
              };
            }
            chart.draw(data, option);
          }
        }, 1);
      }
    }
  };
});

module.run(function ($window, $timeout, $windowSize) {
  var DELAY_MS = 100;
  var counter = 0;
  angular.element($window).bind("resize", function () {
    counter++;
    var currentCounter = counter; // capture the current value via the closure.
    $timeout(function () {
      if (currentCounter == counter) {
        $windowSize.height = $window.innerHeight;
        $windowSize.width = $window.innerWidth;
        $windowSize.version++;
      }
    }, DELAY_MS);
  });
});

module.factory("$google", function ($rootScope) {
  // abstraction over the async loading of google libraries.
  // registered functions are either invoked immediately (if the library finished loading).
  // or queued in an array.
  var googleFunctions = [];
  var googleLoaded = false;
  google.setOnLoadCallback(function () {
    $rootScope.$apply(function () {
      googleLoaded = true;
      googleFunctions.map(function (googleFunction) {
        googleFunction();
      });
      googleFunctions.length = 0; // clear the array.
    });
  });

  return function (func) {
    if (googleLoaded) {
      func();
    } else {
      googleFunctions.push(func);
    }
  };
});

// A counter is an object with .inc() and .dec() methods,
// as well as .pos() and .zero() predicates.
function Counter() {
  var count = 0;
  this.inc = function () {
    count++;
  };
  this.dec = function () {
    count--;
  };
  this.pos = function () {
    return count > 0;
  };
  this.zero = function () {
    return count === 0;
  };
};

// A singleton Counter for launched queries.
module.factory("$launchedQueries", function () {
  return new Counter();
});

// A singleton Counter for waiting charts
module.factory("$chartWaiting", function () {
  return new Counter();
});

module.service("timedQuery", function ($http, $q, $launchedQueries) {
  return function (params) {
    var resultPromise = $q.defer(); // Will be resolved with received value.
    var start = new Date();
    $launchedQueries.inc();
    var request = $http.get("/query", {
      params: params
    }).success(function (data, status, headers, config) {
      $launchedQueries.dec();
      resolve(data);
    }).error(function (data, status, headers, config) {
      $launchedQueries.dec();
      resolve(data);
    });

    function resolve(data) {
      var elapsedMs = new Date().getTime() - start.getTime();
      resultPromise.resolve({
        elapsedMs: elapsedMs,
        payload: data
      });
    }
    return resultPromise.promise;
  };
});

module.controller("CommonController", function (
  $chartWaiting,
  $launchedQueries,
  $location,
  $scope
) {
  $scope.inputModel = {
    profile: false,
    query: "",
    renderType: "line"
  };
  $scope.hasProfiling = hasProfiling;
  $scope.screenState = function () {
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
  $scope.hidden = {
    all: ($location.search().hideall || "").toLowerCase() == "true",
  };
  $scope.hidden = {
    explore: ($location.search().explore || "").toLowerCase() == "hide" || $scope.hidden.all,
    legend: ($location.search().legend || "").toLowerCase() == "hide" || $scope.hidden.all,
    xaxis: ($location.search().xaxis || "").toLowerCase() == "hide" || $scope.hidden.all,
    yaxis: ($location.search().yaxis || "").toLowerCase() == "hide" || $scope.hidden.all,
  };

  $scope.applyDefault = function (name, value) {
    if ($location.search()[name] !== undefined) {
      return $location.search()[name];
    }
    return value;
  };

  $scope.selectOptions = {
    legend: {
      position: $scope.hidden.legend ? "none" : "bottom"
    },
    title: $scope.applyDefault("title", ""),
    chartArea: {
      left: $scope.applyDefault("marginleft", "10px"),
      right: $scope.applyDefault("marginright", $scope.hidden.yaxis ? "10px" : "50px"),
      top: $scope.applyDefault("margintop", "10px"),
      bottom: $scope.applyDefault("marginbottom", (($scope.hidden.legend ? 15 : 25) + ($scope.hidden.xaxis ? 0 : 15)) + "px")
    },
    series: null,
    vAxes: {
      0: {
        title: ""
      },
      1: {
        title: ""
      }
    },
    vAxis: {
      textPosition: $scope.hidden.yaxis ? "none" : "out",
      format: "short"
    },
    hAxis: {
      textPosition: $scope.hidden.xaxis ? "none" : "out"
    }
  };
  $scope.queryResultIsEmpty = function () {
    var result = $scope.queryResult;
    if (!result || result.name != "select") {
      return false;
    }
    for (var i = 0; i < result.body.length; i++) {
      if (result.body[i].series.length === 0) {
        if (result.body.length == 1) {
          $scope.queryEmptyMessage = "the query resulted in 0 series";
        } else {
          $scope.queryEmptyMessage = "expression " + (i + 1) + " (of " + result.body.length + ") resulted in 0 series";
        }
        return true;
      }
    }
    return false;
  };
  $scope.$watch("inputModel.renderType", function (newValue) {
    if (newValue === "area") {
      $scope.selectOptions.isStacked = true;
    } else {
      $scope.selectOptions.isStacked = false;
    }
  });

  $scope.maxResult = MAX_RENDERED;
  $scope.setQueryResult = function (queryResult) {
    $scope.queryResult = queryResult;
    var selectResponse = convertSelectResponse(queryResult);
    if (selectResponse) {
      $scope.selectResult = selectResponse.dataTable;
      $scope.selectOptions.series = selectResponse.seriesOptions;
    } else {
      $scope.selectResult = null;
      $scope.selectOptions.series = null;
    }
    $scope.totalSeriesCount = 0;
    $scope.totalScalarsCount = 0;
    $scope.profileResult = convertProfileResponse(queryResult);
    if (queryResult && queryResult.body) {
      for (var i = 0; i < queryResult.body.length; i++) {
        // Each of these is a list of series
        if (queryResult.body[i].type == "series") {
          $scope.totalSeriesCount += queryResult.body[i].series.length;
        }
        if (queryResult.body[i].type == "scalars") {
          $scope.totalScalarsCount += queryResult.body[i].scalars.length;
        }
      }
    }
  };
});

module.controller("DiscoverController", function (
  $controller,
  $location,
  $scope,
  timedQuery
) {
  $scope.tags = {};
  $scope.metricSelected = false;
  $scope.metrics = [];
  $scope.models = {};
  $scope.tagSearchText = {};
  $scope.to = new Date();
  $scope.from = new Date($scope.to.getTime() - 3600*1000); // -1h

  $scope.getMetrics = function() {
    if($scope.searchText) {
      return _.filter($scope.metrics, function(metric) { return s(metric).include($scope.searchText); });
    }
    else {
      return $scope.metrics;
    }
  };

  $scope.getTags = function() {
    return $scope.tags;
  }

  $scope.describe = function (metric) {
    $scope.models = {};
    $scope.metricSelected = true;
    timedQuery({
      profile: false,
      query: "describe " + "`" + metric + "`"
    }).then(
      function (data) {
        $scope.tags = data.payload.body;
      },
      function (error) {
        console.log("MQE request failed: " + error);
      }
    );
  };

  $scope.onTagSelect = function() {
    // Generate a query and update available tags
    var whereClause = [];
    _.each(_.pairs($scope.models), function(pair) {
      if(pair[1]) {
        whereClause.push(pair[0] + " = '" + pair[1] + "'");
      }
    });
    var query = "describe " + "`" + $scope.metric + "`";
    if(whereClause.length > 0) {
      query += " where " + whereClause.join(" and ");
    }
    timedQuery({
      profile: false,
      query: query
    }).then(
      function (data) {
          if(_.keys(data.payload.body).length > 0) {
            $scope.tags = data.payload.body;
          }
      },
      function (error) {
        console.log("MQE request failed: " + error);
      }
    );
  }

  $scope.onSubmit = function() {
    // Generate a query and call parent scope
    var whereClause = [];
    _.each(_.pairs($scope.models), function(pair) {
      if(pair[1]) {
        whereClause.push(pair[0] + " = '" + pair[1] + "'");
      }
    });
    var query = "select " + "`" + $scope.metric + "`";
    if(whereClause.length > 0) {
      query += " where " + whereClause.join(" and ");
    }
    query += " from '" + $scope.from.getTime() + "' to '" + $scope.to.getTime() + "';";
    $scope.onSubmitQuery(query);
  };

  timedQuery({
    profile: false,
    query: "describe all"
  }).then(
    function (data) {
      $scope.metrics = data.payload.body;
    },
    function (error) {
      console.log("MQE request failed: " + error);
    }
  );


});

module.controller("MainController", function (
  $chartWaiting,
  $controller,
  $google,
  $launchedQueries,
  $location,
  $mdDialog,
  $scope,
  timedQuery,
  $window,
  localStorageService
) {
  $controller("CommonController", {
    $scope: $scope
  });

  $scope.localStorageService = localStorageService;

  function DialogController($scope, $mdDialog, query, localStorageService) {
    $scope.hide = function () {
      $mdDialog.hide();
    };
    $scope.cancel = function () {
      $mdDialog.cancel();
    };
    $scope.save = function (answer) {
      localStorageService.set($scope.name, query);
      $mdDialog.hide(answer);
    };
  }

  $scope.showSaveDialog = function (ev) {
    $mdDialog.show({
        controller: DialogController,
        template: '<md-dialog id="save-dialog">' +
          '<md-dialog-content class="md-dialog-content">' +
            '<h2 class="md-title">Save Query</h2>' +
            '<md-input-container>' +
              '<label>Name</label>' +
              '<input ng-model="name"></input>' +
            '</md-input-container>' +
          '</md-dialog-content>' +
          '<md-dialog-actions>' +
            '<md-button class="md-raised" ng-click="cancel()">' +
              'Cancel' +
            '</md-button>' +
            '<md-button class="md-raised md-primary" ng-click="save()">' +
              'Save' +
            '</md-button>' +
            '</md-dialog-actions>' +
          '</md-dialog>',
        parent: angular.element(document.body),
        targetEvent: ev,
        clickOutsideToClose: true,
        locals: {
          query: $scope.query
        },
      });
  };

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
  $scope.onSubmitQuery = function (query) {
    if (query) {
      $scope.inputModel.query = query;
    }
    $location.search("query", $scope.inputModel.query);
    $location.search("renderType", $scope.inputModel.renderType);
    $location.search("profile", $scope.inputModel.profile.toString());
  };

  $scope.$on("$locationChangeSuccess", function () {
    // this triggers at least once (in the beginning).
    var queries = $location.search();
    $scope.inputModel.query = queries.query || "";
    $scope.inputModel.renderType = queries.renderType || "line";
    $scope.inputModel.profile = queries.profile === "true";
    var trimmedQuery = $scope.inputModel.query.trim();
    if (trimmedQuery) {
      timedQuery({
        profile: $scope.inputModel.profile,
        query: $scope.inputModel.query
      }).then(function (data) {
        $scope.setQueryResult(data.payload);
        $scope.elapsedMs = data.elapsedMs;
        updateEmbed();
      }, function (error) { console.log(error); });
    }
  });

  $scope.historySelect = function (query) {
    $scope.inputModel.query = query;
  };

  // true if the output should be tabular.
  $scope.isTabular = function () {
    return ["describe all", "describe metrics", "describe"].indexOf($scope.queryResult.name) >= 0;
  };
  updateEmbed();
});

module.controller("QueryController", function() {

});

module.controller("EmbedController", function (
  $chartWaiting,
  $controller,
  timedQuery,
  $launchedQueries,
  $location,
  $google,
  $scope
) {
  $controller("CommonController", {
    $scope: $scope
  });
  $scope.queryResult = null;

  $scope.selectOptions.chartArea.top = $scope.applyDefault("margintop", $scope.hidden.explore ? "20px" : "40px");

  var queries = $location.search();
  // Store the $inputModel so that the view can change it through inputs.
  $scope.inputModel.profile = false;
  $scope.inputModel.query = "";
  $scope.inputModel.renderType = queries.renderType || "line";
  timedQuery({
    profile: false,
    query: queries.query || ""
  }).then(function (data) {
    $scope.setQueryResult(data.payload);
  });

  var url = $location.absUrl();
  var at = url.indexOf("?");
  $scope.metricsURL = $location.protocol() + "://" + $location.host() + ":" + $location.port() + "/ui" + url.substring(at);
});

// utility functions
function convertProfileResponse(object) {
  if (!(object && object.profile)) {
    return null;
  }
  var dataTable = new google.visualization.DataTable();
  dataTable.addColumn({
    type: 'string',
    id: 'Name'
  });
  dataTable.addColumn({
    type: 'number',
    id: 'Start'
  });
  dataTable.addColumn({
    type: 'number',
    id: 'End'
  });
  var minValue = Number.POSITIVE_INFINITY;
  function epoch(timestamp) {
    return new Date(timestamp).getTime();
  }
  for (var i = 0; i < object.profile.length; i++) {
    var profile = object.profile[i];
    minValue = Math.min(epoch(profile.start), minValue);
    minValue = Math.min(epoch(profile.finish), minValue);
  }
  function normalize(value) {
    return value - minValue;
  }
  for (i = 0; i < object.profile.length; i++) {
    var profile = object.profile[i];
    var row = [profile.name + " - " + profile.description, normalize(epoch(profile.start)), normalize(epoch(profile.finish))];
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
  var seriesOptions = {};
  var series = [];
  var labels = ["Time"];
  var table = [labels];
  var onlySingleSeries = object.body.length === 1;
  for (var i = 0; i < object.body.length; i++) {
    // Each of these is a list of series
    var serieslist = object.body[i];
    for (var j = 0; j < serieslist.series.length; j++) {
      if (series.length < MAX_RENDERED) {
        var s = object.body[i].series[j];
        var singleSeriesOption = {};
        series.push(s);
        seriesOptions[series.length - 1] = singleSeriesOption;
        // special tags.
        if (s.tagset.$secondaxis === "true") {
          singleSeriesOption.targetAxisIndex = 0;
        } else {
          singleSeriesOption.targetAxisIndex = 1;
        }
        if (s.tagset.$color) {
          singleSeriesOption.color = s.tagset.$color;
        }
        if (s.tagset.$linewidth) {
          singleSeriesOption.lineWidth = parseFloat(s.tagset.$linewidth);
        }
        labels.push(makeLabel(onlySingleSeries, serieslist, s));
      } else {
        break;
      }
    }
  }
  // Next, add each row.
  var timerange = object.body[0].timerange;
  for (var t = 0; t < series[0].values.length; t++) {
    var row = [dateFromIndex(t, timerange)];
    for (i = 0; i < series.length; i++) {
      var cell = series[i].values[t];
      if (cell === null) {
        row.push(NaN);
      } else {
        row.push(parseFloat(cell.toFixed(2)));
      }
    }
    table.push(row);
  }
  return {
    dataTable: google.visualization.arrayToDataTable(table),
    seriesOptions: seriesOptions
  };
}


function dateFromIndex(index, timerange) {
  return new Date(timerange.start + timerange.resolution * index);
}

function makeLabel(onlySingleSeries, serieslist, series) {
  var tagsets = [];
  var label;
  for (var key in series.tagset) {
    if (key[0] !== "$") {
      tagsets.push(key + ":" + series.tagset[key]);
    }
  }
  if (onlySingleSeries) {
    if (tagsets.length > 0) {
      // for a single graph, only return the tags.
      return tagsets.join(" ");
    }
    // if no tags, then return the name.
    return serieslist.name;
  } else {
    // return both name and tags.
    return serieslist.name + " " + tagsets.join(" ");
  }
  return label;
}

function hasProfiling(data) {
  return data && data.profile;
}
