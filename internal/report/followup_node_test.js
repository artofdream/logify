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

// FR-021: edit notes/owner/due, export/import round-trip, overdue rule, safe text.
(function followUpDetails() {
  var storage = memoryStorage();
  var s = store([event()], { storage: storage, storageKey: 'details-key' });
  var id = s.createFromEvent(event()).issue.id;
  var xss = '</script><img src=x onerror=alert(1)>';

  assert.strictEqual(s.setOwner(id, '  Ada  ').issue.owner, 'Ada');
  assert.strictEqual(s.setOwner('missing', 'x').error, 'unknown issue');
  assert.strictEqual(s.setOwner(id, new Array(follow.LIMITS.maxOwner + 2).join('a')).error.indexOf('owner exceeds') !== -1, true);
  assert.strictEqual(s.get(id).owner, 'Ada');
  assert.strictEqual(s.setOwner(id, '   ').issue.owner, null);

  assert.strictEqual(s.setDue(id, '2020-01-01').issue.due, '2020-01-01');
  assert.strictEqual(s.setDue(id, '2026-02-30').error, 'due must be YYYY-MM-DD or empty');
  assert.strictEqual(s.get(id).due, '2020-01-01');
  assert.strictEqual(s.setDue(id, '').issue.due, null);
  assert.strictEqual(s.setDue(id, '2026-09-03').issue.due, '2026-09-03');

  assert.strictEqual(s.setNotes(id, xss).issue.notes, xss);
  assert.strictEqual(s.setNotes(id, new Array(follow.LIMITS.maxNotes + 2).join('n')).error.indexOf('notes exceed') !== -1, true);
  assert.strictEqual(s.get(id).notes, xss);
  assert.strictEqual(s.setNotes(id, '').issue.notes, null);
  s.setNotes(id, xss);
  s.setOwner(id, xss);

  var raw = s.exportJSON();
  var parsed = JSON.parse(raw);
  assert.strictEqual(parsed.schema, follow.SCHEMA);
  assert.strictEqual(parsed.schemaVersion, 1);
  assert.strictEqual(parsed.issues[0].owner, xss);
  assert.strictEqual(parsed.issues[0].due, '2026-09-03');
  assert.strictEqual(parsed.issues[0].notes, xss);

  var restored = follow.createStore([event()], {
    now: now, storage: storage, storageKey: 'details-key', reportRoot: '/tmp/case'
  });
  assert.strictEqual(restored.loadLocal().loaded, 1);
  assert.strictEqual(restored.get(id).owner, xss);
  assert.strictEqual(restored.get(id).due, '2026-09-03');
  assert.strictEqual(restored.get(id).notes, xss);

  var other = store([event()]);
  var imported = other.importJSON(raw);
  assert.strictEqual(imported.loaded, 1);
  assert.strictEqual(other.get(id).owner, xss);
  assert.strictEqual(other.get(id).notes, xss);
  assert.strictEqual(other.get(id).due, '2026-09-03');
  assert.strictEqual(other.filter({ owner: 'script' }).length, 1);
  assert.strictEqual(other.filter({ text: 'onerror' }).length, 1);

  // Clock is 2026-09-04 UTC. Due 2026-09-03 is overdue while open.
  assert.strictEqual(follow.isOverdue({ due: '2026-09-03', state: 'open' }, '2026-09-04T12:00:00.000Z'), true);
  assert.strictEqual(follow.isOverdue({ due: '2026-09-04', state: 'open' }, '2026-09-04T12:00:00.000Z'), false);
  assert.strictEqual(follow.isOverdue({ due: '2026-09-05', state: 'open' }, '2026-09-04T12:00:00.000Z'), false);
  assert.strictEqual(follow.isOverdue({ due: '2026-09-03', state: 'investigating' }, '2026-09-04T12:00:00.000Z'), true);
  assert.strictEqual(follow.isOverdue({ due: '2026-09-03', state: 'blocked' }, '2026-09-04T12:00:00.000Z'), true);
  assert.strictEqual(follow.isOverdue({ due: '2026-09-03', state: 'resolved' }, '2026-09-04T12:00:00.000Z'), false);
  assert.strictEqual(follow.isOverdue({ due: '2026-09-03', state: 'dismissed' }, '2026-09-04T12:00:00.000Z'), false);
  assert.strictEqual(follow.isOverdue({ due: null, state: 'open' }, '2026-09-04T12:00:00.000Z'), false);
  assert.strictEqual(s.isOverdue(s.get(id)), true);
  s.setState(id, 'resolved');
  assert.strictEqual(s.isOverdue(s.get(id)), false);
  assert.strictEqual(s.filter({ overdue: true }).length, 0);
  s.setState(id, 'open');
  assert.strictEqual(s.filter({ overdue: true }).length, 1);
  assert.strictEqual(follow.validDue('2026-02-30'), false);
  assert.strictEqual(follow.validDue('2026-01-31'), true);

  var fs = require('fs');
  var path = require('path');
  var page = fs.readFileSync(path.join(__dirname, 'page.js'), 'utf8');
  assert.strictEqual(page.indexOf('innerHTML'), -1);
  assert.ok(page.indexOf('setOwner') !== -1);
  assert.ok(page.indexOf('setDue') !== -1);
  assert.ok(page.indexOf('setNotes') !== -1);
  assert.ok(page.indexOf('overdue-badge') !== -1);
  assert.ok(page.indexOf('textContent') !== -1);
})();

