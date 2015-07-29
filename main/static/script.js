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

var module = angular.module("main",[]);

var MAX_RENDERED = 200;

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
    template: "<div style='width:100%;height:100px'></div>",
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
          var data = scope.data();
          var option = scope.option();
          if (data && option) {
            $chartWaiting.inc();
            google.visualization.events.addListener(chart, "ready", function() {
              scope.$apply(function() { $chartWaiting.dec(); });
            });
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
  var keywords = [
    "all",
    "by",
    "collapse",
    "describe",
    "from",
    "group",
    "match",
    "metrics",
    "now",
    "resolution",
    "sample",
    "select",
    "to",
    "where"
  ];
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
    var start = new Date();
    var ticket = $queryTicketBooth.next();
    $launchedQueries.inc();
    var request = $http.get("/query", {
      params:params
    }).success(function(data, status, headers, config) {
      $launchedQueries.dec();
      resolve(data);
    }).error(function(data, status, headers, config) {
      $launchedQueries.dec();
      resolve(data);
    });
    function resolve(data) {
      var elapsedMs = new Date().getTime() - start.getTime();
      if (ticket.valid()) {
        resultPromise.resolve({elapsedMs: elapsedMs, payload: data});
      } else {
        resultPromise.reject(null);
      }
    }
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
  $scope.selectOptions = {
    legend:    {position: "bottom"},
    title:     $location.search()["title"],
    chartArea: {left: "5%", width:"90%", top: "5%", height: "80%"},
    series:    null,
    vAxes: {
      0: {title: ""},
      1: {title: ""}
    },
  };
  $scope.$watch("inputModel.renderType", function(newValue) {
    if (newValue === "area") {
      $scope.selectOptions.isStacked = true;
    } else {
      $scope.selectOptions.isStacked = false;
    }
  });
  $scope.timezone = (function() {
    if ($location.search().timezone) {
      return $location.search().timezone;
    } else {
      return "LOCAL";
    }
  })();
  $scope.displayTimezone = (function() {
    return getZoneName($scope.timezone);
  })();

  $scope.maxResult = MAX_RENDERED;
  $scope.setQueryResult = function(queryResult) {
    $scope.queryResult =   queryResult;
    var selectResponse = convertSelectResponse(queryResult, $scope.timezone);
    if (selectResponse) {
      $scope.selectResult = selectResponse.dataTable;
      $scope.selectOptions.series = selectResponse.seriesOptions;
    } else {
      $scope.selectResult = null;
      $scope.selectOptions.series = null;
    }
    $scope.totalResult = 0;
    $scope.profileResult = convertProfileResponse(queryResult);
    if ($scope.selectResult) {
      for (var i = 0; i < queryResult.body.length; i++) {
        // Each of these is a list of series
        $scope.totalResult += queryResult.body[i].series.length;
      }
    }
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
        $scope.setQueryResult(data.payload);
        $scope.elapsedMs = data.elapsedMs;
        updateEmbed();
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
  updateEmbed();
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
    $scope.setQueryResult(data.payload);
  });

  var url = $location.absUrl();
  var at = url.indexOf("?");
  $scope.metricsURL = $location.protocol() + "://" + $location.host() + ":" + $location.port()
    + "/ui" + url.substring(at);
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

function convertSelectResponse(object, timezone) {
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
        seriesOptions[series.length-1] = singleSeriesOption;
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
    var row = [dateFromIndex(t, timerange, timezone)];
    for (var i = 0; i < series.length; i++) {
      var cell = series[i].values[t];
      if (cell === null) {
        row.push(NaN);
      } else {
        row.push(parseFloat(cell.toFixed(2)));
      }
    }
    table.push(row);
  }
  var chartTable = google.visualization.arrayToDataTable(table);
  return {
    dataTable: chartTable,
    seriesOptions: seriesOptions
  }
}

