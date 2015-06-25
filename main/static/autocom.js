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

// Creates an Autocompleting textarea within the given HTML element.
// The element should be relative/absolute/fixed position to render correctly.
// CSS classes can select the following elements:
//   * autocom-tooltip-line       a single line within the tooltip
//   * autocom-tooltip-selected   the selected line within the tooltip
//   * autocom-tooltip-box        the box which contains the lines for the tooltip

// Example JS:
/*
	// Select the div you want to insert your textarea into:
	var inputDiv = document.getElementById("color_entry");

	// Create an Autocom object (it creates the textarea + other elements for you)
	var colorCom = new Autocom(inputDiv);

	// Assign the options for our Autocom
	colorCom.options = ["red", "orange", "yellow", "yellow-green", "green", "cyan", "blue", "purple"];
	
	// Add the hyphen to the possible set of letters:
	colorCom.letters = "[a-z\\-]";

	// Adjust the tooltip position offset:
	colorCom.tooltipX = 25;
	colorCom.tooltipY = 0;

	// Use the input:
	function submit() {
		var text = colorCom.getValue();

		// ...
	}
*/

/*
Fields:

	elements: a set of the HTML elements associated with the Autocom form.
		input: the input textarea
		shadow: an invisible (0-height) div used to calculate the marker location
		prior: a span inside shadow used to calculate the marker position
		marker: a span inside shadow used to calculate cursor position
		after: a span inside shadow used to calculate marker position
		holder: the element originally passed as an argument which holds all of these
		tooltip: the tooltip elements which shows the suggests
	
	options: (default []) an array of strings; these are the options which autocomplete suggests
	
	letter: (default "[a-zA-Z]") the regex to match a single non-separator

	threshold: (default 0)    // Autocomplete score threshold required to show (for fuzzy suggestions)
	skipGiven: (default 2)    // Penalty for skipping a letter in the input
	skipWord:  (default 0.25) // Pentalty for skipping a prefixed letter in the suggestion

	tooltipX: (default 0) the X offset for the tooltip (relative to the cursor)
	tooltipY: (default 0) the Y offset for the tooltip (relative to the cursor)

Methods:

	refresh(): redraw the tooltip (or hide it) taking into account any changes to focus or fields.
	         (this rarely needs to be called)
	
	restyle(): assigns the styles of the input to match the holder. Call it when the holder's font changes.

	setOptions(options): assigns the options and calls refresh()

*/

