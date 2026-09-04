'use strict';

var assert = require('assert');
var follow = require('./followup.js');

function event(id) {
  return {
    evidenceId: id || 'evidence-v1-aaa',
    signature: 'sig1',
    instance: 'tomcat-a',
    file: 'tomcat-a/catalina.out',
    line: 10,
    message: 'OutOfMemoryError: heap\n  at x',
    occurrences: 3,
    firstSeen: '2026-09-04T10:00:00.000Z',
    lastSeen: '2026-09-04T10:05:00.000Z',
    severity: 'ERROR',
    sourceType: 'tomcat-java',
    hasTimestamp: true,
    timestamp: '2026-09-04T10:00:00.000Z'
  };
}

function memoryStorage() {
  var m = Object.create(null);
  return {
    getItem: function (k) { return Object.prototype.hasOwnProperty.call(m, k) ? m[k] : null; },
    setItem: function (k, v) { m[k] = String(v); },
    removeItem: function (k) { delete m[k]; }
  };
}

var clock = 0;
function now() {
  clock += 1;
  return new Date(Date.UTC(2026, 8, 4, 12, 0, clock)).toISOString();
}

function store(events, extra) {
  return follow.createStore(events || [event()], Object.assign({
    now: now,
    storage: memoryStorage(),
    storageKey: 'test-key',
    reportRoot: '/tmp/case'
  }, extra || {}));
}

// FR-017 create from one event / group, stable id, editable title, evidence link.
(function createIssue() {
  var s = store();
  var first = s.createFromEvent(event());
  assert.strictEqual(first.created, true);
  assert.strictEqual(first.issue.id, 'issue-v1-aaa');
  assert.strictEqual(first.issue.state, 'open');
  assert.strictEqual(first.issue.flagged, false);
  assert.strictEqual(first.issue.title, 'OutOfMemoryError: heap');
  assert.strictEqual(first.issue.evidence.signature, 'sig1');
  assert.strictEqual(first.issue.evidence.instance, 'tomcat-a');
  assert.strictEqual(first.issue.evidence.file, 'tomcat-a/catalina.out');
  assert.strictEqual(first.issue.evidence.line, 10);
  assert.strictEqual(first.issue.evidence.occurrences, 3);
  assert.ok(first.issue.evidence.firstSeen);
  assert.ok(first.issue.evidence.lastSeen);
  var again = s.createFromEvent(event());
  assert.strictEqual(again.created, false);
  assert.strictEqual(again.issue.id, first.issue.id);
  s.updateTitle(first.issue.id, '</script><img src=x onerror=alert(1)>');
  assert.strictEqual(s.get(first.issue.id).title, '</script><img src=x onerror=alert(1)>');
  assert.strictEqual(s.issueCount(), 1);
})();

// FR-018 tags: add/remove, trim, case-insensitive unique, filter.
(function tags() {
  var s = store();
  var id = s.createFromEvent(event()).issue.id;
  assert.strictEqual(s.addTag(id, '  DB  ').added, true);
  assert.strictEqual(s.addTag(id, 'db').added, false);
  assert.strictEqual(s.addTag(id, 'Network').added, true);
  assert.deepStrictEqual(s.get(id).tags, ['DB', 'Network']);
  s.removeTag(id, 'db');
  assert.deepStrictEqual(s.get(id).tags, ['Network']);
  s.addTag(id, 'disk');
  assert.strictEqual(s.filter({ tags: 'network' }).length, 1);
  assert.strictEqual(s.filter({ tags: 'network,disk' }).length, 1);
  assert.strictEqual(s.filter({ tags: 'network,missing' }).length, 0);
})();