var timezones = {
  EDT: -4,
  EST: -5,
  MDT: -6,
  MST: -7,
  PDT: -7,
  PST: -8,
  JST: +9,
  UTC: 0,
  ET: {
    D: "EDT",
    S: "EST"
  },
  MT: {
    D: "MDT",
    S: "MST"
  },
  PT: {
    D: "PDT",
    S: "PST"
  },
  LOCAL: -(new Date().getTimezoneOffset() / 60)
};
var extraTimezones = {
  "ACDT": "+10.5", // Australian Central Daylight Saving Time
  "ACST": "+09.5", // Australian Central Standard Time
  "ACT": "−05", // Acre Time
  "ACT": "+08", // ASEAN Common Time
  "ADT": "−03", // Atlantic Daylight Time
  "AEDT": "+11", // Australian Eastern Daylight Saving Time
  "AEST": "+10", // Australian Eastern Standard Time
  "AFT": "+04.5", // Afghanistan Time
  "AKDT": "−08", // Alaska Daylight Time
  "AKST": "−09", // Alaska Standard Time
  "AMST": "−03", // Amazon Summer Time (Brazil)[1]
  "AMST": "+05", // Armenia Summer Time
  "AMT": "−04", // Amazon Time (Brazil)[2]
  "AMT": "+04", // Armenia Time
  "ART": "−03", // Argentina Time
  "AST": "+03", // Arabia Standard Time
  "AST": "−04", // Atlantic Standard Time
  "AWDT": "+09", // Australian Western Daylight Time
  "AWST": "+08", // Australian Western Standard Time
  "AZOST": "−01", // Azores Standard Time
  "AZT": "+04", // Azerbaijan Time
  "BDT": "+08", // Brunei Time
  "BDT": "+06", // Bangladesh Daylight Time (Bangladesh Daylight saving time keeps UTC+06 offset) [3]
  "BIOT": "+06", // British Indian Ocean Time
  "BIT": "−12", // Baker Island Time
  "BOT": "−04", // Bolivia Time
  "BRST": "−02", // Brasilia Summer Time
  "BRT": "−03", // Brasilia Time
  "BST": "+06", // Bangladesh Standard Time
  "BST": "+11", // Bougainville Standard Time[4]
  "BST": "+01", // British Summer Time (British Standard Time from Feb 1968 to Oct 1971)
  "BTT": "+06", // Bhutan Time
  "CAT": "+02", // Central Africa Time
  "CCT": "+06.5", // Cocos Islands Time
  "CDT": "−05", // Central Daylight Time (North America)
  "CDT": "−04", // Cuba Daylight Time[5]
  "CEDT": "+02", // Central European Daylight Time
  "CEST": "+02", // Central European Summer Time (Cf. HAEC)
  "CET": "+01", // Central European Time
  "CHADT": "+13.75", // Chatham Daylight Time
  "CHAST": "+12.75", // Chatham Standard Time
  "CHOT": "+08", // Choibalsan
  "ChST": "+10", // Chamorro Standard Time
  "CHUT": "+10", // Chuuk Time
  "CIST": "−08", // Clipperton Island Standard Time
  "CIT": "+08", // Central Indonesia Time
  "CKT": "−10", // Cook Island Time
  "CLST": "−03", // Chile Summer Time
  "CLT": "−04", // Chile Standard Time
  "COST": "−04", // Colombia Summer Time
  "COT": "−05", // Colombia Time
  "CST": "−06", // Central Standard Time (North America)
  "CST": "+08", // China Standard Time
  "CST": "+09.5", // Central Standard Time (Australia)
  "CST": "+10.5", // Central Summer Time (Australia)
  "CST": "−05", // Cuba Standard Time
  "CT": "+08", // China time
  "CVT": "−01", // Cape Verde Time
  "CWST": "+08.75", // Central Western Standard Time (Australia) unofficial
  "CXT": "+07", // Christmas Island Time
  "DAVT": "+07", // Davis Time
  "DDUT": "+10", // Dumont d'Urville Time
  "DFT": "+01", // AIX specific equivalent of Central European Time[6]
  "EASST": "−05", // Easter Island Standard Summer Time
  "EAST": "−06", // Easter Island Standard Time
  "EAT": "+03", // East Africa Time
  "ECT": "−04", // Eastern Caribbean Time (does not recognise DST)
  "ECT": "−05", // Ecuador Time
  "EDT": "−04", // Eastern Daylight Time (North America)
  "EEDT": "+03", // Eastern European Daylight Time
  "EEST": "+03", // Eastern European Summer Time
  "EET": "+02", // Eastern European Time
  "EGST": "+00", // Eastern Greenland Summer Time
  "EGT": "−01", // Eastern Greenland Time
  "EIT": "+09", // Eastern Indonesian Time
  "EST": "−05", // Eastern Standard Time (North America)
  // "EST": "+10", // Eastern Standard Time (Australia) - aliases with American Eastern Standard Time
  "FET": "+03", // Further-eastern European Time
  "FJT": "+12", // Fiji Time
  "FKST": "−03", // Falkland Islands Standard Time
  "FKST": "−03", // Falkland Islands Summer Time
  "FKT": "−04", // Falkland Islands Time
  "FNT": "−02", // Fernando de Noronha Time
  "GALT": "−06", // Galapagos Time
  "GAMT": "−09", // Gambier Islands
  "GET": "+04", // Georgia Standard Time
  "GFT": "−03", // French Guiana Time
  "GILT": "+12", // Gilbert Island Time
  "GIT": "−09", // Gambier Island Time
  "GMT": "", // Greenwich Mean Time
  "GST": "−02", // South Georgia and the South Sandwich Islands
  "GST": "+04", // Gulf Standard Time
  "GYT": "−04", // Guyana Time
  "HADT": "−09", // Hawaii-Aleutian Daylight Time
  "HAEC": "+02", // Heure Avancée d'Europe Centrale francised name for CEST
  "HAST": "−10", // Hawaii-Aleutian Standard Time
  "HKT": "+08", // Hong Kong Time
  "HMT": "+05", // Heard and McDonald Islands Time
  "HOVT": "+07", // Khovd Time
  "HST": "−10", // Hawaii Standard Time
  "ICT": "+07", // Indochina Time
  "IDT": "+03", // Israel Daylight Time
  "IOT": "+03", // Indian Ocean Time
  "IRDT": "+04.5", // Iran Daylight Time
  "IRKT": "+08", // Irkutsk Time
  "IRST": "+03.5", // Iran Standard Time
  "IST": "+05.5", // Indian Standard Time
  "IST": "+01", // Irish Standard Time[7]
  "IST": "+02", // Israel Standard Time
  "JST": "+09", // Japan Standard Time
  "KGT": "+06", // Kyrgyzstan time
  "KOST": "+11", // Kosrae Time
  "KRAT": "+07", // Krasnoyarsk Time
  "KST": "+09", // Korea Standard Time
  "LHST": "+10.5", // Lord Howe Standard Time
  "LHST": "+11", // Lord Howe Summer Time
  "LINT": "+14", // Line Islands Time
  "MAGT": "+12", // Magadan Time
  "MART": "−09.5", // Marquesas Islands Time
  "MAWT": "+05", // Mawson Station Time
  "MDT": "−06", // Mountain Daylight Time (North America)
  "MET": "+01", // Middle European Time Same zone as CET
  "MEST": "+02", // Middle European Saving Time Same zone as CEST
  "MHT": "+12", // Marshall Islands
  "MIST": "+11", // Macquarie Island Station Time
  "MIT": "−09.5", // Marquesas Islands Time
  "MMT": "+06.5", // Myanmar Time
  "MSK": "+03", // Moscow Time
  // "MST": "+08", // Malaysia Standard Time - aliases with American Mountain Standard Time
  "MST": "−07", // Mountain Standard Time (North America)
  // "MST": "+06.5", // Myanmar Standard Time - aliases with American Mountain Standard Time
  "MUT": "+04", // Mauritius Time
  "MVT": "+05", // Maldives Time
  "MYT": "+08", // Malaysia Time
  "NCT": "+11", // New Caledonia Time
  "NDT": "−02.5", // Newfoundland Daylight Time
  "NFT": "+11.5", // Norfolk Time
  "NPT": "+05.75", // Nepal Time
  "NST": "−03.5", // Newfoundland Standard Time
  "NT": "−03.5", // Newfoundland Time
  "NUT": "−11", // Niue Time
  "NZDT": "+13", // New Zealand Daylight Time
  "NZST": "+12", // New Zealand Standard Time
  "OMST": "+06", // Omsk Time
  "ORAT": "+05", // Oral Time
  "PDT": "−07", // Pacific Daylight Time (North America)
  "PET": "−05", // Peru Time
  "PETT": "+12", // Kamchatka Time
  "PGT": "+10", // Papua New Guinea Time
  "PHOT": "+13", // Phoenix Island Time
  "PKT": "+05", // Pakistan Standard Time
  "PMDT": "−02", // Saint Pierre and Miquelon Daylight time
  "PMST": "−03", // Saint Pierre and Miquelon Standard Time
  "PONT": "+11", // Pohnpei Standard Time
  "PST": "−08", // Pacific Standard Time (North America)
  // "PST": "+08", // Philippine Standard Time - aliases with Pacific Standard Type
  "PYST": "−03", // Paraguay Summer Time (South America)[8]
  "PYT": "−04", // Paraguay Time (South America)[9]
  "RET": "+04", // Réunion Time
  "ROTT": "−03", // Rothera Research Station Time
  "SAKT": "+11", // Sakhalin Island time
  "SAMT": "+04", // Samara Time
  "SAST": "+02", // South African Standard Time
  "SBT": "+11", // Solomon Islands Time
  "SCT": "+04", // Seychelles Time
  "SGT": "+08", // Singapore Time
  "SLST": "+05.5", // Sri Lanka Standard Time
  "SRET": "+11", // Srednekolymsk Time
  "SRT": "−03", // Suriname Time
  "SST": "−11", // Samoa Standard Time
  "SST": "+08", // Singapore Standard Time
  "SYOT": "+03", // Showa Station Time
  "TAHT": "−10", // Tahiti Time
  "THA": "+07", // Thailand Standard Time
  "TFT": "+05", // Indian/Kerguelen
  "TJT": "+05", // Tajikistan Time
  "TKT": "+13", // Tokelau Time
  "TLT": "+09", // Timor Leste Time
  "TMT": "+05", // Turkmenistan Time
  "TOT": "+13", // Tonga Time
  "TVT": "+12", // Tuvalu Time
  "UCT": "", // Coordinated Universal Time
  "ULAT": "+08", // Ulaanbaatar Time
  "USZ1": "+02", // Kaliningrad Time
  "UTC": "0", // Coordinated Universal Time
  "UYST": "−02", // Uruguay Summer Time
  "UYT": "−03", // Uruguay Standard Time
  "UZT": "+05", // Uzbekistan Time
  "VET": "−04.5", // Venezuelan Standard Time
  "VLAT": "+10", // Vladivostok Time
  "VOLT": "+04", // Volgograd Time
  "VOST": "+06", // Vostok Station Time
  "VUT": "+11", // Vanuatu Time
  "WAKT": "+12", // Wake Island Time
  "WAST": "+02", // West Africa Summer Time
  "WAT": "+01", // West Africa Time
  "WEDT": "+01", // Western European Daylight Time
  "WEST": "+01", // Western European Summer Time
  "WET": "", // Western European Time
  "WIT": "+07", // Western Indonesian Time
  "WST": "+08", // Western Standard Time
  "YAKT": "+09", // Yakutsk Time
  "YEKT": "+05", // Yekaterinburg Time
  "Z": "" // Zulu Time (Coordinated Universal Time)
}