// Empty events/warnings equivalent: store must not throw.
(function emptySafe() {
  var s = follow.createStore(null, { now: now });
  assert.strictEqual(s.eventGroupCount(), 0);
  assert.strictEqual(s.observedRecords(), 0);
  assert.strictEqual(s.createFromEvent({}).error, 'event is missing evidenceId');
})();

// FR-024: link/unlink additional evidence, retain refs, import match review, no silent state change.
(function mergeRecurringEvidence() {
  var e1 = event('evidence-v1-aaa');
  var e2 = {
    evidenceId: 'evidence-v1-bbb',
    signature: 'sig1',
    instance: 'tomcat-a',
    file: 'tomcat-a/localhost.log',
    line: 44,
    message: 'OutOfMemoryError: heap',
    occurrences: 2,
    firstSeen: '2026-09-04T11:00:00.000Z',
    lastSeen: '2026-09-04T11:05:00.000Z',
    severity: 'ERROR',
    sourceType: 'tomcat-java'
  };
  var e3 = {
    evidenceId: 'evidence-v1-ccc',
    signature: 'sig-other',
    instance: 'tomcat-a',
    file: 'tomcat-a/other.log',
    line: 1,
    message: 'unrelated',
    occurrences: 1,
    severity: 'WARN',
    sourceType: 'tomcat-java'
  };

  var s = store([e1, e2, e3]);
  var created = s.createFromEvent(e1);
  assert.strictEqual(created.created, true);
  var id = created.issue.id;
  assert.strictEqual(s.linkEvidence(id, e2).linked, true);
  assert.strictEqual(s.get(id).evidence.id, 'evidence-v1-aaa');
  assert.strictEqual(s.get(id).linkedEvidence.length, 1);
  assert.strictEqual(s.get(id).linkedEvidence[0].id, 'evidence-v1-bbb');
  assert.strictEqual(s.issueForEvidence('evidence-v1-bbb').id, id);
  assert.strictEqual(s.linkEvidence(id, e2).linked, false);
  assert.strictEqual(s.createFromEvent(e2).created, false);
  assert.strictEqual(s.createFromEvent(e2).issue.id, id);
  assert.ok(s.unlinkEvidence(id, 'evidence-v1-aaa').error);
  assert.strictEqual(s.unlinkEvidence(id, 'evidence-v1-bbb').unlinked, true);
  assert.strictEqual(s.get(id).linkedEvidence.length, 0);
  s.linkEvidence(id, e2);
  assert.strictEqual(s.linkEvidence(id, e3).linked, true);
  assert.strictEqual(s.get(id).linkedEvidence.length, 2);
  assert.strictEqual(s.filter({ text: 'localhost.log' }).length, 1);
  assert.strictEqual(s.filter({ instance: 'tomcat-a' }).length, 1);
  s.unlinkEvidence(id, 'evidence-v1-ccc');

  var parsed = JSON.parse(s.exportJSON());
  assert.strictEqual(parsed.schema, follow.SCHEMA);
  assert.strictEqual(parsed.schemaVersion, 1);
  assert.strictEqual(parsed.issues[0].evidence.id, 'evidence-v1-aaa');
  assert.strictEqual(parsed.issues[0].linkedEvidence.length, 1);
  assert.strictEqual(parsed.issues[0].linkedEvidence[0].id, 'evidence-v1-bbb');
  assert.deepStrictEqual(parsed.issues[0].ignoredEvidence, []);

  var round = store([e1, e2]);
  assert.strictEqual(round.importJSON(s.exportJSON()).loaded, 1);
  assert.strictEqual(round.get(id).linkedEvidence[0].id, 'evidence-v1-bbb');
  assert.strictEqual(round.get(id).evidence.id, 'evidence-v1-aaa');
  assert.strictEqual(round.evidenceRefs(round.get(id)).length, 2);

  var src = store([e1]);
  var srcId = src.createFromEvent(e1).issue.id;
  src.setState(srcId, 'resolved');
  var exported = src.exportJSON();
  assert.strictEqual(JSON.parse(exported).issues[0].state, 'resolved');
  assert.strictEqual(JSON.parse(exported).issues[0].evidence.occurrences, 3);

  var newer = Object.assign({}, e1, {
    occurrences: 9,
    lastSeen: '2026-09-05T10:00:00.000Z'
  });
  var rotated = Object.assign({}, e2, {
    evidenceId: 'evidence-v1-ddd',
    file: 'tomcat-a/catalina.out.1',
    line: 3,
    occurrences: 4
  });
  var dest = store([newer, rotated]);
  var imp = dest.importJSON(exported);
  assert.strictEqual(imp.loaded, 1);
  assert.strictEqual(dest.get(srcId).state, 'resolved');
  assert.strictEqual(dest.get(srcId).evidence.occurrences, 3);
  assert.strictEqual(dest.get(srcId).linkedEvidence.length, 0);
  assert.ok(imp.occurrenceUpdates.length >= 1);
  assert.strictEqual(imp.occurrenceUpdates[0].liveOccurrences, 9);
  assert.strictEqual(imp.occurrenceUpdates[0].previousOccurrences, 3);
  assert.ok(imp.candidates.some(function (c) { return c.evidenceId === 'evidence-v1-ddd'; }));
  assert.strictEqual(dest.listReviews().candidates.length, 1);

  dest.acknowledgeOccurrences(srcId, 'evidence-v1-aaa');
  assert.strictEqual(dest.get(srcId).state, 'resolved');
  assert.strictEqual(dest.get(srcId).evidence.occurrences, 9);
  assert.strictEqual(dest.listReviews().occurrenceUpdates.length, 0);

  dest.linkEvidence(srcId, rotated);
  assert.strictEqual(dest.get(srcId).state, 'resolved');
  assert.strictEqual(dest.get(srcId).linkedEvidence[0].id, 'evidence-v1-ddd');
  assert.strictEqual(dest.listReviews().candidates.length, 0);

  var dest2 = store([newer, rotated]);
  dest2.importJSON(exported);
  dest2.dismissCandidate(srcId, 'evidence-v1-ddd');
  assert.strictEqual(dest2.get(srcId).state, 'resolved');
  assert.strictEqual(dest2.listReviews().candidates.length, 0);
  assert.ok(dest2.get(srcId).ignoredEvidence.indexOf('evidence-v1-ddd') !== -1);
  var dismissedExport = JSON.parse(dest2.exportJSON());
  assert.ok(dismissedExport.issues[0].ignoredEvidence.indexOf('evidence-v1-ddd') !== -1);

  var clash = store([e1, e2]);
  clash.createFromEvent(e1);
  clash.createFromEvent(e2);
  assert.ok(clash.linkEvidence('issue-v1-aaa', e2).error);
  assert.strictEqual(clash.get('issue-v1-aaa').state, 'open');
  assert.strictEqual(clash.get('issue-v1-bbb').state, 'open');

  // FR-024 / FR-022: importJSON must not attach one evidence id to two issues.
  var dupPayload = {
    schema: follow.SCHEMA,
    schemaVersion: follow.SCHEMA_VERSION,
    issues: [
      {
        id: 'issue-v1-import-a',
        title: 'first owner',
        state: 'open',
        flagged: false,
        tags: [],
        evidence: { id: 'evidence-v1-aaa', signature: 'sig1', instance: 'tomcat-a', file: 'f', line: 1 }
      },
      {
        id: 'issue-v1-import-b',
        title: 'second owner',
        state: 'open',
        flagged: false,
        tags: [],
        evidence: { id: 'evidence-v1-aaa', signature: 'sig1', instance: 'tomcat-a', file: 'f', line: 1 }
      }
    ]
  };
  var dupStore = store([e1, e2]);
  var dup = dupStore.importJSON(JSON.stringify(dupPayload));
  assert.strictEqual(dup.loaded, 1);
  assert.strictEqual(dup.invalid.length, 1);
  assert.strictEqual(dupStore.issueForEvidence('evidence-v1-aaa').id, 'issue-v1-import-a');
  assert.ok(!dupStore.get('issue-v1-import-b'));
  assert.ok(dup.invalid[0].reason.indexOf('evidence already linked to issue-v1-import-a') !== -1);

  var localOwner = store([e1, e2]);
  localOwner.createFromEvent(e1);
  var steal = {
    schema: follow.SCHEMA,
    schemaVersion: follow.SCHEMA_VERSION,
    issues: [{
      id: 'issue-v1-import-c',
      title: 'steals aaa',
      state: 'investigating',
      flagged: false,
      tags: [],
      evidence: { id: 'evidence-v1-bbb', signature: 'sig1', instance: 'tomcat-a', file: 'f', line: 2 },
      linkedEvidence: [{ id: 'evidence-v1-aaa', signature: 'sig1', instance: 'tomcat-a', file: 'f', line: 1 }]
    }]
  };
  var stolen = localOwner.importJSON(JSON.stringify(steal));
  assert.strictEqual(stolen.loaded, 0);
  assert.strictEqual(stolen.invalid.length, 1);
  assert.strictEqual(localOwner.issueForEvidence('evidence-v1-aaa').id, 'issue-v1-aaa');
  assert.ok(!localOwner.get('issue-v1-import-c'));
  assert.ok(stolen.invalid[0].reason.indexOf('evidence already linked to issue-v1-aaa') !== -1);

  var sameAgain = store([e1]);
  sameAgain.createFromEvent(e1);
  var reload = sameAgain.importJSON(sameAgain.exportJSON());
  assert.strictEqual(reload.loaded, 1);
  assert.strictEqual(reload.invalid.length, 0);
  assert.strictEqual(sameAgain.issueForEvidence('evidence-v1-aaa').id, 'issue-v1-aaa');

  // lastSeen-only delta (equal counts, newer lastSeen) is a reviewable update.
  var storedSnap = { id: 'evidence-v1-aaa', occurrences: 3, lastSeen: '2026-09-04T10:05:00.000Z' };
  var liveSameCount = { occurrences: 3, lastSeen: '2026-09-06T10:00:00.000Z' };
  var lastSeenDelta = follow.occurrenceDelta(storedSnap, liveSameCount);
  assert.ok(lastSeenDelta);
  assert.strictEqual(lastSeenDelta.newOccurrences, 0);
  assert.strictEqual(lastSeenDelta.liveLastSeen, '2026-09-06T10:00:00.000Z');
  assert.strictEqual(follow.occurrenceDelta(storedSnap, {
    occurrences: 3,
    lastSeen: '2026-09-04T10:05:00.000Z'
  }), null);

  var lastSeenLive = Object.assign({}, e1, {
    occurrences: 3,
    lastSeen: '2026-09-06T10:00:00.000Z'
  });
  var lastSeenDest = store([lastSeenLive]);
  var lastSeenImp = lastSeenDest.importJSON(exported);
  assert.strictEqual(lastSeenImp.loaded, 1);
  assert.strictEqual(lastSeenImp.occurrenceUpdates.length, 1);
  assert.strictEqual(lastSeenImp.occurrenceUpdates[0].newOccurrences, 0);
  assert.strictEqual(lastSeenImp.occurrenceUpdates[0].previousOccurrences, 3);
  assert.strictEqual(lastSeenImp.occurrenceUpdates[0].liveOccurrences, 3);
  assert.strictEqual(lastSeenImp.occurrenceUpdates[0].liveLastSeen, '2026-09-06T10:00:00.000Z');
  assert.ok(lastSeenImp.occurrenceUpdates[0].previousLastSeen);

  var fs = require('fs');
  var path = require('path');
  var page = fs.readFileSync(path.join(__dirname, 'page.js'), 'utf8');
  var html = fs.readFileSync(path.join(__dirname, 'page.html'), 'utf8');
  assert.strictEqual(page.indexOf('innerHTML'), -1);
  assert.ok(page.indexOf('Link to existing issue') !== -1);
  assert.ok(page.indexOf('listReviews') !== -1);
  assert.ok(page.indexOf('linkEvidence') !== -1);
  assert.ok(page.indexOf('unlinkEvidence') !== -1);
  assert.ok(page.indexOf('Newer last-seen time on ') !== -1);
  assert.ok(page.indexOf('Occurrence count is unchanged') !== -1);
  assert.ok(html.indexOf('Recurring evidence review') !== -1);
})();

console.log('followup_node_test.js: ok');
