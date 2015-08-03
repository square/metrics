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

var CLASS_TOOLTIP_LINE = "autocom-tooltip-line";
var CLASS_TOOLTIP_SELECTED = "autocom-tooltip-selected";
var CLASS_TOOLTIP_BOX = "autocom-tooltip-box";

var TOOLTIP_ZINDEX = 1000;

window.Autocom = (function(){
"use strict";

// Creates an Autocompleting textarea within the given HTML element.
// The element should be relative/absolute/fixed position to render correctly.
// CSS classes can select the following elements:
//   * autocom-tooltip-line       a single line within the tooltip (also stored as a JS variable CLASS_TOOLTIP_LINE)
//   * autocom-tooltip-selected   the selected line within the tooltip (also stored as a JS variable CLASS_TOOLTIP_SELECTED)
//   * autocom-tooltip-box        the box which contains the lines for the tooltip (also stored as a JS variable CLASS_TOOLTIP_BOX)

// Example JS:
/*
	// Select a textarea
	var input = document.getElementById("color_entry");

	// Create an Autocom object
	var colorCom = new Autocom(input);

	// Assign the options for our Autocom
	colorCom.options = ["red", "orange", "yellow", "yellow-green", "green", "cyan", "blue", "purple"];
	
	// Words are made of lowercase letters and hyphens, but can't start with a hyphen
	colorCom.prefixPattern = "[a-z][a-z\\-]+";
	// The continue pattern is checked against, but its actual value is discard. Matching one letter is enough:
	colorCom.prefixContinue = "[a-z\\-];"

	// Adjust the tooltip position offset
	colorCom.tooltipX = 25;
	colorCom.tooltipY = 0;

	// Configure the number of shown options  
	colorCom.config.count = 3;
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
	
	prefixPattern: (default "[a-zA-Z]+") the regex to match the prefix of a valid word
	prefixContinue: (default "[a-zA-Z]+") the regex to match the "middle" of a valid word

	config:
		threshold:        (default 0)    // Autocomplete score threshold required to show (for fuzzy suggestions)
		skipGiven:        (default 2)    // Penalty for skipping a letter in the input
		skipSpecialGiven: (default 0.25) // Penalty for skipping a special character in the input ("non-letter" having same upper- and lower-case form)
		skipWord:         (default 0.25) // Penalty for skipping a prefixed letter in the suggestion
		skipWordEnd:      (default 0)    // Penalty for skipping characters at the end of the candidate word (for most applications, it should be much smaller than skipWord)
		count:            (default 8)    // The maximum number of autocomplete suggestions shown 

	tooltipX: (default 0) the X offset for the tooltip (relative to the cursor)
	tooltipY: (default 0) the Y offset for the tooltip (relative to the cursor)

	hook: a callback which is called whenever an insertion is made

Methods:

	refresh(): redraw the tooltip (or hide it) taking into account any changes to focus or fields.
	         (this rarely needs to be called)

*/

function isLetter(c) {
	return c.toLowerCase() != c.toUpperCase();
}

// scoreTable uses the table for memoization and computes a modified LCS on A and B.
// The config contains parameters for this procedure.
function scoreTable(table, A, B, i, j, config) {
	if (i >= A.length || j >= B.length) {
		return config.skipGiven * (i - A.length) + config.skipWordEnd * (j - B.length);
	}
	if (table[i][j] !== null) {
		return table[i][j];
	}
	if (A[i] === B[j]) {
		return table[i][j] = 1 / Math.sqrt(j+1) + scoreTable(table, A, B, i+1, j+1, config);
	}
	if (A[i].toLowerCase() === B[j].toLowerCase()) {
		return table[i][j] = 1 / Math.sqrt(j+1) * 0.5 + scoreTable(table, A, B, i+1, j+1, config);
	}

	// Skipping letters in the left are penalized (but penalized less if they're special characters).
	var advanceI = scoreTable(table, A, B, i+1, j, config) - (isLetter(A[i]) ? config.skipGiven : config.skipSpecialGiven);
	var advanceJ = scoreTable(table, A, B, i, j+1, config) - config.skipWord;

	return table[i][j] = Math.max( advanceI, advanceJ );
}


function serializeConfig(config) {
	var configs = [];
	for (var i in config) {
		configs.push(i + ": " + config[i]);
	}
	configs.sort();
	return configs.join(", ");
}

var scoreAgainstCache = {};

// scoreAgainst computes how well the "letters" match "word" (this is not commutative).
// The config contains parameters for the algorithm.
function scoreAgainst(letters, word, config) {
	var cacheIndex = letters + "||" + word + "||" + serializeConfig(config);
	if (scoreAgainstCache[cacheIndex]) {
		return scoreAgainstCache[cacheIndex];
	}
	var table = [];
	for (var i = 0; i < letters.length; i++) {
		table[i] = [];
			for (var j = 0; j < word.length; j++) {
			table[i][j] = null;
		}
	}
	var s = scoreTable(table, letters, word, 0, 0, config);
	return scoreAgainstCache[cacheIndex] = s;
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
	tooltip.className = CLASS_TOOLTIP_BOX;
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
function predictReady(input, prefixPattern, continuePattern) {
	if (input.selectionStart !== input.selectionEnd) {
		return null;
	}
	var at = input.selectionStart;
	if (input.value.substring(at).match("^(" + continuePattern + ")")) {
		// Can't be going from the middle of a word.
		return null;
	}
	// Must be a word prior to the cursor.
	var before = input.value.substring(0, at).match("(" + prefixPattern + ")$");
	if (before && before.length > 0) {
		return {from: at - before[0].length, word: before[0], to: at};
	}
	return null;
}

// filterCandidates finds a list of candidates, pursuant to the config, among the options
// from the given prefix `word`.
function filterCandidates(word, givenOptions, config) {
	// Compute the scores for each word.
	var options = [];
	for (var i = 0; i < givenOptions.length; i++) {
		options[i] = { word: givenOptions[i], score: scoreAgainst(word, givenOptions[i], config) };
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
	if (words.length === 1 && words[0] === word) {
		return [];
	}
	return words;
}

// Predicts the possible autocompletions based on `at` which includes the word we want, and our options.
function predict(at, options, config) {
	if (!at) {
		return null;
	}
	var words = filterCandidates(at.word, options, config);
	if (words.length === 0) {
		return null;
	}
	return {at: at, words: words};
}

// Moves the tooltip to its location proper.
function moveTooltip(elements, offsetX, offsetY) {
	// Resize the shadower:
	elements.shadow.style.width = getComputedStyle(elements.input).width;
	// Fill prior and after with text:
	elements.prior.textContent = elements.input.value.substring(0, elements.input.selectionStart);
	elements.after.textContent = elements.input.value.substring(elements.input.selectionStart);
	// Use marker's location to reposition tooltip:
	elements.tooltip.style.left = Math.floor(elements.input.offsetLeft + elements.marker.offsetLeft + offsetX) + "px";
	elements.tooltip.style.top = Math.floor(elements.input.offsetTop + elements.marker.offsetTop - elements.input.scrollTop + offsetY) + "px";
}

// Creates and adds to the tooltip one row, with the given word/index.
// If clicked, it invokes the given selectedCallback.
function generateTooltipRow(tooltip, word, index, selectedIndex, selectedCallback) {
	var row = document.createElement("div");
	row.className = CLASS_TOOLTIP_LINE;
	row.addEventListener("mousedown", function() {
		selectedCallback(index);
	});
	row.appendChild(document.createTextNode(word));
	if (index == selectedIndex) {
		row.className += " " + CLASS_TOOLTIP_SELECTED;
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
function insertWord(input, at, word, supressCallback, hook) {
	input.value = input.value.substring(0, at.from) + word + input.value.substring(at.to);
	setTimeout(function() {
		input.focus();
		input.selectionStart = input.selectionEnd = at.from + word.length;
		supressCallback();
		if (hook) {
			hook();
		}
	}, 1);
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
		skipSpecialGiven: 0.25,
		skipWord: 0.25,
		skipWordEnd: 0,
		count: 8
	};

	var tooltipState = {active: false, index: 0};
	var tooltipSuppress = false;

	function supressCallback() {
		tooltipSuppress = true;
	}

	function keyPress(e) {
		if (tooltipState.active && !tooltipSuppress) {
			if (e.keyCode == 9 || e.keyCode == 13) { // TAB or ENTER
				// Tab
				e.preventDefault();
				insertWord(input, tooltipState.at, tooltipState.words[tooltipState.index], supressCallback, self.hook);
				tooltipSuppress = true;
				refresh();
				return;
			} else if (e.keyCode == 38 && !e.shiftKey) { // UP
				e.preventDefault();
				tooltipState.index--;
				if (tooltipState.index < 0) {
					tooltipState.index = tooltipState.words.length-1;
				}
				return;
			} else if (e.keyCode == 40 && !e.shiftKey) { // DOWN
				e.preventDefault();
				tooltipState.index++;
				if (tooltipState.index >= tooltipState.words.length) {
					tooltipState.index = 0;
				}
				return;
			} else if (e.keyCode == 27) { // ESC
				tooltipSuppress = true;
				return;
			}
		}

		if (
			e.keyCode == 16 ||
			e.getModifierState("Shift")
		) {
			// shift is pressed
			tooltipSuppress = true;
		} else {
			tooltipSuppress = false; // Start showing the tooltip again.
			refresh();
		}
	}
	input.addEventListener("keydown", keyPress, false);

	function completeSelect(index) {
		insertWord(input, tooltipState.at, tooltipState.words[index], supressCallback, self.hook);
		refresh();
	}
	function renderTooltip() {
		moveTooltip(elements, self.tooltipX, self.tooltipY);
		var result = predict(predictReady(input, self.prefixPattern, self.continuePattern), self.options.slice(0), self.config);
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
				elements.tooltip.style.left = Math.floor(input.offsetLeft + input.offsetWidth - elements.tooltip.offsetWidth) + "px";
			}
		} else {
			tooltipState.active = false;
		}
		elements.tooltip.hidden = !tooltipState.active || tooltipSuppress;
	}
	var refresh = function() {
		setTimeout(renderTooltip, 0);
		var inputStyle = getComputedStyle(input);

		// Make input have style that matches the holder.
		// (prior/post must match exactly too).
		var textProperties = ["fontSize", "fontFamily", "lineHeight", "color", "fontWeight", "whiteSpace", "wordWrap"];
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

		// Nowrap on the marker
		elements.marker.style.whiteSpace = "nowrap";

		// Tooltip
		elements.tooltip.style.position = "absolute";
		elements.tooltip.style.zIndex = TOOLTIP_ZINDEX;
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
Autocom.prototype.prefixPattern = "[A-Za-z]+";
return Autocom;
})();