function Autocom(input) {
	// The element which holds our input should be a relative-div (ideally).
	var holder = input.parentElement;

	var hider = document.createElement("div");
	// Hider completely hides its contents
	hider.style.width = "0";
	hider.style.height = "0";
	hider.style.margin = "0";
	hider.style.padding = "0";
	hider.style.outline = "0";
	hider.style.border = "0";
	hider.style.overflowX = "hidden";
	hider.style.overflowY = "hidden";
	holder.insertBefore(hider, input);

	var shadow = document.createElement("div");
	// Flow can be made possible simpler/more predictable if we insert before our input.
	hider.appendChild(shadow);

	var prior = document.createElement("span");
	shadow.appendChild(prior);

	var marker = document.createElement("span");
	shadow.appendChild(marker);

	var after = document.createElement("span");
	shadow.appendChild(after);

	var tooltip = document.createElement("div");
	tooltip.className = "autocom-tooltip-box";
	holder.appendChild(tooltip);

	var self = this;

	// Remember these as fields.
	this.elements = {
		input: input,
		holder: holder,
		prior: prior,
		after: after,
		marker: marker,
		tooltip: tooltip,
		shadow: shadow
	};

	// Style them.
	this.restyle();

	this.options = [];

	function scoreTable(table, A, B, i, j) {
		if (i >= A.length || j >= B.length) {
			return self.skipGiven * (i - A.length); // Don't penalize for the end of the second word
		}
		if (table[i][j] !== -1) {
			return table[i][j];
		}
		if (A[i] == B[j]) {
			return table[i][j] = 1 / Math.sqrt(j+1) + scoreTable(table, A, B, i+1, j+1);
		}
		if (A[i].toLowerCase() === B[j].toLowerCase()) {
			return table[i][j] = 1 / Math.sqrt(j+1) * 0.5 + scoreTable(table, A, B, i+1, j+1);
		}

		// Skipping letters in the left are penalized (but penalized less if they're special characters).
		var advanceI = scoreTable(table, A, B, i+1, j) - self.skipGiven;
		var advanceJ = scoreTable(table, A, B, i, j+1) - self.skipWord;

		return table[i][j] = Math.max( advanceI, advanceJ );
	}

	function scoreAgainst(letters, word) {
		var table = [];
		for (var i = 0; i < letters.length; i++) {
			table[i] = [];
			for (var j = 0; j < word.length; j++) {
				table[i][j] = -1;
			}
		}
		var s = scoreTable(table, letters, word, 0, 0);
		return s;
	}

	function predictReady() {
		if (input.selectionStart !== input.selectionEnd) {
			return false;
		}
		var at = input.selectionStart;
		if (input.value.substring(at).match("^" + self.letter)) {
			// Can't be going from the middle of a word.
			return false;
		}
		// Must be a word prior to the cursor.
		var before = input.value.substring(0, at).match("(" + self.letter + ")+$");
		if (before && before.length > 0) {
			return {from: at - before[0].length, word: before[0], to: at};
		}
		return false;
	}

	function predict() {
		var at = predictReady();
		if (!at) {
			return null;
		}

		var candidates = self.options.slice(0);
		for (var i = 0; i < candidates.length; i++) {
			candidates[i] = { word: candidates[i], score: scoreAgainst(at.word, candidates[i]) };
		}
		candidates.sort(function(a, b) {
			return b.score - a.score;
		});

		var words = [];
		for (var i = 0; i < candidates.length; i++) {
			if (candidates[i].score < self.threshold) {
				break;
			}
			words[i] = candidates[i];
		}
		if (words.length == 0) {
			return null;
		}
		if (words.length == 1 && words[0].word == at.word) {
			return null;
		}
		return {words: words, at: at};
	}

	var tooltipActive = false;
	var tooltipSuppress = false;

	function insertWord(active) {
		input.value = input.value.substring(0, active.at.from) + active.words[active.index].word + input.value.substring(active.at.to);
		input.selectionStart = input.selectionEnd = active.at.from + active.words[active.index].word.length;
		onInputChange();
	}

	function keyPress(e) {
		if (tooltipActive && !tooltipSuppress) {
			if (e.keyCode == 9 || e.keyCode == 13) { // TAB or ENTER
				// Tab
				e.preventDefault();
				insertWord(tooltipActive);
				tooltipSuppress = true;
				return;
			}
			if (e.keyCode == 38) { // UP
				e.preventDefault();
				tooltipActive.index--;
				if (tooltipActive.index < 0) {
					tooltipActive.index = tooltipActive.words.length-1;
				}
				return;
			}
			if (e.keyCode == 40) { // DOWN
				e.preventDefault();
				tooltipActive.index++;
				if (tooltipActive.index >= tooltipActive.words.length) {
					tooltipActive.index = 0;
				}
				return;
			}
			if (e.keyCode == 27) {
				tooltipSuppress = true;
				return;
			}
		}
		tooltipSuppress = false;
		onInputChange();
	}
	input.addEventListener("keydown", keyPress, false);

	function renderTooltip() {
		shadow.style.width = getComputedStyle(input).width;

		prior.innerHTML = input.value.substring(0, input.selectionStart);
		after.innerHTML = input.value.substring(input.selectionStart);
		tooltip.style.left = ((input.offsetLeft + marker.offsetLeft + self.tooltipX)|0) + "px";
		tooltip.style.top = ((input.offsetTop + marker.offsetTop - input.scrollTop + self.tooltipY)|0) + "px";
		var result = predict();
		if (result && !tooltipSuppress && document.activeElement === input) {
			tooltipActive = tooltipActive || {index: 0};
			tooltipActive.words = result.words;
			tooltipActive.at = result.at;
			if (tooltipActive.index >= tooltipActive.words.length) {
				tooltipActive.index = 0;
			}
			while (tooltip.firstChild) {
				tooltip.removeChild(tooltip.firstChild);
			}
			for (var i = 0; i < result.words.length; i++) {
				var line = document.createElement("div");
				line.className = "autocom-tooltip-line";
				(function(index) {
					line.addEventListener("mousedown", function() {
						tooltipActive.index = index;
						insertWord(tooltipActive);
					});
				})(i);
				line.appendChild(document.createTextNode(result.words[i].word));
				tooltip.appendChild(line);
				if (i == tooltipActive.index) {
					line.className += " autocom-tooltip-selected";
				}
			}
			if (marker.offsetLeft + tooltip.offsetWidth > input.offsetWidth) {
				tooltip.style.left = ((input.offsetLeft + input.offsetWidth - tooltip.offsetWidth)|0) + "px";
			}
		} else {
			tooltipActive = false;
		}
		tooltip.hidden = !tooltipActive || tooltipSuppress;
	}

	// Whenever a change occurs, wait (which allows the interaction, like moving the cursor or entering text) to complete.
	// Then render the tooltip.
	function onInputChange() {
		self.restyle();
		setTimeout(renderTooltip, 0);
	}
	this.refresh = onInputChange;
	input.addEventListener("input", onInputChange);
	input.addEventListener("mousedown", onInputChange);
	input.addEventListener("mouseup", onInputChange);
	input.addEventListener("keyup", onInputChange);
	input.addEventListener("keydown", onInputChange);
	input.addEventListener("resize", onInputChange);
	input.addEventListener("blur", onInputChange); // blur is the "unfocus" event
}
Autocom.prototype.restyle = function() {
	var inputStyle = getComputedStyle(this.elements.input);

	// Make input have style that mtches the holder.
	// (prior/post must match exactly too).
	var textProperties = ["fontSize", "fontFamily", "lineHeight", "color", "fontWeight"];
	for (var i = 0; i < textProperties.length; i++) {
		var property = textProperties[i];
		this.elements.prior.style[property] = inputStyle[property];
		this.elements.after.style[property] = inputStyle[property];
	}

	var boxProperties = ["padding-left", "padding-right", "padding-top", "padding-bottom", "width"];
	for (var i = 0; i < boxProperties.length; i++) {
		var property = boxProperties[i];
		this.elements.shadow.style[property] = inputStyle[property];
	}

	// Set shadow position:
	this.elements.shadow.style.position = "relative";
	this.elements.shadow.style.left = "0";
	this.elements.shadow.style.top = "0";
	this.elements.shadow.style.margin = "0";
	this.elements.shadow.style.border = "0";
	this.elements.shadow.style.height = "0"; // 0 height so that it is not visibile.
	this.elements.shadow.style.overflowY = "hidden";

	this.elements.input.style.whiteSpace = "pre-wrap";
	this.elements.prior.style.whiteSpace = this.elements.input.style.whiteSpace;
	this.elements.after.style.whiteSpace = this.elements.input.style.whiteSpace;

	this.elements.input.style.wordWrap = "normal";
	this.elements.prior.style.wordWrap = this.elements.input.style.wordWrap;
	this.elements.after.style.wordWrap = this.elements.input.style.wordWrap;

	// Nowrap on the marker
	this.elements.marker.style.whiteSpace = "nowrap";

	// Input size
	this.elements.input.style.height = "100%";
	this.elements.input.style.overflowX = "hidden";

	// Tooltip
	this.elements.tooltip.style.position = "absolute";
	this.elements.tooltip.style.zIndex = 1000;
};
// Sets the options which will be autocompleted.
// This also triggers a redraw of the tooltip suggests.
Autocom.prototype.setOptions = function(options) {
	this.options = options;
	this.refresh();
};
Autocom.prototype.tooltipX = 0;
Autocom.prototype.tooltipY = 35;
Autocom.prototype.letter = "[A-Za-z]";
// These parameters are used by the fuzzy autocomplete system.
Autocom.prototype.threshold = 0;
Autocom.prototype.skipGiven = 2;
Autocom.prototype.skipWord = 0.25;