function isLocalDaylightSavings() {
  return new Date().getTimezoneOffset() == new Date( new Date().getFullYear(), 6, 1 ).getTimezoneOffset();
}

function getZoneName(zone) {
  if (parseFloat(zone)) {
    if (zone == 0) {
     return "UTC"; 
    }
    var positive = zone > 0;
    var hour = (Math.abs(zone) | 0) + "";
    var minutes = Math.abs(zone) % 1 * 60 + "";
    if (hour.length == 1) {
      hour = "0" + hour;
    }
    if (minutes.length == 1) {
      minutes = "0" + minutes;
    }
    return "UTC" + (positive ? "+" : "-") + hour + ":" + minutes;
  }
  zone = zone.toUpperCase();
  if (zone === "LOCAL") {
    return getZoneName(timezones["LOCAL"]);
  }
  while (typeof timezones[zone] == "object") {
    if (isLocalDaylightSavings()) {
      zone = timezones[zone].D;
    } else {
      zone = timezones[zone].S;
    }
  }
  return zone;
}
function getOffset(zone) {
  if (parseFloat(zone)) {
    return parseFloat(zone);
  }
  if (zone == "LOCAL") {
    return timezones["LOCAL"];
  }
  var result = timezones[getZoneName(zone)] || extraTimezones[zone];
  if (result === undefined) {
    return 0;
  }
  return result;
}

function dateFromIndex(index, timerange, timezone) {
  if (!parseFloat(timezone)) {
    timezone = getOffset(timezone);
  }
  var offset = new Date().getTimezoneOffset() * 60 * 1000;
  return new Date(timerange.start + timerange.resolution * index + offset + timezone * 60 * 60 * 1000);
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

