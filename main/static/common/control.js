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

module.directive("autocom", function($http) {
  console.log("autocom");
  return {
    template: "",
    restrict: "A",
    link: function(scope, element) {
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
  $controller,
  $location,
  $scope
) {
  $controller("commonController", {$scope: $scope});
  //
  $scope.queryHistory = [];

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

  $scope.selectOptions = {
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

  _receiveProfile(function(profile) {
    $scope.profileMode = "rendering";
    _chart.sets("profile", {
      profile: profile,
      chartType: "timeline",
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

  $scope.historySelect = function(query) {
    $scope.inputModel.query = query;
    $scope.onSubmitQuery();
  };

  $scope.onSubmitQuery();

  _chart.addListener("select/waiting", function(state) {
    $scope.selectRendered = !state.waiting;
    console.log(state);
    if (state.data) {
      $scope.renderedCount = state.data.renderedCount;
      $scope.totalCount = state.data.totalCount; 
    }
  });

  $scope.getEmbedLink = function() {
    var url = $location.absUrl();
    var queryAt = url.indexOf("?");
    if (queryAt !== -1) {
      return $location.protocol() + "://" + $location.host() + ":" + $location.port() + "/example/embed.html" + url.substring(queryAt);
    } else {
      return "";
    }
  };
});

module.controller("embedController", function(
  _launchRequest,
  $controller,
  $location,
  $scope
  ) {
  $controller("commonController", {$scope: $scope});
  // Send the request for the query in the URL
  $scope.inputModel = {
    renderType: $location.search().renderType,
    query: $location.search().query,
  };

  function applyDefault(name, value) {
    if ($location.search()[name] !== undefined) {
      return $location.search()[name];
    }
    return value;
  }

  $scope.getUILink = function() {
    var url = $location.absUrl();
    var queryAt = url.indexOf("?");
    if (queryAt !== -1) {
      return $location.protocol() + "://" + $location.host() + ":" + $location.port() + "/example/ui.html" + url.substring(queryAt);
    } else {
      return "";
    }
  };

  $scope.selectOptions = {
    legend:    {position: "bottom"},
    title:     applyDefault("title", ""),
    chartArea: {
      left: applyDefault("marginleft", "15px"),
      right: applyDefault("marginright", "50px"),
      top: applyDefault("margintop", "35px"),
      bottom: applyDefault("marginbottom", "50px")
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

  _launchRequest({
    profile: false,
    query: $location.search().query
  });
});

module.controller("commonController", function(
  _receiveSuccess,
  _receiveFailure,
  _chart,
  $scope
  ) {
  _receiveSuccess(function(name, body, latency) {
    $scope.mode = name;
    $scope.result = body;
    $scope.resultLatency = latency;
    if (name == "select") {
      _chart.sets("select", {
        body: body,
        chartType: $scope.inputModel.renderType,
        option: $scope.selectOptions
      });
    }
    $scope.profileMode = "waiting";
  });
  _receiveFailure(function(message) {
    $scope.mode = "error";
    $scope.errorMessage = message;
    $scope.profileMode = "waiting";
  });
});
