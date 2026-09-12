'use strict';

// NFR-013: in-repo source-contract + generated-chrome audit.
// Stdlib Go contrast/label checks in a11y.go cover the written HTML.
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

// AC1: explicit labels; placeholders are examples, not the only name.
[
  'for="q"', 'for="sev"', 'for="inst"', 'for="src"',
  'for="iq"', 'for="istate"', 'for="itags"', 'for="iflag"',
  'for="iowner"', 'for="isev"', 'for="iinst"', 'for="ioverdue"'
].forEach(function (attr) {
  contains(html, attr, attr);
});
contains(html, '<label for="q">Search</label>', 'search label text');
contains(html, '<label for="sev">Severity</label>', 'severity label text');
contains(html, '<label for="ioverdue">Overdue only</label>', 'overdue label text');
contains(html, 'aria-label="Follow-up JSON file"', 'hidden file input name');
contains(js, "function labeledField(", 'labeledField helper');
contains(js, "lab.setAttribute('for', control.id)", 'dynamic for= binding');
contains(js, "labeledField('New tag', tagInput)", 'New tag labeledField');
contains(js, "dueLabel.setAttribute('for', dueInput.id)", 'due date for=');
contains(js, "labeledField('Issue to link', linkSelect)", 'link picker label');
assert.ok(js.indexOf("dueLabel.appendChild(dueRow)") === -1, 'due label must not wrap the clear button');

// AC2: severity is named in text, not color alone.
contains(js, "'Severity: ' + sevText", 'Severity: text');
contains(js, "aria-label', 'Severity: '", 'severity aria-label');
var renderTL = extractFunction(js, 'renderTimeline');
assert.ok(renderTL.indexOf("el('div', 'sev', event.severity") === -1, 'severity cell must not be color/class-only');

// AC3: AA control-border token exists (numeric ratios are checked in Go).
contains(css, '--control-border:', 'control-border token');
contains(css, 'border: 1px solid var(--control-border)', 'control border uses AA token');
contains(css, 'button:focus-visible', 'focus-visible buttons');

// NFR-021 leftover: card list is windowed.
contains(js, 'var ISSUE_PAGE_SIZE = 25', 'ISSUE_PAGE_SIZE');
contains(js, 'visible.slice(start, start + ISSUE_PAGE_SIZE)', 'page slice');
contains(html, 'id="issue-pager"', 'pager');
contains(html, 'id="issue-prev"', 'previous page');
contains(html, 'id="issue-next"', 'next page');
contains(html, 'aria-label="Issue list pages"', 'pager name');
contains(css, '.issue-pager[hidden]', 'hidden pager not display:flex');

assert.strictEqual(js.indexOf('innerHTML'), -1, 'page.js must not assign innerHTML');
assert.ok(!/https?:\/\//.test(html.replace(/\{\{\.Data\}\}/g, '')), 'page.html has no network URL');
assert.ok(!/cdn|unpkg|jsdelivr/i.test(html + js + css), 'no CDN references');

console.log('nfr013_a11y_test.js: ok');
