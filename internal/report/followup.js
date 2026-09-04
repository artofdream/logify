(function (root, factory) {
  var api = factory();
  root.LogifyFollowUp = api;
  if (typeof module !== 'undefined' && module.exports) {
    module.exports = api;
  }
})(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  'use strict';

  var SCHEMA = 'logify-follow-up-v1';
  var SCHEMA_VERSION = 1;
  var STATES = ['open', 'investigating', 'blocked', 'resolved', 'dismissed'];
  var CLOSED = { resolved: true, dismissed: true };
  var LIMITS = {
    maxBytes: 5 * 1024 * 1024,
    maxIssues: 10000,
    maxTitle: 200,
    maxTags: 50,
    maxTagLength: 64,
    maxNotes: 8000,
    maxOwner: 200
  };
  var ISSUE_FIELDS = {
    id: true, title: true, state: true, flagged: true, tags: true,
    owner: true, due: true, notes: true, createdAt: true, modifiedAt: true,
    evidence: true
  };

  function isArray(v) { return Array.isArray(v); }
  function isObject(v) { return v && typeof v === 'object' && !isArray(v); }
  function text(v) { return v == null ? '' : String(v); }
  function tagKey(v) { return text(v).trim().toLowerCase(); }

  function defaultTitle(message) {
    var first = text(message).split(/\r?\n/, 1)[0].trim();
    return (first || 'Follow up on timeline evidence').slice(0, LIMITS.maxTitle);
  }

  function isoNow(clock) {
    var d = clock ? clock() : new Date();
    if (typeof d === 'string') return d;
    return d.toISOString();
  }

  function issueIDFromEvidence(evidenceId) {
    return 'issue-v1-' + text(evidenceId).replace(/^evidence-v1-/, '');
  }

  function clone(v) {
    return JSON.parse(JSON.stringify(v));
  }

  function evidenceSnapshot(event) {
    if (!event) return null;
    return {
      id: event.evidenceId,
      signature: event.signature,
      instance: event.instance,
      file: event.file,
      line: event.line,
      firstSeen: event.firstSeen || null,
      lastSeen: event.lastSeen || null,
      occurrences: event.occurrences,
      severity: event.severity,
      sourceType: event.sourceType
    };
  }

  function normalizeTag(raw) {
    var value = text(raw).trim();
    if (!value) return { error: 'empty tag' };
    if (value.length > LIMITS.maxTagLength) {
      return { error: 'tag exceeds ' + LIMITS.maxTagLength + ' characters' };
    }
    return { value: value };
  }

  function validState(v) {
    return STATES.indexOf(v) !== -1;
  }

  function validDue(v) {
    if (v == null || v === '') return true;
    return /^\d{4}-\d{2}-\d{2}$/.test(String(v));
  }

  function isOverdue(issue, nowIso) {
    if (!issue.due || CLOSED[issue.state]) return false;
    var today = (nowIso || isoNow()).slice(0, 10);
    return issue.due < today;
  }

  function extraFields(raw) {
    var extra = {};
    if (!isObject(raw)) return extra;
    Object.keys(raw).forEach(function (key) {
      if (!ISSUE_FIELDS[key]) extra[key] = raw[key];
    });
    return extra;
  }

  function validateIssue(raw, index) {
    var loc = 'issues[' + index + ']';
    if (!isObject(raw)) return { error: loc + ' is not an object' };
    if (typeof raw.id !== 'string' || raw.id.indexOf('issue-v1-') !== 0) {
      return { error: loc + ' has an invalid id' };
    }
    if (typeof raw.title !== 'string' || raw.title.length > LIMITS.maxTitle) {
      return { error: loc + ' has an invalid title' };
    }
    if (!validState(raw.state)) return { error: loc + ' has an invalid state' };
    if (typeof raw.flagged !== 'boolean') return { error: loc + ' flagged must be boolean' };
    if (!isArray(raw.tags)) return { error: loc + ' tags must be an array' };
    if (raw.tags.length > LIMITS.maxTags) return { error: loc + ' has too many tags' };
    var seen = {};
    for (var i = 0; i < raw.tags.length; i++) {
      if (typeof raw.tags[i] !== 'string') return { error: loc + ' tags must be strings' };
      var n = normalizeTag(raw.tags[i]);
      if (n.error) return { error: loc + ' ' + n.error };
      var k = tagKey(n.value);
      if (seen[k]) return { error: loc + ' has duplicate tag ' + n.value };
      seen[k] = true;
    }
    if (raw.owner != null && (typeof raw.owner !== 'string' || raw.owner.length > LIMITS.maxOwner)) {
      return { error: loc + ' has an invalid owner' };
    }
    if (raw.notes != null && (typeof raw.notes !== 'string' || raw.notes.length > LIMITS.maxNotes)) {
      return { error: loc + ' has invalid notes' };
    }
    if (!validDue(raw.due)) return { error: loc + ' due must be YYYY-MM-DD or null' };
    if (!isObject(raw.evidence) || typeof raw.evidence.id !== 'string' || raw.evidence.id.indexOf('evidence-v1-') !== 0) {
      return { error: loc + ' evidence.id is missing or unversioned' };
    }
    return { issue: raw };
  }

  function parseExport(text) {
    if (typeof text !== 'string') return { error: 'import must be a JSON string' };
    if (text.length > LIMITS.maxBytes) {
      return { error: 'import exceeds ' + LIMITS.maxBytes + ' bytes' };
    }
    var data;
    try {
      data = JSON.parse(text);
    } catch (err) {
      return { error: 'import is not valid JSON' };
    }
    if (!isObject(data)) return { error: 'import root must be an object' };
    if (data.schema !== SCHEMA || data.schemaVersion !== SCHEMA_VERSION) {
      return { error: 'unsupported schema (want ' + SCHEMA + ' version ' + SCHEMA_VERSION + ')' };
    }
    if (!isArray(data.issues)) return { error: 'issues must be an array' };
    if (data.issues.length > LIMITS.maxIssues) {
      return { error: 'import exceeds ' + LIMITS.maxIssues + ' issues' };
    }
    return { data: data };
  }

  function createStore(events, options) {
    options = options || {};
    var list = isArray(events) ? events : [];
    var byEvidence = {};
    list.forEach(function (event) {
      if (event && event.evidenceId) byEvidence[event.evidenceId] = event;
    });
    var issues = {};
    var order = [];
    var dirtySinceExport = false;
    var storage = options.storage || null;
    var storageKey = options.storageKey || '';

    function now() { return isoNow(options.now); }

    function get(id) { return issues[id] || null; }

    function all() {
      return order.map(function (id) { return issues[id]; }).filter(Boolean);
    }

    function touch(issue) {
      issue.modifiedAt = now();
      dirtySinceExport = true;
    }

    function remember(issue) {
      if (!issues[issue.id]) order.push(issue.id);
      issues[issue.id] = issue;
    }

    function persist() {
      if (!storage || !storageKey) return { persisted: false, reason: 'no local storage' };
      try {
        storage.setItem(storageKey, exportJSON());
        return { persisted: true };
      } catch (err) {
        return { persisted: false, reason: 'local storage write failed' };
      }
    }

    function loadLocal() {
      if (!storage || !storageKey) return { loaded: 0 };
      var raw;
      try { raw = storage.getItem(storageKey); } catch (err) { return { loaded: 0, error: 'local storage read failed' }; }
      if (!raw) return { loaded: 0 };
      var result = importJSON(raw, { fromLocal: true });
      dirtySinceExport = false;
      return result;
    }

    function clearLocal() {
      issues = {};
      order = [];
      dirtySinceExport = false;
      if (storage && storageKey) {
        try { storage.removeItem(storageKey); } catch (err) { /* ignore */ }
      }
    }

    function createFromEvent(event) {
      if (!event || !event.evidenceId) return { error: 'event is missing evidenceId' };
      var id = issueIDFromEvidence(event.evidenceId);
      var existing = get(id);
      if (existing) return { issue: existing, created: false };
      if (order.length >= LIMITS.maxIssues) return { error: 'issue limit reached' };
      var issue = {
        id: id,
        title: defaultTitle(event.message),
        state: 'open',
        flagged: false,
        tags: [],
        owner: null,
        due: null,
        notes: null,
        createdAt: now(),
        modifiedAt: now(),
        evidence: evidenceSnapshot(event),
        evidenceMatched: true,
        extra: {}
      };
      remember(issue);
      dirtySinceExport = true;
      persist();
      return { issue: issue, created: true };
    }

    function updateTitle(id, title) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      var next = text(title).slice(0, LIMITS.maxTitle);
      issue.title = next;
      touch(issue);
      persist();
      return { issue: issue };
    }

    function addTag(id, raw) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      var n = normalizeTag(raw);
      if (n.error) return { error: n.error };
      if (issue.tags.some(function (t) { return tagKey(t) === tagKey(n.value); })) {
        return { issue: issue, added: false };
      }
      if (issue.tags.length >= LIMITS.maxTags) return { error: 'tag limit reached' };
      issue.tags.push(n.value);
      touch(issue);
      persist();
      return { issue: issue, added: true };
    }

    function removeTag(id, raw) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      var k = tagKey(raw);
      var before = issue.tags.length;
      issue.tags = issue.tags.filter(function (t) { return tagKey(t) !== k; });
      if (issue.tags.length !== before) {
        touch(issue);
        persist();
      }
      return { issue: issue };
    }

    function setFlagged(id, flagged) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      issue.flagged = !!flagged;
      touch(issue);
      persist();
      return { issue: issue };
    }

    function setState(id, state) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      if (!validState(state)) return { error: 'invalid state' };
      issue.state = state;
      touch(issue);
      persist();
      return { issue: issue };
    }

    function filter(criteria) {
      criteria = criteria || {};
      var tags = [];
      if (typeof criteria.tags === 'string') {
        tags = criteria.tags.split(',').map(function (t) { return t.trim(); }).filter(Boolean);
      } else if (isArray(criteria.tags)) {
        tags = criteria.tags.filter(Boolean);
      }
      return all().filter(function (issue) {
        if (criteria.text) {
          var q = String(criteria.text).toLowerCase();
          var hay = [
            issue.title, issue.id, issue.tags.join(' '), issue.owner || '',
            issue.evidence.file, issue.evidence.signature, issue.evidence.instance
          ].join(' ').toLowerCase();
          if (hay.indexOf(q) === -1) return false;
        }
        for (var i = 0; i < tags.length; i++) {
          var want = tagKey(tags[i]);
          if (!issue.tags.some(function (t) { return tagKey(t) === want; })) return false;
        }
        if (criteria.flagged === true && !issue.flagged) return false;
        if (criteria.state && issue.state !== criteria.state) return false;
        if (criteria.owner) {
          if (text(issue.owner).toLowerCase().indexOf(String(criteria.owner).toLowerCase()) === -1) return false;
        }
        if (criteria.severity && issue.evidence.severity !== criteria.severity) return false;
        if (criteria.instance && issue.evidence.instance !== criteria.instance) return false;
        if (criteria.overdue === true && !isOverdue(issue, now())) return false;
        return true;
      });
    }

    function exportObject() {
      return {
        schema: SCHEMA,
        schemaVersion: SCHEMA_VERSION,
        exportedAt: now(),
        reportRoot: options.reportRoot || '',
        issues: all().map(function (issue) {
          var row = Object.assign({}, issue.extra || {}, {
            id: issue.id,
            title: issue.title,
            state: issue.state,
            flagged: issue.flagged,
            tags: issue.tags.slice(),
            owner: issue.owner,
            due: issue.due,
            notes: issue.notes,
            createdAt: issue.createdAt,
            modifiedAt: issue.modifiedAt,
            evidence: clone(issue.evidence)
          });
          return row;
        })
      };
    }

    function exportJSON() {
      return JSON.stringify(exportObject(), null, 2);
    }

    function hydrate(raw, liveEvent) {
      var tags = raw.tags.map(function (t) { return normalizeTag(t).value; });
      var evidence = Object.assign({}, raw.evidence);
      var matched = !!(liveEvent && liveEvent.evidenceId === evidence.id);
      if (matched) {
        evidence.occurrences = liveEvent.occurrences;
        evidence.firstSeen = liveEvent.firstSeen || evidence.firstSeen || null;
        evidence.lastSeen = liveEvent.lastSeen || evidence.lastSeen || null;
        evidence.severity = liveEvent.severity;
        evidence.sourceType = liveEvent.sourceType;
      }
      return {
        id: raw.id,
        title: raw.title,
        state: raw.state,
        flagged: raw.flagged,
        tags: tags,
        owner: raw.owner == null || raw.owner === '' ? null : raw.owner,
        due: raw.due == null || raw.due === '' ? null : raw.due,
        notes: raw.notes == null || raw.notes === '' ? null : raw.notes,
        createdAt: raw.createdAt || now(),
        modifiedAt: raw.modifiedAt || now(),
        evidence: evidence,
        evidenceMatched: matched,
        extra: extraFields(raw)
      };
    }

    function importJSON(text, meta) {
      var parsed = parseExport(text);
      if (parsed.error) return { error: parsed.error, loaded: 0, invalid: [], unmatched: [] };
      var invalid = [];
      var unmatched = [];
      var loaded = 0;
      parsed.data.issues.forEach(function (raw, index) {
        var checked = validateIssue(raw, index);
        if (checked.error) {
          invalid.push({ index: index, id: raw && raw.id, reason: checked.error });
          return;
        }
        var live = byEvidence[checked.issue.evidence.id];
        var issue = hydrate(checked.issue, live);
        if (!issue.evidenceMatched) {
          unmatched.push({ id: issue.id, evidenceId: issue.evidence.id });
        }
        remember(issue);
        loaded += 1;
      });
      if (!meta || !meta.fromLocal) dirtySinceExport = true;
      persist();
      return { loaded: loaded, invalid: invalid, unmatched: unmatched };
    }

    return {
      createFromEvent: createFromEvent,
      updateTitle: updateTitle,
      addTag: addTag,
      removeTag: removeTag,
      setFlagged: setFlagged,
      setState: setState,
      get: get,
      list: all,
      filter: filter,
      exportJSON: exportJSON,
      exportObject: exportObject,
      importJSON: importJSON,
      persist: persist,
      markExported: function () { dirtySinceExport = false; },
      loadLocal: loadLocal,
      clearLocal: clearLocal,
      isOverdue: function (issue) { return isOverdue(issue, now()); },
      hasUnexportedChanges: function () { return dirtySinceExport; },
      observedRecords: function () {
        return list.reduce(function (n, event) { return n + (Number(event.occurrences) || 0); }, 0);
      },
      eventGroupCount: function () { return list.length; },
      issueCount: function () { return order.length; },
      storageDescription: function () {
        if (storage && storageKey) {
          return 'Browser local storage key ' + storageKey + ' (convenience only). Export JSON is the portable copy. Nothing is sent over a network.';
        }
        return 'This tab only until you export JSON. Nothing is sent over a network.';
      }
    };
  }

  return {
    SCHEMA: SCHEMA,
    SCHEMA_VERSION: SCHEMA_VERSION,
    STATES: STATES,
    LIMITS: LIMITS,
    defaultTitle: defaultTitle,
    issueIDFromEvidence: issueIDFromEvidence,
    parseExport: parseExport,
    createStore: createStore
  };
});