// FR-019 flag/unflag, filter, persist via storage.
(function flags() {
  var storage = memoryStorage();
  var s = store([event()], { storage: storage, storageKey: 'flag-key' });
  var id = s.createFromEvent(event()).issue.id;
  s.setFlagged(id, true);
  assert.strictEqual(s.get(id).flagged, true);
  assert.strictEqual(s.filter({ flagged: true }).length, 1);
  s.setFlagged(id, false);
  assert.strictEqual(s.filter({ flagged: true }).length, 0);
  s.setFlagged(id, true);
  var restored = follow.createStore([event()], {
    now: now, storage: storage, storageKey: 'flag-key', reportRoot: '/tmp/case'
  });
  var load = restored.loadLocal();
  assert.strictEqual(load.loaded, 1);
  assert.strictEqual(restored.get(id).flagged, true);
})();

// FR-020 workflow states, default open, filter, modified time, resolve keeps evidence.
(function workflow() {
  var s = store();
  var issue = s.createFromEvent(event()).issue;
  assert.strictEqual(issue.state, 'open');
  var before = issue.modifiedAt;
  s.setState(issue.id, 'investigating');
  assert.strictEqual(s.get(issue.id).state, 'investigating');
  assert.notStrictEqual(s.get(issue.id).modifiedAt, before);
  s.setState(issue.id, 'resolved');
  var resolved = s.get(issue.id);
  assert.strictEqual(resolved.state, 'resolved');
  assert.strictEqual(resolved.evidence.signature, 'sig1');
  assert.strictEqual(resolved.evidence.occurrences, 3);
  assert.strictEqual(s.setState(issue.id, 'nope').error, 'invalid state');
  assert.strictEqual(s.filter({ state: 'resolved' }).length, 1);
  assert.strictEqual(s.filter({ state: 'open' }).length, 0);
})();

// FR-022 export/import: schema, validation, unmatched ids, no network.
(function exportImport() {
  var s = store();
  var id = s.createFromEvent(event()).issue.id;
  s.addTag(id, 'db');
  s.setFlagged(id, true);
  s.setState(id, 'blocked');
  var raw = s.exportJSON();
  var parsed = JSON.parse(raw);
  assert.strictEqual(parsed.schema, follow.SCHEMA);
  assert.strictEqual(parsed.schemaVersion, follow.SCHEMA_VERSION);
  assert.ok(parsed.exportedAt);
  assert.strictEqual(parsed.issues.length, 1);
  assert.strictEqual(parsed.issues[0].id, id);
  assert.strictEqual(parsed.issues[0].evidence.id, 'evidence-v1-aaa');
  assert.strictEqual(parsed.issues[0].notes, null);
  assert.strictEqual(parsed.issues[0].owner, null);

  var other = store([event('evidence-v1-bbb')]);
  var result = other.importJSON(raw);
  assert.strictEqual(result.loaded, 1);
  assert.strictEqual(result.unmatched.length, 1);
  assert.strictEqual(other.get(id).evidenceMatched, false);
  assert.strictEqual(other.get(id).flagged, true);
  assert.strictEqual(other.get(id).state, 'blocked');

  var bad = other.importJSON('{"schema":"nope","schemaVersion":9,"issues":[]}');
  assert.ok(bad.error);
  assert.strictEqual(bad.loaded, 0);

  var mixed = {
    schema: follow.SCHEMA,
    schemaVersion: follow.SCHEMA_VERSION,
    issues: [
      parsed.issues[0],
      { id: 'not-an-issue', title: 'x', state: 'open', flagged: false, tags: [], evidence: { id: 'evidence-v1-x' } },
      Object.assign({}, parsed.issues[0], { id: 'issue-v1-from-bbb', evidence: { id: 'evidence-v1-bbb', signature: 's', instance: 'i', file: 'f', line: 1 } })
    ]
  };
  var target = store([event('evidence-v1-bbb')]);
  var mix = target.importJSON(JSON.stringify(mixed));
  assert.strictEqual(mix.loaded, 2);
  assert.strictEqual(mix.invalid.length, 1);
  assert.ok(target.get('issue-v1-from-bbb').evidenceMatched);
})();

