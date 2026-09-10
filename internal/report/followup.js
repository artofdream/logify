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
    maxOwner: 200,
    maxLinkedEvidence: 50
  };
  var ISSUE_FIELDS = {
    id: true, title: true, state: true, flagged: true, tags: true,
    owner: true, due: true, notes: true, createdAt: true, modifiedAt: true,
    evidence: true, linkedEvidence: true, ignoredEvidence: true
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

  function allEvidence(issue) {
    if (!issue || !issue.evidence) return [];
    var extra = isArray(issue.linkedEvidence) ? issue.linkedEvidence : [];
    return [issue.evidence].concat(extra);
  }

  function signatureKey(ref) {
    if (!ref) return '';
    return text(ref.signature) + '\0' + text(ref.instance);
  }

  function validEvidenceID(id) {
    return typeof id === 'string' && id.indexOf('evidence-v1-') === 0;
  }

  function normalizeEvidenceRef(raw) {
    if (!isObject(raw) || !validEvidenceID(raw.id)) return null;
    return {
      id: raw.id,
      signature: raw.signature,
      instance: raw.instance,
      file: raw.file,
      line: raw.line,
      firstSeen: raw.firstSeen || null,
      lastSeen: raw.lastSeen || null,
      occurrences: raw.occurrences,
      severity: raw.severity,
      sourceType: raw.sourceType
    };
  }

  function occurrenceDelta(stored, live) {
    if (!stored || !live) return null;
    var prevOcc = Number(stored.occurrences) || 0;
    var liveOcc = Number(live.occurrences) || 0;
    var prevLast = stored.lastSeen || '';
    var liveLast = live.lastSeen || '';
    var newerLast = !!(liveLast && (!prevLast || liveLast > prevLast));
    if (liveOcc <= prevOcc && !newerLast) return null;
    return {
      evidenceId: stored.id,
      previousOccurrences: prevOcc,
      liveOccurrences: liveOcc,
      newOccurrences: Math.max(0, liveOcc - prevOcc),
      previousLastSeen: stored.lastSeen || null,
      liveLastSeen: live.lastSeen || null
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
    if (!/^\d{4}-\d{2}-\d{2}$/.test(String(v))) return false;
    var parts = String(v).split('-');
    var y = Number(parts[0]);
    var m = Number(parts[1]);
    var d = Number(parts[2]);
    var dt = new Date(Date.UTC(y, m - 1, d));
    return dt.getUTCFullYear() === y && dt.getUTCMonth() === m - 1 && dt.getUTCDate() === d;
  }

  function calendarDateUTC(nowIso) {
    return (nowIso || isoNow()).slice(0, 10);
  }

  // Overdue = due strictly before the UTC calendar date of the clock, and the
  // issue is not resolved or dismissed (FR-021 / ADR-0003). Due today is not overdue.
  function isOverdue(issue, nowIso) {
    if (!issue || !issue.due || CLOSED[issue.state]) return false;
    return issue.due < calendarDateUTC(nowIso);
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
    if (!isObject(raw.evidence) || !validEvidenceID(raw.evidence.id)) {
      return { error: loc + ' evidence.id is missing or unversioned' };
    }
    if (raw.linkedEvidence != null && !isArray(raw.linkedEvidence)) {
      return { error: loc + ' linkedEvidence must be an array' };
    }
    if (raw.ignoredEvidence != null && !isArray(raw.ignoredEvidence)) {
      return { error: loc + ' ignoredEvidence must be an array' };
    }
    if (isArray(raw.linkedEvidence) && raw.linkedEvidence.length >= LIMITS.maxLinkedEvidence) {
      return { error: loc + ' has too many linked evidence refs' };
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

    function issueForEvidence(evidenceId) {
      var want = text(evidenceId);
      if (!want) return null;
      var found = null;
      all().some(function (issue) {
        return allEvidence(issue).some(function (ev) {
          if (ev && ev.id === want) {
            found = issue;
            return true;
          }
          return false;
        });
      });
      return found;
    }

    function refreshMatch(issue) {
      issue.evidenceMatched = allEvidence(issue).some(function (ev) {
        return !!(ev && byEvidence[ev.id]);
      });
      return issue;
    }

    function createFromEvent(event) {
      if (!event || !event.evidenceId) return { error: 'event is missing evidenceId' };
      var linked = issueForEvidence(event.evidenceId);
      if (linked) return { issue: linked, created: false };
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
        linkedEvidence: [],
        ignoredEvidence: [],
        evidenceMatched: true,
        extra: {}
      };
      remember(issue);
      dirtySinceExport = true;
      persist();
      return { issue: issue, created: true };
    }

    function linkEvidence(id, event) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      if (!event || !validEvidenceID(event.evidenceId)) return { error: 'event is missing evidenceId' };
      var owner = issueForEvidence(event.evidenceId);
      if (owner && owner.id === issue.id) return { issue: issue, linked: false };
      if (owner) return { error: 'evidence already linked to ' + owner.id };
      if (allEvidence(issue).length >= LIMITS.maxLinkedEvidence) {
        return { error: 'linked evidence limit reached' };
      }
      if (!isArray(issue.linkedEvidence)) issue.linkedEvidence = [];
      issue.linkedEvidence.push(evidenceSnapshot(event));
      issue.ignoredEvidence = (issue.ignoredEvidence || []).filter(function (eid) {
        return eid !== event.evidenceId;
      });
      refreshMatch(issue);
      touch(issue);
      persist();
      return { issue: issue, linked: true };
    }

    function unlinkEvidence(id, evidenceId) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      if (issue.evidence && issue.evidence.id === evidenceId) {
        return { error: 'cannot unlink originating evidence' };
      }
      if (!isArray(issue.linkedEvidence)) issue.linkedEvidence = [];
      var before = issue.linkedEvidence.length;
      issue.linkedEvidence = issue.linkedEvidence.filter(function (ev) {
        return ev.id !== evidenceId;
      });
      if (issue.linkedEvidence.length === before) {
        return { error: 'evidence is not linked to this issue' };
      }
      refreshMatch(issue);
      touch(issue);
      persist();
      return { issue: issue, unlinked: true };
    }

    function dismissCandidate(id, evidenceId) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      if (!validEvidenceID(evidenceId)) return { error: 'evidenceId is missing or unversioned' };
      var owner = issueForEvidence(evidenceId);
      if (owner && owner.id === issue.id) {
        return { error: 'cannot dismiss already linked evidence' };
      }
      if (!isArray(issue.ignoredEvidence)) issue.ignoredEvidence = [];
      if (issue.ignoredEvidence.indexOf(evidenceId) === -1) {
        issue.ignoredEvidence.push(evidenceId);
        touch(issue);
        persist();
      }
      return { issue: issue, dismissed: true };
    }

    function acknowledgeOccurrences(id, evidenceId) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      var updated = 0;
      function apply(ev) {
        if (!ev) return;
        if (evidenceId && ev.id !== evidenceId) return;
        var live = byEvidence[ev.id];
        if (!live) return;
        ev.occurrences = live.occurrences;
        ev.firstSeen = live.firstSeen || ev.firstSeen || null;
        ev.lastSeen = live.lastSeen || ev.lastSeen || null;
        ev.severity = live.severity;
        ev.sourceType = live.sourceType;
        updated += 1;
      }
      apply(issue.evidence);
      (issue.linkedEvidence || []).forEach(apply);
      if (!updated) {
        return { error: evidenceId ? 'evidence is not in this report' : 'no matching evidence in this report' };
      }
      touch(issue);
      persist();
      return { issue: issue, acknowledged: updated };
    }

    function listReviews() {
      var occurrenceUpdates = [];
      var candidates = [];
      var linkedIds = {};
      all().forEach(function (issue) {
        allEvidence(issue).forEach(function (ev) {
          if (ev && ev.id) linkedIds[ev.id] = issue.id;
        });
      });
      all().forEach(function (issue) {
        allEvidence(issue).forEach(function (ev) {
          var upd = occurrenceDelta(ev, byEvidence[ev.id]);
          if (upd) {
            occurrenceUpdates.push({
              issueId: issue.id,
              title: issue.title,
              state: issue.state,
              evidenceId: upd.evidenceId,
              previousOccurrences: upd.previousOccurrences,
              liveOccurrences: upd.liveOccurrences,
              newOccurrences: upd.newOccurrences,
              previousLastSeen: upd.previousLastSeen,
              liveLastSeen: upd.liveLastSeen
            });
          }
        });
        var keys = {};
        allEvidence(issue).forEach(function (ev) {
          var key = signatureKey(ev);
          if (key !== '\0') keys[key] = true;
        });
        var ignored = {};
        (issue.ignoredEvidence || []).forEach(function (eid) { ignored[eid] = true; });
        list.forEach(function (event) {
          if (!event || !validEvidenceID(event.evidenceId)) return;
          if (linkedIds[event.evidenceId]) return;
          if (ignored[event.evidenceId]) return;
          if (!keys[signatureKey(event)]) return;
          candidates.push({
            issueId: issue.id,
            title: issue.title,
            state: issue.state,
            evidenceId: event.evidenceId,
            signature: event.signature,
            instance: event.instance,
            file: event.file,
            line: event.line,
            occurrences: event.occurrences,
            lastSeen: event.lastSeen || null,
            severity: event.severity
          });
        });
      });
      return { occurrenceUpdates: occurrenceUpdates, candidates: candidates };
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

    function setOwner(id, owner) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      var value = text(owner).trim();
      if (value.length > LIMITS.maxOwner) {
        return { error: 'owner exceeds ' + LIMITS.maxOwner + ' characters' };
      }
      issue.owner = value || null;
      touch(issue);
      persist();
      return { issue: issue };
    }

    function setDue(id, due) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      if (due == null || due === '') {
        issue.due = null;
        touch(issue);
        persist();
        return { issue: issue };
      }
      if (!validDue(due)) return { error: 'due must be YYYY-MM-DD or empty' };
      issue.due = String(due);
      touch(issue);
      persist();
      return { issue: issue };
    }

    function setNotes(id, notes) {
      var issue = get(id);
      if (!issue) return { error: 'unknown issue' };
      if (notes == null || notes === '') {
        issue.notes = null;
        touch(issue);
        persist();
        return { issue: issue };
      }
      var value = text(notes);
      if (value.length > LIMITS.maxNotes) {
        return { error: 'notes exceed ' + LIMITS.maxNotes + ' characters' };
      }
      issue.notes = value;
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
        var refs = allEvidence(issue);
        if (criteria.text) {
          var q = String(criteria.text).toLowerCase();
          var hay = [
            issue.title, issue.id, issue.tags.join(' '), issue.owner || '',
            issue.notes || ''
          ];
          refs.forEach(function (ev) {
            hay.push(ev.file, ev.signature, ev.instance, ev.id);
          });
          if (hay.join(' ').toLowerCase().indexOf(q) === -1) return false;
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
        if (criteria.severity) {
          var sevOk = refs.some(function (ev) {
            var live = byEvidence[ev.id];
            return (live ? live.severity : ev.severity) === criteria.severity;
          });
          if (!sevOk) return false;
        }
        if (criteria.instance) {
          var instOk = refs.some(function (ev) {
            var liveInst = byEvidence[ev.id];
            return (liveInst ? liveInst.instance : ev.instance) === criteria.instance;
          });
          if (!instOk) return false;
        }
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
            evidence: clone(issue.evidence),
            linkedEvidence: clone(issue.linkedEvidence || []),
            ignoredEvidence: (issue.ignoredEvidence || []).slice()
          });
          return row;
        })
      };
    }

    function exportJSON() {
      return JSON.stringify(exportObject(), null, 2);
    }

    function hydrate(raw) {
      var tags = raw.tags.map(function (t) { return normalizeTag(t).value; });
      var evidence = normalizeEvidenceRef(raw.evidence) || Object.assign({}, raw.evidence);
      var linked = [];
      var seen = {};
      seen[evidence.id] = true;
      if (isArray(raw.linkedEvidence)) {
        raw.linkedEvidence.forEach(function (item) {
          var ref = normalizeEvidenceRef(item);
          if (!ref || seen[ref.id]) return;
          seen[ref.id] = true;
          linked.push(ref);
        });
      }
      var ignored = [];
      if (isArray(raw.ignoredEvidence)) {
        raw.ignoredEvidence.forEach(function (eid) {
          if (!validEvidenceID(eid) || ignored.indexOf(eid) !== -1) return;
          ignored.push(eid);
        });
      }
      var issue = {
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
        linkedEvidence: linked,
        ignoredEvidence: ignored,
        evidenceMatched: false,
        extra: extraFields(raw)
      };
      return refreshMatch(issue);
    }

    function importJSON(text, meta) {
      var parsed = parseExport(text);
      if (parsed.error) return { error: parsed.error, loaded: 0, invalid: [], unmatched: [] };
      var invalid = [];
      var unmatched = [];
      var loaded = 0;
      var claimed = {};
      all().forEach(function (issue) {
        allEvidence(issue).forEach(function (ev) {
          if (ev && ev.id) claimed[ev.id] = issue.id;
        });
      });
      parsed.data.issues.forEach(function (raw, index) {
        var checked = validateIssue(raw, index);
        if (checked.error) {
          invalid.push({ index: index, id: raw && raw.id, reason: checked.error });
          return;
        }
        var issue = hydrate(checked.issue);
        var previous = issues[issue.id];
        if (previous) {
          allEvidence(previous).forEach(function (ev) {
            if (ev && ev.id && claimed[ev.id] === previous.id) delete claimed[ev.id];
          });
        }
        var owner = null;
        allEvidence(issue).some(function (ev) {
          if (ev && ev.id && claimed[ev.id] && claimed[ev.id] !== issue.id) {
            owner = claimed[ev.id];
            return true;
          }
          return false;
        });
        if (owner) {
          if (previous) {
            allEvidence(previous).forEach(function (ev) {
              if (ev && ev.id) claimed[ev.id] = previous.id;
            });
          }
          invalid.push({
            index: index,
            id: issue.id,
            reason: 'evidence already linked to ' + owner
          });
          return;
        }
        allEvidence(issue).forEach(function (ev) {
          if (!ev || !ev.id) return;
          claimed[ev.id] = issue.id;
          if (!byEvidence[ev.id]) unmatched.push({ id: issue.id, evidenceId: ev.id });
        });
        remember(issue);
        loaded += 1;
      });
      if (!meta || !meta.fromLocal) dirtySinceExport = true;
      persist();
      var reviews = listReviews();
      return {
        loaded: loaded,
        invalid: invalid,
        unmatched: unmatched,
        candidates: reviews.candidates,
        occurrenceUpdates: reviews.occurrenceUpdates
      };
    }

    return {
      createFromEvent: createFromEvent,
      linkEvidence: linkEvidence,
      unlinkEvidence: unlinkEvidence,
      dismissCandidate: dismissCandidate,
      acknowledgeOccurrences: acknowledgeOccurrences,
      issueForEvidence: issueForEvidence,
      listReviews: listReviews,
      evidenceRefs: allEvidence,
      updateTitle: updateTitle,
      addTag: addTag,
      removeTag: removeTag,
      setFlagged: setFlagged,
      setState: setState,
      setOwner: setOwner,
      setDue: setDue,
      setNotes: setNotes,
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
    allEvidence: allEvidence,
    occurrenceDelta: occurrenceDelta,
    parseExport: parseExport,
    createStore: createStore,
    isOverdue: isOverdue,
    validDue: validDue
  };
});
