'use strict';

// NFR-021 AC1–AC3: in-repo source-contract audit of the offline issue UI.
// This is not a WCAG engine, browser AT run, or axe-core scan.

var assert = require('assert');
var fs = require('fs');
var path = require('path');

var html = fs.readFileSync(path.join(__dirname, 'page.html'), 'utf8');
var js = fs.readFileSync(path.join(__dirname, 'page.js'), 'utf8');
var css = fs.readFileSync(path.join(__dirname, 'page.css'), 'utf8');

function contains(src, needle, label) {
  assert.ok(src.indexOf(needle) !== -1, (label || needle) + ' missing');
}

// AC1: keyboard-operable native controls and labels.
contains(html, 'role="tablist"', 'tablist');
contains(html, 'aria-orientation="horizontal"', 'tab orientation');
contains(html, 'id="tab-timeline"', 'timeline tab');
contains(html, 'id="tab-issues"', 'issues tab');
contains(html, 'aria-controls="view-issues"', 'issues tab controls');
contains(html, '<label>Search issues', 'search issues label');
contains(html, '<label>State', 'state filter label');
contains(html, '<label>Flag', 'flag filter label');
contains(html, 'Overdue only', 'overdue filter label');
contains(html, 'issue-workflow-hint', 'keyboard hint');
contains(html, 'Enter adds a tag', 'Enter-to-tag hint');
contains(js, "e.key === 'Enter'", 'tag Enter handler');
contains(js, "e.key === 'ArrowRight'", 'tab ArrowRight');
contains(js, "e.key === 'ArrowLeft'", 'tab ArrowLeft');
contains(js, "function rememberFocus(", 'rememberFocus');
contains(js, "function applyPendingFocus(", 'applyPendingFocus');
contains(js, "rememberFocus(issue.id, 'flag')", 'flag restores focus');
contains(js, "rememberFocus(issue.id, 'state')", 'state restores focus');
contains(js, "rememberFocus(issue.id, 'tag-input')", 'tag restores focus');
contains(js, "data-control', 'flag'", 'flag data-control');
contains(js, "data-control', 'state'", 'state data-control');
contains(js, "data-control', 'tag-input'", 'tag-input data-control');
contains(js, "el('label', '', 'New tag')", 'New tag label');
contains(js, "el('button', '', existing ? 'Open issue' : 'Create issue')", 'create/open button');
contains(js, "btn.type = 'button'", 'create button type');
contains(js, "flagBtn.type = 'button'", 'flag button type');
contains(js, "addBtn.type = 'button'", 'add-tag button type');
contains(css, 'button:focus-visible', 'focus-visible buttons');
contains(css, 'input:focus-visible', 'focus-visible inputs');
contains(css, 'select:focus-visible', 'focus-visible selects');

// AC2: flag and state are named in text, not color alone.
contains(js, "el('span', 'flag-badge', 'Flagged')", 'Flagged text badge');
contains(js, "'State: ' + stateLabel(issue.state)", 'State text badge');
contains(js, "issue.flagged ? 'Unflag' : 'Flag for attention'", 'flag button text');
contains(js, "aria-pressed", 'flag aria-pressed');
contains(js, "aria-label', 'Flag: flagged'", 'flag badge name');
contains(js, "'Workflow state: ' + stateLabel(issue.state)", 'state badge name');
contains(html, 'Flagged only', 'flag filter option text');
contains(css, '.flag-badge', 'flag badge style');
contains(css, '.state-badge', 'state badge style');

// AC3: immediate visible confirmation.
contains(html, 'id="issue-feedback"', 'issue status region');
contains(html, 'aria-live="polite"', 'polite live region');
contains(html, 'aria-atomic="true"', 'atomic status');
contains(html, 'id="issue-summary"', 'visible issue count');
contains(js, "'Showing ' + visible.length + ' of ' + total + ' issue(s).'", 'filter count summary');
contains(js, "announceCount: true", 'filter announces count');
contains(js, "(next ? 'Flagged ' : 'Unflagged ') + issue.id", 'flag confirmation');
contains(js, "'State for ' + issue.id + ' is now '", 'state confirmation');
contains(js, "result.added ? 'Tagged ' + issue.id", 'tag confirmation');
contains(js, "result.created ? 'Created ' + result.issue.id", 'create confirmation');
contains(html, 'role="status"', 'status role');

function extractFunction(src, name) {
  var start = src.indexOf('function ' + name + '(');
  assert.ok(start !== -1, name + ' missing');
  var depth = 0;
  var brace = src.indexOf('{', start);
  for (var i = brace; i < src.length; i++) {
    if (src[i] === '{') depth++;
    else if (src[i] === '}') {
      depth--;
      if (depth === 0) return src.slice(start, i + 1);
    }
  }
  throw new Error('unclosed ' + name);
}

// Confirmations must cancel the delayed detail write so it cannot overwrite them.
var showFb = extractFunction(js, 'showFeedback');
assert.ok(showFb.indexOf('detailFeedbackTimer') !== -1, 'showFeedback cancels detailFeedbackTimer');
assert.ok(showFb.indexOf('clearTimeout') !== -1, 'showFeedback clearTimeout');

// Empty-list early returns must apply/clear pendingFocus (do not steal focus later).
var renderIss = extractFunction(js, 'renderIssues');
assert.ok((renderIss.match(/applyPendingFocus\(\)/g) || []).length >= 3,
  'renderIssues applies pendingFocus on empty early returns and the populated path');

// XSS / offline discipline kept while hardening the workflow.
assert.strictEqual(js.indexOf('innerHTML'), -1, 'page.js must not assign innerHTML');
assert.ok(!/https?:\/\//.test(html.replace(/\{\{\.Data\}\}/g, '')), 'page.html has no network URL');
assert.ok(!/cdn|unpkg|jsdelivr/i.test(html + js + css), 'no CDN references');

console.log('nfr021_a11y_test.js: ok');
