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
window.Autocom = (function(){
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

// scoreTable uses the table for memoization and computes a modified LCS on A and B.
// The config contains parameters for this procedure.
function scoreTable(table, A, B, i, j, config) {
	if (i >= A.length || j >= B.length) {
		return config.skipGiven * (i - A.length); // Don't penalize for the end of the second word
	}
	if (table[i][j] !== null) {
		return table[i][j];
	}
	if (A[i] == B[j]) {
		return table[i][j] = 1 / Math.sqrt(j+1) + scoreTable(table, A, B, i+1, j+1, config);
	}
	if (A[i].toLowerCase() === B[j].toLowerCase()) {
		return table[i][j] = 1 / Math.sqrt(j+1) * 0.5 + scoreTable(table, A, B, i+1, j+1, config);
	}

	// Skipping letters in the left are penalized (but penalized less if they're special characters).
	var advanceI = scoreTable(table, A, B, i+1, j, config) - config.skipGiven;
	var advanceJ = scoreTable(table, A, B, i, j+1, config) - config.skipWord;

	return table[i][j] = Math.max( advanceI, advanceJ );
}

// scoreAgainst computes how well the "letters" match "word" (this is not commutative).
// The config contains parameters for the algorithm.
function scoreAgainst(letters, word, config) {
	var table = [];
	for (var i = 0; i < letters.length; i++) {
		table[i] = [];
			for (var j = 0; j < word.length; j++) {
			table[i][j] = null;
		}
	}
	var s = scoreTable(table, letters, word, 0, 0, config);
	return s;
}

// generateElements creates the elements needed by an Autocom, based on the given input.
function generateElements(input) {
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

	return {
		input:   input,
		holder:  holder,
		hider:   hider,
		shadow:  shadow,
		prior:   prior,
		marker:  marker,
		after:   after,
		tooltip: tooltip
	};
}

// predictReady determines whether or not the input is ready for prediction,
// and it asks for something that matches non-separators.
function predictReady(input, letter) {
	if (input.selectionStart !== input.selectionEnd) {
		return false;
	}
	var at = input.selectionStart;
	if (input.value.substring(at).match("^" + letter)) {
		// Can't be going from the middle of a word.
		return false;
	}
	// Must be a word prior to the cursor.
	var before = input.value.substring(0, at).match("(" + letter + ")+$");
	if (before && before.length > 0) {
		return {from: at - before[0].length, word: before[0], to: at};
	}
	return false;
}

// filterCandidates finds a list of candidates, pursuant to the config, among the options
// from the given prefix `word`.
function filterCandidates(word, options, config) {
	// Compute the scores for each word.
	for (var i = 0; i < options.length; i++) {
		options[i] = { word: options[i], score: scoreAgainst(word, options[i], config) };
	}
	options.sort(function(a, b) {
		return b.score - a.score;
	});
	var words = [];
	for (var i = 0; i < options.length; i++) {
		if (options[i].score < config.threshold || (config.count !== null && i >= config.count)) {
			break;
		}
		words[i] = options[i].word;
	}
	if (words.length == 0) {
		return null;
	}
	if (words.length == 1 && words[0].word == word) {
		return null;
	}
	return words;
}

// Predicts the possible autocompletions based on `at` which includes the word we want, and our options.
function predict(at, options, config) {
	if (!at) {
		return null;
	}
	var words = filterCandidates(at.word, options, config);
	if (!words) {
		return null;
	}
	return {at: at, words: words};
}

// Moves the tooltip to its location proper.
function moveTooltip(elements, offsetX, offsetY) {
	// Resize the shadower:
	elements.shadow.style.width = getComputedStyle(elements.input).width;
	// Fill prior and after with text:
	elements.prior.innerHTML = elements.input.value.substring(0, elements.input.selectionStart);
	elements.after.innerHTML = elements.input.value.substring(elements.input.selectionStart);
	// Use marker's location to reposition tooltip:
	elements.tooltip.style.left = ((elements.input.offsetLeft + elements.marker.offsetLeft + offsetX)|0) + "px";
	elements.tooltip.style.top = ((elements.input.offsetTop + elements.marker.offsetTop - elements.input.scrollTop + offsetY)|0) + "px";
}

// Creates and adds to the tooltip one row, with the given word/index.
// If clicked, it invokes the given selectedCallback.
function generateTooltipRow(tooltip, word, index, selectedIndex, selectedCallback) {
	var row = document.createElement("div");
	row.className = "autocom-tooltip-line";
	row.addEventListener("mousedown", function() {
		selectedCallback(index);
	});
	row.appendChild(document.createTextNode(word));
	if (index == selectedIndex) {
		row.className += " autocom-tooltip-selected";
	}
	tooltip.appendChild(row);
}

// Creates and adds all rows to the given tooltip. Each row invokes the given callback when clicked.
function generateTooltipContents(tooltip, words, index, selectedCallback) {
	// First, empty the old tooltip.
	while (tooltip.firstChild) {
		tooltip.removeChild(tooltip.firstChild);
	}
	// Then, fill each row.
	for (var i = 0; i < words.length; i++) {
		generateTooltipRow(tooltip, words[i], i, index, selectedCallback);
	}
}

// Inserts `word` at the location specified by `at` inside of `input`.
function insertWord(input, at, word) {
	input.value = input.value.substring(0, at.from) + word + input.value.substring(at.to);
	input.selectionStart = input.selectionEnd = at.from + word.length;
}


// Creates an autocom for the element.

function Autocom(input) {
	var self = this;
	// The element which holds our input should be a relative-div (ideally).
	var elements = generateElements(input);
	self.elements = elements;

	self.options = [];

	self.config = {
		threshold: 0,
		skipGiven: 2,
		skipWord: 0.25,
		count: 8
	};

	var tooltipState = {active: false, index: 0};
	var tooltipSuppress = false;

	function keyPress(e) {
		if (tooltipState.active && !tooltipSuppress) {
			if (e.keyCode == 9 || e.keyCode == 13) { // TAB or ENTER
				// Tab
				e.preventDefault();
				insertWord(input, tooltipState.at, tooltipState.words[tooltipState.index]);
				refresh();
				tooltipSuppress = true;
				return;
			}
			if (e.keyCode == 38) { // UP
				e.preventDefault();
				tooltipState.index--;
				if (tooltipState.index < 0) {
					tooltipState.index = tooltipState.words.length-1;
				}
				return;
			}
			if (e.keyCode == 40) { // DOWN
				e.preventDefault();
				tooltipState.index++;
				if (tooltipState.index >= tooltipState.words.length) {
					tooltipState.index = 0;
				}
				return;
			}
			if (e.keyCode == 27) {
				tooltipSuppress = true;
				return;
			}
		}
		tooltipSuppress = false;
		refresh();
	}
	input.addEventListener("keydown", keyPress, false);

	function completeSelect(index) {
		insertWord(input, tooltipState.at, tooltipState.words[index]);
		refresh();
	}

	function renderTooltip() {
		moveTooltip(elements, self.tooltipX, self.tooltipY);

		var result = predict(predictReady(input, self.letter), self.options.slice(0), self.config);
		if (result && !tooltipSuppress && document.activeElement === input) {
			// If it's not currently active, then become active.
			tooltipState = {
				active: true,
				index: tooltipState.active ? tooltipState.index : 0,
				words: result.words,
				at: result.at,
			};
			if (tooltipState.index >= tooltipState.words.length) {
				tooltipState.index = 0;
			}
			generateTooltipContents(elements.tooltip, tooltipState.words, tooltipState.index, completeSelect);
			
			if (elements.marker.offsetLeft + elements.tooltip.offsetWidth > input.offsetWidth) {
				tooltip.style.left = ((input.offsetLeft + input.offsetWidth - elements.tooltip.offsetWidth)|0) + "px";
			}
		} else {
			tooltipState.active = false;
		}
		elements.tooltip.hidden = !tooltipState.active || tooltipSuppress;
	}
	var refresh = function() {
		setTimeout(renderTooltip, 0);
		var inputStyle = getComputedStyle(input);

		// Make input have style that mtches the holder.
		// (prior/post must match exactly too).
		var textProperties = ["fontSize", "fontFamily", "lineHeight", "color", "fontWeight"];
		for (var i = 0; i < textProperties.length; i++) {
			var property = textProperties[i];
			elements.prior.style[property] = inputStyle[property];
			elements.after.style[property] = inputStyle[property];
		}

		var boxProperties = ["padding-left", "padding-right", "padding-top", "padding-bottom", "width"];
		for (var i = 0; i < boxProperties.length; i++) {
			var property = boxProperties[i];
			elements.shadow.style[property] = inputStyle[property];
		}

		// Set shadow position:
		elements.shadow.style.position = "relative";
		elements.shadow.style.left = "0";
		elements.shadow.style.top = "0";
		elements.shadow.style.margin = "0";
		elements.shadow.style.border = "0";
		elements.shadow.style.height = "0"; // 0 height so that it is not visibile.
		elements.shadow.style.overflowY = "hidden";

		input.style.whiteSpace = "pre-wrap";
		elements.prior.style.whiteSpace = input.style.whiteSpace;
		elements.after.style.whiteSpace = input.style.whiteSpace;

		input.style.wordWrap = "normal";
		elements.prior.style.wordWrap = input.style.wordWrap;
		elements.after.style.wordWrap = input.style.wordWrap;

		// Nowrap on the marker
		elements.marker.style.whiteSpace = "nowrap";

		// Input size
		input.style.height = "100%";
		input.style.overflowX = "hidden";

		// Tooltip
		elements.tooltip.style.position = "absolute";
		elements.tooltip.style.zIndex = 1000;
	};
	input.addEventListener("input", refresh);
	input.addEventListener("mousedown", refresh);
	input.addEventListener("mouseup", refresh);
	input.addEventListener("keyup", refresh);
	input.addEventListener("keydown", refresh);
	input.addEventListener("resize", refresh);
	input.addEventListener("blur", refresh); // blur is the "unfocus" event
	refresh();

	self.refresh = refresh;
}

Autocom.prototype.tooltipX = 0;
Autocom.prototype.tooltipY = 0;
Autocom.prototype.letter = "[A-Za-z]";
return Autocom;
})();
