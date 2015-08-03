// Copyright 2015 Square Inc
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

module.factory("_windowSize", function($window) {
  return {
    height:  $window.innerHeight,
    width:   $window.innerWidth,
    version: 0 // updated whenever width or height is updated, so this object can be watched.
  }
});

module.directive("myAutocom", function($http) {
  console.log("autocom directive");
  return {
    template: "",
    restrict: "A",
    link: function(scope, element) {
      console.log("autocom");
      var input = element[0];
      var autocom = new Autocom(input);
      autocom.options = ["describe", "all", "metrics", "where", "now", "from", "to"];
      autocom.prefixPattern = "`[a-zA-Z_][a-zA-Z._-]*`?|[a-zA-Z_][a-zA-Z._-]*";
      autocom.continuePattern = "[a-zA-Z_`.-]";
      autocom.tooltipX = 0;
      autocom.tooltipY = 20;
      autocom.config.skipWord = 0.05; // make it 5x cheaper to skip letters in a candidate word
      autocom.config.skipWordEnd = 0.01; // add a small cost to skipping ends of words, which benefits shorter candidates
      autocom.hook = function() {
        scope.inputModel.query = input.value; // A bit hacky, but everything else is more Angular-y
      };
      $http.get("/token").success(function(data) {
        if (!data.success || !data.body) {
          return;
        }
        if (data.body.functions) {
          autocom.options = autocom.options.concat(data.body.functions);
        }
        if (data.body.metrics) {
          autocom.options = autocom.options.concat( data.body.metrics.map(function(name) {
            if (name.indexOf("-") >= 0) {
              return "`" + name + "`";
            } else {
              return name;
            }
          }));
        }
      });
    }
  }
});

module.controller("uiController", function(
  _receiveSuccess,
  _receiveFailure,
  _receiveProfile,
  _chart,
  _launchRequest,
  $scope,
  $location
) {
  //

  function applyDefault(name, value) {
    if ($location.search()[name] !== undefined) {
      return $location.search()[name];
    }
    return value;
  }

  $scope.inputModel = {
    profile: applyDefault("profile", false) == "true",
    query: applyDefault("query", ""),
    renderType: applyDefault("renderType", "line")
  };

  var selectOptions = {
    legend:    {position: "bottom"},
    title:     applyDefault("title", ""),
    chartArea: {
      left: applyDefault("marginleft", "50px"),
      right: applyDefault("marginright", "50px"),
      top: applyDefault("margintop", "10px"),
      bottom: applyDefault("marginbottom", "45px")
    },
    series:    null,
    vAxes: {
      0: {title: ""},
      1: {title: ""}
    },
    vAxis: {
      textPosition: "out",
      format: "short"
    },
    hAxis: {
      textPosition: "out"
    },
  };
  
  $scope.$watch("inputModel.renderType", function(newValue) {
    _chart.set("select", "chartType", newValue);
    $scope.selectRendered = false;
    $location.search("renderType", $scope.inputModel.renderType);
  });

  _receiveSuccess(function(name, body, latency) {
    $scope.mode = name;
    $scope.result = body;
    $scope.resultLatency = latency;
    if (name == "select") {
      var converted = convertSelectResponse(body);
      selectOptions.series = converted.options;
      _chart.sets("select", {
        data: converted.dataTable,
        chartType: $scope.inputModel.renderType,
        option: selectOptions
      });
    }
    $scope.profileMode = "waiting";
  });
  _receiveFailure(function(message) {
    $scope.mode = "error";
    $scope.errorMessage = message;
    $scope.profileMode = "waiting";
  });
  _receiveProfile(function(profile) {
    $scope.profileMode = "rendering";
    var converted = convertProfileResponse(profile);
    _chart.sets("profile", {
      data: converted,
      chartType: "timeline",
      option: converted
    });
  });

  $scope.onSubmitQuery = function() {
    // Trim the query
    var trimmedQuery = $scope.inputModel.query.trim();
    if (trimmedQuery !== "" && $scope.queryHistory.indexOf(trimmedQuery) === -1) {
      $scope.queryHistory.push(trimmedQuery);
    }
    if (trimmedQuery == "") {
      return; // stop without doing anything
    }
    // Send the request to MQE
    _launchRequest({
      profile: $scope.inputModel.profile,
      query: $scope.inputModel.query
    });
    // Mark the rendering state as "loading"
    $scope.mode = "loading";
    $scope.profileMode = "waiting";
    $scope.selectRendered = false;
    // Save the parameters in the URL
    $location.search("query", $scope.inputModel.query);
    $location.search("renderType", $scope.inputModel.renderType);
    $location.search("profile", $scope.inputModel.profile.toString());
  };

  $scope.onSubmitQuery();

  _chart.addListener("select/waiting", function(state) {
    $scope.selectRendered = !state.waiting;
  });

  $scope.getEmbedLink = function() {
    var url = $location.absUrl();
    var queryAt = url.indexOf("?");
    if (queryAt !== -1) {
      return $location.protocol() + "://" + $location.host() + ":" + $location.port() + "/embed" + url.substring(queryAt);
    } else {
      return "";
    }
  }
});


module.controller("mainCtrl", function(
  _chartWaiting,
  $controller,
  _google,
  _launchRequest,
  _launchedQueries,
  $location,
  $scope
  ) {
  $controller("commonCtrl", {$scope: $scope});
  $scope.queryHistory = [];
  $scope.queryResult = "";

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
      _launchRequest({
        profile: $scope.inputModel.profile,
        query:   $scope.inputModel.query
      }).then(function(data) {
        $scope.setQueryResult(data.payload);
        $scope.elapsedMs = data.elapsedMs;
        updateEmbed();
      });
    }
  });

  $scope.historySelect = function(query) {
    $scope.inputModel.query = query;
  }

});
