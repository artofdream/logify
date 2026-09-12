'use strict';

// NFR-021 AC4: measure store.filter() with 10,000 synthetic issues.
// The live page windows matching cards to ISSUE_PAGE_SIZE (see page.js).
// This probe still does not paint DOM cards and is not a published workstation spec.

var assert = require('assert');
var os = require('os');
var follow = require('./followup.js');

var COUNT = 10000;
var ISSUE_PAGE_SIZE = 25;
var INTERACTIVE_MS = 100;
var CI_GUARD_MS = 500;
var STATES = follow.STATES;

function pad(n, width) {
  var s = String(n);
  while (s.length < width) s = '0' + s;
  return s;
}

function hostInfo() {
  var cpu = os.cpus()[0] || {};
  return {
    platform: os.platform(),
    release: os.release(),
    arch: os.arch(),
    cpus: os.cpus().length,
    cpuModel: cpu.model || 'unknown',
    totalMemMB: Math.round(os.totalmem() / 1024 / 1024),
    node: process.version
  };
}

function syntheticIssue(i) {
  var id = pad(i, 5);
  return {
    id: 'issue-v1-' + id,
    title: 'Synthetic issue ' + id,
    state: STATES[i % STATES.length],
    flagged: i % 7 === 0,
    tags: ['tag-' + (i % 20), 'batch'],
    owner: i % 3 === 0 ? 'owner-' + (i % 10) : null,
    due: i % 11 === 0 ? '2020-01-01' : null,
    notes: 'notes for ' + id,
    createdAt: '2026-09-04T12:00:00.000Z',
    modifiedAt: '2026-09-04T12:00:00.000Z',
    evidence: {
      id: 'evidence-v1-' + id,
      signature: 'sig-' + (i % 50),
      instance: 'inst-' + (i % 5),
      file: 'logs/file-' + id + '.log',
      line: i + 1,
      firstSeen: '2026-09-04T12:00:00.000Z',
      lastSeen: '2026-09-04T12:00:00.000Z',
      occurrences: 1,
      severity: i % 2 === 0 ? 'ERROR' : 'INFO',
      sourceType: 'tomcat-java'
    }
  };
}

function nowMs() {
  var hr = process.hrtime();
  return hr[0] * 1e3 + hr[1] / 1e6;
}

function time(fn) {
  var start = nowMs();
  var result = fn();
  return { ms: nowMs() - start, result: result };
}

function median(values) {
  var sorted = values.slice().sort(function (a, b) { return a - b; });
  var mid = Math.floor(sorted.length / 2);
  if (!sorted.length) return 0;
  if (sorted.length % 2) return sorted[mid];
  return (sorted[mid - 1] + sorted[mid]) / 2;
}

function measure(fn, rounds) {
  var times = [];
  var last;
  for (var i = 0; i < rounds; i++) {
    var run = time(fn);
    times.push(run.ms);
    last = run.result;
  }
  return {
    count: last && typeof last.length === 'number' ? last.length : null,
    minMs: Math.min.apply(null, times),
    medianMs: median(times),
    maxMs: Math.max.apply(null, times)
  };
}

var payload = {
  schema: follow.SCHEMA,
  schemaVersion: follow.SCHEMA_VERSION,
  issues: []
};
for (var i = 0; i < COUNT; i++) payload.issues.push(syntheticIssue(i));

var store = follow.createStore([], {
  now: function () { return '2026-09-04T12:00:00.000Z'; }
});
var loaded = store.importJSON(JSON.stringify(payload));
assert.strictEqual(loaded.error, undefined, loaded.error);
assert.strictEqual(loaded.loaded, COUNT);
assert.strictEqual(store.issueCount(), COUNT);

var cases = [
  { name: 'all', criteria: {} },
  { name: 'text-few', criteria: { text: 'Synthetic issue 00042' } },
  { name: 'text-many', criteria: { text: 'Synthetic' } },
  { name: 'flagged', criteria: { flagged: true } },
  { name: 'state-open', criteria: { state: 'open' } },
  { name: 'tags', criteria: { tags: 'tag-3,batch' } },
  { name: 'owner', criteria: { owner: 'owner-1' } },
  { name: 'overdue', criteria: { overdue: true } },
  { name: 'combined', criteria: { flagged: true, state: 'open', tags: 'batch', severity: 'ERROR' } }
];

// Warm the hidden class / IC cache before timed rounds.
cases.forEach(function (c) { store.filter(c.criteria); });

var results = cases.map(function (c) {
  var stats = measure(function () { return store.filter(c.criteria); }, 5);
  var windowStats = measure(function () {
    return store.filter(c.criteria).slice(0, ISSUE_PAGE_SIZE);
  }, 5);
  return {
    name: c.name,
    matched: stats.count,
    minMs: Number(stats.minMs.toFixed(3)),
    medianMs: Number(stats.medianMs.toFixed(3)),
    maxMs: Number(stats.maxMs.toFixed(3)),
    windowed: windowStats.count,
    windowMedianMs: Number(windowStats.medianMs.toFixed(3)),
    windowMaxMs: Number(windowStats.maxMs.toFixed(3))
  };
});

assert.strictEqual(results[0].matched, COUNT);
assert.ok(results[1].matched >= 1);
assert.strictEqual(results[2].matched, COUNT);
assert.ok(results[3].matched > 0 && results[3].matched < COUNT);

var worst = results.reduce(function (max, row) {
  return row.maxMs > max ? row.maxMs : max;
}, 0);
var worstMedian = results.reduce(function (max, row) {
  return row.medianMs > max ? row.medianMs : max;
}, 0);

var report = {
  requirement: 'NFR-021',
  acceptance: 'AC4',
  issueCount: COUNT,
  issuePageSize: ISSUE_PAGE_SIZE,
  interactiveTargetMs: INTERACTIVE_MS,
  ciGuardMs: CI_GUARD_MS,
  host: hostInfo(),
  disclaimer: 'Probe host is not a published operator reference workstation. Times measure store.filter() and a 25-item window slice, not painted DOM cards. The live report pages matching cards (ISSUE_PAGE_SIZE=25). A green CI check is not proof for every machine or for Q-001 reference hardware.',
  filters: results,
  worstMedianMs: Number(worstMedian.toFixed(3)),
  worstMaxMs: Number(worst.toFixed(3)),
  interactiveTargetMetOnHost: worstMedian <= INTERACTIVE_MS
};

console.log(JSON.stringify(report, null, 2));

if (worst > CI_GUARD_MS) {
  console.error('NFR-021 AC4 CI guard failed: worst filter ' + worst.toFixed(3) + 'ms > ' + CI_GUARD_MS + 'ms');
  process.exit(1);
}

console.log('nfr021_filter_probe.js: ok (worst median ' + worstMedian.toFixed(3) + 'ms, interactive target ' + INTERACTIVE_MS + 'ms ' + (report.interactiveTargetMetOnHost ? 'met' : 'not met') + ' on this host)');
