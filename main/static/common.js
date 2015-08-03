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
"use strict"

// This file contains minimal logic needed to make a UI for MQE.
// It should be easy to extend.

var module = angular.module("main",[]);
// This loads the google packages needed for charting. 
google.load("visualization", "1.0", {"packages":["corechart", "timeline"]});


// The basic structure can be described in several parts.

// There is a component which wraps up logic relating to Google Charts.
// It talks to the rest of the code through the Chart singleton.
// Each chart has a name, which is looked up in the Chart singleton to communicate.

// To perform a query through the UI, inject the `sendRequest` function.
// This function takes the query and other parameters (like profiling) as arguments.

// To capture the result of the request, inject `addResultListener` and call it with your function.
// The listening function will be called with the result when it is received.
// To get the current status of the request, you can inject and examine the RequestStatus singleton.

////////////////////////////////////////////////////////////////////////////////////////////////////

module.config(function($locationProvider) {
  $locationProvider.html5Mode(true);
});

////////////////////////////////////////////////////////////////////////////////////////////////////

// _state is for communicating between things

module.factory("_state", function() {
  return function() {
    var data = {};
    var listeners = {};
    var state = {
      update: function(name, fun) {
        if (!data[name]) {
          data[name] = {};
          listeners[name] = [];
        }
        fun(data[name]);
        for (var i = 0; i < listeners[name].length; i++) {
          listeners[name][i](data[name]);
        }
      },
      replace: function(name, fun) {
        if (!data[name]) {
          data[name] = {};
          listeners[name] = [];
        }
        data[name] = fun(data[name]);
        for (var i = 0; i < listeners[name].length; i++) {
          listeners[name][i](data[name]);
        }
      },
      set: function(name, key, value) {
        state.update(name, function(state) {
          data[name][key] = value;
        });
      },
      sets: function(name, map) {
        state.update(name, function(state) {
          for (var key in map) {
            data[name][key] = map[key];
          }
        });
      },
      addListener: function(name, fun) {
        if (!data[name]) {
          data[name] = {};
          listeners[name] = [];
        } else {
          fun(data[name]);
        }
        listeners[name].push(fun);
      }
    };
    return state;
  };
});

////////////////////////////////////////////////////////////////////////////////////////////////////

// The chart singleton stores data about charts:
// * options IN  (margins, things to hide, etc.)
// * data    IN  (the values to draw)
// * status  OUT (whether the chart has yet to draw, is rendering, or is rendered)
module.factory("_chart", function(_state) {
  return _state();
});