// FR-023 combined filters, overdue imported due, counts.
(function queueFilters() {
  var s = store([event(), {
    evidenceId: 'evidence-v1-ccc',
    signature: 'sig2',
    instance: 'httpd',
    file: 'httpd/access_log',
    line: 2,
    message: 'GET / 500',
    occurrences: 1,
    severity: 'ERROR',
    sourceType: 'apache-access'
  }]);
  assert.strictEqual(s.observedRecords(), 4);
  assert.strictEqual(s.eventGroupCount(), 2);
  var a = s.createFromEvent(event()).issue;
  var b = s.createFromEvent({
    evidenceId: 'evidence-v1-ccc',
    signature: 'sig2',
    instance: 'httpd',
    file: 'httpd/access_log',
    line: 2,
    message: 'GET / 500',
    occurrences: 1,
    severity: 'ERROR',
    sourceType: 'apache-access'
  }).issue;
  s.addTag(a.id, 'db');
  s.setFlagged(a.id, true);
  s.setState(b.id, 'investigating');
  var exported = JSON.parse(s.exportJSON());
  exported.issues[0].owner = 'ada';
  exported.issues[0].due = '2020-01-01';
  var round = store([event(), {
    evidenceId: 'evidence-v1-ccc', signature: 'sig2', instance: 'httpd',
    file: 'httpd/access_log', line: 2, message: 'GET / 500', occurrences: 1,
    severity: 'ERROR', sourceType: 'apache-access'
  }]);
  round.importJSON(JSON.stringify(exported));
  assert.strictEqual(round.filter({ flagged: true, tags: 'db', owner: 'ada', overdue: true, severity: 'ERROR', instance: 'tomcat-a' }).length, 1);
  assert.strictEqual(round.filter({ state: 'investigating' }).length, 1);
  assert.strictEqual(round.filter({ text: 'access_log' }).length, 1);
  assert.strictEqual(round.issueCount(), 2);
})();

// FR-023: overdue uses the operator local calendar day, not the UTC ISO date.
(function overdueUsesLocalDate() {
  function pad(n) { return n < 10 ? '0' + n : String(n); }
  function localYMD(d) {
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate());
  }
  // 04:00 UTC is the previous local day west of UTC; 20:00 UTC is the next
  // local day east of UTC. Either can disagree with toISOString().slice(0, 10).
  var offset = new Date(Date.UTC(2026, 8, 4, 12, 0, 0)).getTimezoneOffset();
  var instant = offset > 0
    ? new Date(Date.UTC(2026, 8, 4, 4, 0, 0))
    : new Date(Date.UTC(2026, 8, 4, 20, 0, 0));
  var clockNow = function () { return instant.toISOString(); };
  var localDay = localYMD(instant);
  var utcDay = instant.toISOString().slice(0, 10);
  var s = store([event()], { now: clockNow });
  var id = s.createFromEvent(event()).issue.id;
  var exported = JSON.parse(s.exportJSON());

  exported.issues[0].due = localDay;
  var today = store([event()], { now: clockNow });
  today.importJSON(JSON.stringify(exported));
  assert.strictEqual(today.isOverdue(today.get(id)), false);
  assert.strictEqual(today.filter({ overdue: true }).length, 0);

  exported.issues[0].due = utcDay;
  var utc = store([event()], { now: clockNow });
  utc.importJSON(JSON.stringify(exported));
  assert.strictEqual(utc.isOverdue(utc.get(id)), utcDay < localDay);
  assert.strictEqual(utc.filter({ overdue: true }).length, utcDay < localDay ? 1 : 0);
})();

// Empty events/warnings equivalent: store must not throw.
(function emptySafe() {
  var s = follow.createStore(null, { now: now });
  assert.strictEqual(s.eventGroupCount(), 0);
  assert.strictEqual(s.observedRecords(), 0);
  assert.strictEqual(s.createFromEvent({}).error, 'event is missing evidenceId');
})();

console.log('followup_node_test.js: ok');