// The googleChart directive is used to create charts.
// It has a mandatory 'name' attribute attached.
module.directive("googleChart", function(_chart, $timeout) {
  return {
    restrict: "E",
    template: "<div></div>",
    scope: true,
    link: function(scope, element, attrs) {
      var chartElement = element[0];
      var name;
      attrs.$observe("chartName", function(value) {
        name = value;
        _chart.addListener(name, render);
      });
      var chart = null;
      function getUnits(value) {
        if (typeof value === "number") {
          return { value: value, units: "px" };
        }
        if (typeof value !== "string") {
          return null;
        }
        var result = value.match(/^([0-9.-]+)(%|px)$/);
        if (result === null) {
          return null;
        } else {
          return { value: parseFloat(result[1]), units: result[2] };
        }
      }
      function unitless(value) {
        return getUnits(value).value;
      }
      function fixUnits(value, total) {
        var match = getUnits(value);
        if (match == null) {
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
      function render(state) {
        if (chart !== null) {
          chart.clearChart();
          chart = null;
        }
        if (state.chartType === "line") {
          chart = new google.visualization.LineChart(chartElement);
        } else if (state.chartType === "area") {
          chart = new google.visualization.AreaChart(chartElement);
        } else if (state.chartType === "timeline") {
          chart = new google.visualization.Timeline(chartElement);
        } else {
          throw { message: "unknown chart type", chartType: state.chartType };
        }
        $timeout(function() {
          var data = state.data;
          var option = state.option;
          if (data && option) {
            _chart.set(name + "/waiting", "waiting", true);
            google.visualization.events.addListener(chart, "ready", function() { _chart.set(name + "/waiting", "waiting" , false); });
            var elementStyle = window.getComputedStyle(chartElement);
            var totalWidth = 800;//unitless(elementStyle.width) * 1;
            var totalHeight = 300;//unitless(elementStyle.height) * 1;
            option = deepCopy(option);
            option.isStacked = state.chartType == "area";
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
                left:   left,
                top:    top,
                width:  width,
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

////////////////////////////////////////////////////////////////////////////////////////////////////

// The _request singleton describes the state of requests that were sent up.
// * status: whether it's "up" or "down"
// * result: the value of the result (only meaningful when status = "down")
// * latency: the amount of time spent on the last query (only meaningful when status = "down")
module.factory("_request", function(_state) {
  return _state();
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

// The _launchRequest function provides a way to send up a query to the MQE server.
// When the request lands, it will be handled elsewhere.
module.factory("_launchRequest", function($http, _request) {
  var ticketBooth = new TicketBooth();
  return function(params) {    
    var start = new Date();
    var ticket = ticketBooth.next();
    _request.set("/query", "status", "up");
    var request = $http.get("/query", {
      params:params
    }).success(function(data, status, headers, config) {
      resolve(data);
    }).error(function(data, status, headers, config) {
      resolve(data);
    });
    function resolve(result) {
      if (!ticket.valid()) {
        return; // nothing left to do
      }
      var latency = new Date().getTime() - start.getTime();
      _request.sets("/query", {
        status: "down",
        latency: latency,
        result: result
      });
    }
  };
});

module.factory("_receiveListeners", function(_request) {
  var listeners = [];
  _request.addListener("/query", function(value) {
    for (var i = 0; i < listeners.length; i++) {
      if (!listeners[i].filter || listeners[i].filter(value)) {
        listeners[i].receive(value);
      }
    }
  });
  return listeners;
});

module.factory("_receiveSuccess", function(_receiveListeners) {
  return function(fun) {
    _receiveListeners.push({
      receive: function(value) {
        if (typeof fun != "function") {
          throw {message: "expected function, got", object: fun};
        }
        fun(value.result.name, value.result.body, value.latency);
      },
      filter: function(value) {
        return value.status === "down" && value.result && value.result.success;
      }
    });
  };
});

module.factory("_receiveProfile", function(_receiveListeners) {
  return function(fun) {
    _receiveListeners.push({
      receive: function(value) {
        if (typeof fun != "function") {
          throw {message: "expected function, got", object: fun};
        }
        fun(value.result.profile);
      },
      filter: function(value) {
        return value.status == "down" && !!value.result.profile;
      }
    });
  }
});

module.factory("_receiveFailure", function(_receiveListeners) {
  return function(fun) {
    _receiveListeners.push({
      receive: function(value) {
        fun(value.result.message);
      },
      filter: function(value) {
        return value.status === "down" && value.result && !value.result.success;
      }
    });
  };
});

////////////////////////////////////////////////////////////////////////////////////////////////////

var MAX_RENDERED = 200;

var seriesFormatters = [];

function formatTargetAxis(tagset, option) {
  if (tagset.$secondaxis === "true") {
    option.targetAxisIndex = 0;
  } else {
    option.targetAxisIndex = 1;
  }
}
function formatColor(tagset, option) {
  if (tagset.$color) {
    option.color = tagset.$color;
  }
}
function formatLineWidth(tagset, option) {
  if (tagset.$linewidth) {
    option.lineWidth = parseFloat(tagset.$linewidth);
  }
}
seriesFormatters.push(formatTargetAxis);
seriesFormatters.push(formatColor);
seriesFormatters.push(formatLineWidth);

function getFormattingOption(tagset) {
  var option = {};
  for (var i = 0; i < seriesFormatters.length; i++) {
    seriesFormatters[i](tagset, option);
  }
  return option;
}

function convertSelectResponse(object) {
  if (!(object &&
        object.length &&
        object[0].series &&
        object[0].series.length &&
        object[0].timerange)) {
    // invalid data.
    return null;
  }
  
  var labels = ["Time"]; // describes the labels for each component in the graph
  var tableArray = [labels]; // This array will become our table; the first row is labels
  var expressionCount = object.length;

  var seriesList = []; // The series to render (maximum of MAX_RENDERED of these)
  var optionList = []; // describes the per-series options (color, etc.)
  
  for (var i = 0; i < object.length; i++) {
    for (var j = 0; j < object[i].series.length; j++) {
      if (seriesList.length >= MAX_RENDERED) {
        break;
      }
      //
      var series = object[i].series[j];
      var option = getFormattingOption(series.tagset);
      //
      seriesList.push(series);
      optionList.push(option);
      labels.push(makeLabel(expressionCount === 1, object[i], series));
    }
  }

  // Take each of these series and add them to the table array.
  var timerange = object[0].timerange; // they all have the same timerange

  for (var t = 0; t < seriesList[0].values.length; t++) {
    var row = [dateFromIndex(t, timerange)]; // the time component
    for (var i = 0; i < seriesList.length; i++) {
      var cell = seriesList[i].values[t];
      if (cell === null) {
        row.push(NaN);
      } else {
        row.push(parseFloat(cell.toFixed(2)));
      }
    }
    tableArray.push(row);
  }

  return {
    dataTable: google.visualization.arrayToDataTable(tableArray),
    options: optionList
  };
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

function dateFromIndex(index, timerange) {
  return new Date(timerange.start + timerange.resolution * index);
}

function convertProfileResponse(profile) {
  if (!profile) {
    return null
  }
  var dataTable = new google.visualization.DataTable();
  dataTable.addColumn({ type: 'string', id: 'Name' });
  dataTable.addColumn({ type: 'number', id: 'Start' });
  dataTable.addColumn({ type: 'number', id: 'End' });
  var minValue = Number.POSITIVE_INFINITY;
  for (var i = 0; i < profile.length; i++) {
    minValue = Math.min(profile[i].start, minValue);
    minValue = Math.min(profile[i].finish, minValue);
  };
  function normalize(value) {
    return value - minValue;
  };
  for (var i = 0; i < profile.length; i++) {
    var row = [ profile[i].name , normalize(profile[i].start), normalize(profile[i].finish) ];
    dataTable.addRows([row]);
  }
  return dataTable;
}

