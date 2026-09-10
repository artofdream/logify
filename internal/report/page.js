(function () {
  'use strict';

  var events = Array.isArray(REPORT.events) ? REPORT.events : [];
  var warnings = Array.isArray(REPORT.warnings) ? REPORT.warnings.map(normalizeWarning) : [];
  function normalizeWarning(w) {
    if (!w || typeof w === 'string') {
      return { file: '', category: '', line: 0, lineEnd: 0, message: String(w || '') };
    }
    return w;
  }
  function warningLocation(w) {
    var file = w.file || '';
    if (w.line > 0 && w.lineEnd > w.line) return file + ':' + w.line + '-' + w.lineEnd;
    if (w.line > 0) return (file ? file + ':' : '') + w.line;
    return file || '(unknown path)';
  }
  var storageKey = 'logify-follow-up-v1:' + String(REPORT.root || '') + ':' + String(REPORT.generatedAt || '');
  var localStore = null;
  try { localStore = window.localStorage; } catch (err) { localStore = null; }
  var store = LogifyFollowUp.createStore(events, {
    storage: localStore,
    storageKey: storageKey,
    reportRoot: REPORT.root || '',
    generatedAt: REPORT.generatedAt || ''
  });

  function $(id) { return document.getElementById(id); }
  function el(tag, className, textValue) {
    var node = document.createElement(tag);
    if (className) node.className = className;
    if (textValue !== undefined) node.textContent = String(textValue);
    return node;
  }
  function clear(node) {
    while (node.firstChild) node.removeChild(node.firstChild);
  }
  function formatTime(value) {
    if (!value) return 'No timestamp';
    var d = new Date(value);
    if (isNaN(d.getTime())) return 'No timestamp';
    return d.toLocaleString();
  }
  function option(select, value, label) {
    var o = document.createElement('option');
    o.value = value;
    o.textContent = label;
    select.appendChild(o);
  }
  function fillUnique(select, values) {
    var seen = {};
    values.forEach(function (v) {
      if (!v || seen[v]) return;
      seen[v] = true;
    });
    Object.keys(seen).sort().forEach(function (v) { option(select, v, v); });
  }
  function digest(id) {
    return String(id || '').replace(/^evidence-v1-/, '').replace(/^issue-v1-/, '');
  }
  function evidenceAnchor(event) { return 'evidence-' + digest(event.evidenceId); }
  function issueAnchor(issue) { return 'issue-' + digest(issue.id); }
  var STATE_LABELS = {
    open: 'Open',
    investigating: 'Investigating',
    blocked: 'Blocked',
    resolved: 'Resolved',
    dismissed: 'Dismissed'
  };
  var pendingFocus = null;
  var detailFeedbackTimer = null;

  function stateLabel(state) {
    return STATE_LABELS[state] || String(state || '');
  }

  function showFeedback(id, message, isError) {
    var node = $(id);
    node.textContent = message || '';
    node.className = isError ? 'feedback error' : 'feedback';
  }

  function showDetailFeedback(message, isError) {
    if (detailFeedbackTimer) clearTimeout(detailFeedbackTimer);
    detailFeedbackTimer = setTimeout(function () {
      showFeedback('issue-feedback', message, isError);
    }, 400);
  }

  function rememberFocus(issueId, control) {
    pendingFocus = { issueId: issueId, control: control || '' };
  }

  function applyPendingFocus() {
    if (!pendingFocus) return;
    var card = document.getElementById(issueAnchor({ id: pendingFocus.issueId }));
    var control = pendingFocus.control;
    pendingFocus = null;
    if (!card) return;
    var target = control ? card.querySelector('[data-control="' + control + '"]') : null;
    if (!target) target = card.querySelector('[data-control="tag-input"]') || card;
    if (target && typeof target.focus === 'function') target.focus();
  }

  function switchView(name) {
    var timeline = name === 'timeline';
    $('view-timeline').hidden = !timeline;
    $('view-issues').hidden = timeline;
    $('tab-timeline').setAttribute('aria-selected', timeline ? 'true' : 'false');
    $('tab-issues').setAttribute('aria-selected', timeline ? 'false' : 'true');
    $('tab-timeline').tabIndex = timeline ? 0 : -1;
    $('tab-issues').tabIndex = timeline ? -1 : 0;
  }

  function renderStats() {
    var items = [
      [store.observedRecords(), 'observed records'],
      [store.eventGroupCount(), 'event groups'],
      [store.issueCount(), 'tracked issues'],
      [REPORT.filesScanned || 0, 'files scanned'],
      [REPORT.filesProcessed || 0, 'processed'],
      [REPORT.filesSkipped || 0, 'skipped'],
      [REPORT.filesFailed || 0, 'failed'],
      [warnings.length, 'warnings']
    ];
    var root = $('stats');
    clear(root);
    items.forEach(function (pair) {
      var box = el('div', 'stat');
      box.appendChild(el('b', '', pair[0]));
      box.appendChild(el('span', 'kind', pair[1]));
      root.appendChild(box);
    });
  }

  function renderStorage() {
    $('storage-notice').textContent = store.storageDescription() +
      ' Clearing local data removes issues from this browser copy; the log bundle is never modified.';
  }

  function eventMatches(event) {
    var q = $('q').value.toLowerCase();
    if (q && [event.message, event.file, event.instance].join(' ').toLowerCase().indexOf(q) === -1) return false;
    if ($('sev').value && event.severity !== $('sev').value) return false;
    if ($('inst').value && event.instance !== $('inst').value) return false;
    if ($('src').value && event.sourceType !== $('src').value) return false;
    return true;
  }

  function showEvidence(event) {
    switchView('timeline');
    ['q', 'sev', 'inst', 'src'].forEach(function (id) { $(id).value = ''; });
    renderTimeline();
    var target = document.getElementById(evidenceAnchor(event));
    if (target) {
      target.focus();
      target.scrollIntoView({ block: 'center' });
    }
  }

  function showIssue(issue) {
    switchView('issues');
    ['iq', 'istate', 'itags', 'iflag', 'iowner', 'isev', 'iinst'].forEach(function (id) { $(id).value = ''; });
    $('ioverdue').checked = false;
    renderIssues();
    var target = document.getElementById(issueAnchor(issue));
    if (target) {
      target.focus();
      target.scrollIntoView({ block: 'center' });
    }
  }

  function renderTimeline() {
    var root = $('timeline');
    clear(root);
    var visible = events.filter(eventMatches);
    if (!visible.length) {
      root.appendChild(el('div', 'empty', 'No matching events.'));
      return;
    }
    visible.forEach(function (event) {
      var article = el('article', 'event ' + (event.severity || 'UNKNOWN'));
      article.id = evidenceAnchor(event);
      article.tabIndex = -1;
      article.appendChild(el('div', 'when', event.hasTimestamp ? formatTime(event.timestamp) : 'No timestamp'));
      article.appendChild(el('div', 'sev', event.severity || 'UNKNOWN'));
      var src = el('div', 'src');
      src.appendChild(document.createTextNode(event.instance || ''));
      src.appendChild(document.createElement('br'));
      src.appendChild(document.createTextNode(event.sourceType || ''));
      article.appendChild(src);
      var body = el('div');
      body.appendChild(el('div', 'msg', event.message));
      var details = (event.file || '') + ':' + event.line + ' • ' + (event.signature || '');
      if ((event.occurrences || 0) > 1) {
        details += ' • ' + event.occurrences + ' occurrences; first ' + formatTime(event.firstSeen) + '; last ' + formatTime(event.lastSeen);
      }
      body.appendChild(el('div', 'detail', details));
      var actions = el('div', 'event-actions');
      var existing = store.issueForEvidence(event.evidenceId);
      var btn = el('button', '', existing ? 'Open issue' : 'Create issue');
      btn.type = 'button';
      btn.setAttribute('aria-label', (existing ? 'Open issue for ' : 'Create issue from ') + (event.file || 'event') + ' line ' + event.line);
      btn.addEventListener('click', function () {
        if (existing) {
          showIssue(existing);
          return;
        }
        var result = store.createFromEvent(event);
        if (result.error) {
          showFeedback('issue-feedback', result.error, true);
          return;
        }
        renderAll();
        showIssue(result.issue);
        showFeedback('issue-feedback', result.created ? 'Created ' + result.issue.id : 'Issue already exists; selected ' + result.issue.id);
      });
      actions.appendChild(btn);
      var picker = null;
      if (!existing && store.issueCount() > 0) {
        var linkBtn = el('button', '', 'Link to existing issue');
        linkBtn.type = 'button';
        linkBtn.setAttribute('aria-label', 'Link ' + (event.file || 'event') + ' line ' + event.line + ' to an existing issue');
        picker = el('div', 'link-picker');
        picker.hidden = true;
        var linkSelect = document.createElement('select');
        option(linkSelect, '', 'Choose an issue');
        store.list().forEach(function (issue) {
          option(linkSelect, issue.id, issue.title + ' (' + issue.id + ')');
        });
        var confirmLink = el('button', '', 'Link');
        confirmLink.type = 'button';
        confirmLink.disabled = true;
        linkSelect.addEventListener('change', function () {
          confirmLink.disabled = !linkSelect.value;
        });
        confirmLink.addEventListener('click', function () {
          if (!linkSelect.value) return;
          var result = store.linkEvidence(linkSelect.value, event);
          if (result.error) {
            showFeedback('issue-feedback', result.error, true);
            return;
          }
          renderAll();
          showIssue(result.issue);
          showFeedback('issue-feedback', 'Linked evidence to ' + result.issue.id + '; state remains ' + result.issue.state);
        });
        picker.appendChild(linkSelect);
        picker.appendChild(confirmLink);
        linkBtn.addEventListener('click', function () {
          picker.hidden = !picker.hidden;
        });
        actions.appendChild(linkBtn);
      }
      body.appendChild(actions);
      if (picker) body.appendChild(picker);
      article.appendChild(body);
      root.appendChild(article);
    });
  }

  function issueFilters() {
    return {
      text: $('iq').value,
      state: $('istate').value,
      tags: $('itags').value,
      flagged: $('iflag').value === 'flagged' ? true : undefined,
      owner: $('iowner').value,
      severity: $('isev').value,
      instance: $('iinst').value,
      overdue: $('ioverdue').checked ? true : undefined
    };
  }

  function field(list, label, value) {
    list.appendChild(el('dt', '', label));
    list.appendChild(el('dd', '', value));
  }

  function renderIssueCard(issue, reviewIndex) {
    var overdue = store.isOverdue(issue);
    var card = el('article', 'issue' + (issue.flagged ? ' flagged' : '') + (overdue ? ' overdue' : ''));
    card.id = issueAnchor(issue);
    card.tabIndex = -1;
    var heading = el('div', 'issue-heading');
    var left = el('div');
    left.appendChild(el('div', 'issue-id', issue.id));
    if (issue.flagged) {
      var flagBadge = el('span', 'flag-badge', 'Flagged');
      flagBadge.setAttribute('aria-label', 'Flag: flagged');
      left.appendChild(flagBadge);
    }
    var stateBadge = el('span', 'state-badge', 'State: ' + stateLabel(issue.state));
    stateBadge.setAttribute('aria-label', 'Workflow state: ' + stateLabel(issue.state));
    left.appendChild(stateBadge);
    if (overdue) left.appendChild(el('span', 'overdue-badge', 'Overdue'));
    if (!issue.evidenceMatched) left.appendChild(el('span', 'unmatched', 'No linked evidence in this report'));
    reviewIndex = reviewIndex || {};
    if (reviewIndex.newOcc && reviewIndex.newOcc[issue.id]) left.appendChild(el('span', 'new-occ-badge', 'New occurrences'));
    if (reviewIndex.candidates && reviewIndex.candidates[issue.id]) left.appendChild(el('span', 'candidate-badge', 'Signature match to review'));
    heading.appendChild(left);
    var tools = el('div', 'event-actions');
    var flagBtn = el('button', '', issue.flagged ? 'Unflag' : 'Flag for attention');
    flagBtn.type = 'button';
    flagBtn.setAttribute('data-control', 'flag');
    flagBtn.setAttribute('aria-pressed', issue.flagged ? 'true' : 'false');
    flagBtn.setAttribute('aria-label', (issue.flagged ? 'Unflag ' : 'Flag for attention ') + issue.id);
    flagBtn.addEventListener('click', function () {
      var next = !issue.flagged;
      store.setFlagged(issue.id, next);
      rememberFocus(issue.id, 'flag');
      renderAll();
      showFeedback('issue-feedback', (next ? 'Flagged ' : 'Unflagged ') + issue.id);
    });
    var evBtn = el('button', '', 'Show evidence');
    evBtn.type = 'button';
    evBtn.setAttribute('data-control', 'evidence');
    evBtn.disabled = !issue.evidenceMatched;
    evBtn.addEventListener('click', function () {
      var refs = store.evidenceRefs(issue);
      var live = null;
      refs.some(function (ref) {
        live = events.filter(function (e) { return e.evidenceId === ref.id; })[0];
        return !!live;
      });
      if (live) showEvidence(live);
    });
    tools.appendChild(flagBtn);
    tools.appendChild(evBtn);
    heading.appendChild(tools);
    card.appendChild(heading);

    var titleLabel = el('label', '', 'Title');
    var title = document.createElement('input');
    title.type = 'text';
    title.id = 'title-' + digest(issue.id);
    title.setAttribute('data-control', 'title');
    title.maxLength = LogifyFollowUp.LIMITS.maxTitle;
    title.value = issue.title;
    title.addEventListener('input', function () {
      store.updateTitle(issue.id, title.value);
      showDetailFeedback('Updated title for ' + issue.id);
    });
    titleLabel.appendChild(title);
    card.appendChild(titleLabel);

    var stateField = el('label', '', 'Workflow state');
    var state = document.createElement('select');
    state.id = 'state-' + digest(issue.id);
    state.setAttribute('data-control', 'state');
    state.setAttribute('aria-label', 'Workflow state for ' + issue.id);
    LogifyFollowUp.STATES.forEach(function (s) { option(state, s, stateLabel(s)); });
    state.value = issue.state;
    state.addEventListener('change', function () {
      store.setState(issue.id, state.value);
      rememberFocus(issue.id, 'state');
      renderAll();
      showFeedback('issue-feedback', 'State for ' + issue.id + ' is now ' + stateLabel(state.value));
    });
    stateField.appendChild(state);
    card.appendChild(stateField);

    var op = el('div', 'box');
    op.appendChild(el('h3', '', 'Operator metadata'));
    var tags = el('div', 'tag-list');
    if (!issue.tags.length) tags.appendChild(el('span', 'muted', 'No tags'));
    issue.tags.forEach(function (tag) {
      var chip = el('span', 'tag', tag);
      var rm = el('button', '', 'Remove');
      rm.type = 'button';
      rm.setAttribute('data-control', 'tag-remove');
      rm.setAttribute('aria-label', 'Remove tag ' + tag + ' from ' + issue.id);
      rm.addEventListener('click', function () {
        store.removeTag(issue.id, tag);
        rememberFocus(issue.id, 'tag-input');
        renderAll();
        showFeedback('issue-feedback', 'Removed tag from ' + issue.id);
      });
      chip.appendChild(rm);
      tags.appendChild(chip);
    });
    op.appendChild(tags);
    var add = el('div', 'tag-add');
    var tagLabel = el('label', '', 'New tag');
    var tagInput = document.createElement('input');
    tagInput.type = 'text';
    tagInput.id = 'tag-' + digest(issue.id);
    tagInput.setAttribute('data-control', 'tag-input');
    tagInput.setAttribute('autocomplete', 'off');
    tagInput.maxLength = LogifyFollowUp.LIMITS.maxTagLength;
    tagInput.placeholder = 'e.g. database';
    tagLabel.appendChild(tagInput);
    var addBtn = el('button', '', 'Add tag');
    addBtn.type = 'button';
    addBtn.setAttribute('data-control', 'tag-add');
    function submitTag() {
      var result = store.addTag(issue.id, tagInput.value);
      if (result.error) {
        showFeedback('issue-feedback', result.error, true);
        return;
      }
      rememberFocus(issue.id, 'tag-input');
      renderAll();
      showFeedback('issue-feedback', result.added ? 'Tagged ' + issue.id : 'Tag already present');
    }
    addBtn.addEventListener('click', submitTag);
    tagInput.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') {
        e.preventDefault();
        submitTag();
      }
    });
    add.appendChild(tagLabel);
    add.appendChild(addBtn);
    op.appendChild(add);

    var details = el('div', 'follow-up-fields');
    var ownerLabel = el('label', '', 'Owner');
    var ownerInput = document.createElement('input');
    ownerInput.type = 'text';
    ownerInput.id = 'owner-' + digest(issue.id);
    ownerInput.setAttribute('data-control', 'owner');
    ownerInput.maxLength = LogifyFollowUp.LIMITS.maxOwner;
    ownerInput.autocomplete = 'off';
    ownerInput.placeholder = 'Optional owner';
    ownerInput.value = issue.owner || '';
    ownerInput.addEventListener('input', function () {
      var result = store.setOwner(issue.id, ownerInput.value);
      if (result.error) {
        showDetailFeedback(result.error, true);
        return;
      }
      showDetailFeedback('Updated owner for ' + issue.id);
    });
    ownerLabel.appendChild(ownerInput);
    details.appendChild(ownerLabel);

    var dueLabel = el('label', '', 'Due date');
    var dueRow = el('div', 'due-row');
    var dueInput = document.createElement('input');
    dueInput.type = 'date';
    dueInput.id = 'due-' + digest(issue.id);
    dueInput.setAttribute('data-control', 'due');
    dueInput.value = issue.due || '';
    dueInput.addEventListener('change', function () {
      var result = store.setDue(issue.id, dueInput.value);
      if (result.error) {
        showFeedback('issue-feedback', result.error, true);
        return;
      }
      rememberFocus(issue.id, 'due');
      renderAll();
      showFeedback('issue-feedback', dueInput.value ? ('Due date for ' + issue.id + ' is ' + dueInput.value) : ('Cleared due date for ' + issue.id));
    });
    var clearDue = el('button', '', 'Clear due date');
    clearDue.type = 'button';
    clearDue.setAttribute('data-control', 'due-clear');
    clearDue.addEventListener('click', function () {
      var result = store.setDue(issue.id, null);
      if (result.error) {
        showFeedback('issue-feedback', result.error, true);
        return;
      }
      rememberFocus(issue.id, 'due');
      renderAll();
      showFeedback('issue-feedback', 'Cleared due date for ' + issue.id);
    });
    dueRow.appendChild(dueInput);
    dueRow.appendChild(clearDue);
    dueLabel.appendChild(dueRow);
    dueLabel.appendChild(el('span', 'hint', 'Overdue when before today (UTC calendar date) and the issue is not resolved or dismissed.'));
    details.appendChild(dueLabel);

    var notesLabel = el('label', '', 'Notes');
    var notes = document.createElement('textarea');
    notes.id = 'notes-' + digest(issue.id);
    notes.setAttribute('data-control', 'notes');
    notes.maxLength = LogifyFollowUp.LIMITS.maxNotes;
    notes.rows = 5;
    notes.placeholder = 'Optional investigation notes';
    notes.value = issue.notes || '';
    notes.addEventListener('input', function () {
      var result = store.setNotes(issue.id, notes.value);
      if (result.error) {
        showDetailFeedback(result.error, true);
        return;
      }
      showDetailFeedback('Updated notes for ' + issue.id);
    });
    notesLabel.appendChild(notes);
    details.appendChild(notesLabel);
    op.appendChild(details);

    var meta = el('dl', 'meta-grid');
    field(meta, 'Last modified', formatTime(issue.modifiedAt));
    op.appendChild(meta);
    card.appendChild(op);

    var ev = el('div', 'box');
    ev.appendChild(el('h3', '', 'Observed evidence'));
    store.evidenceRefs(issue).forEach(function (ref, idx) {
      var live = events.filter(function (e) { return e.evidenceId === ref.id; })[0];
      var item = el('div', 'evidence-item');
      var badges = el('div', 'event-actions');
      if (idx === 0) badges.appendChild(el('span', 'origin-badge', 'Originating'));
      if (!live) badges.appendChild(el('span', 'unmatched', 'Not in this report'));
      var delta = LogifyFollowUp.occurrenceDelta(ref, live);
      if (delta && (delta.newOccurrences > 0 || delta.liveLastSeen !== delta.previousLastSeen)) {
        var label = delta.newOccurrences > 0
          ? (delta.newOccurrences + ' new occurrence(s)')
          : 'Newer last-seen time';
        badges.appendChild(el('span', 'new-occ-badge', label));
      }
      item.appendChild(badges);
      var dl = el('dl', 'evidence');
      field(dl, 'Evidence ID', ref.id);
      field(dl, 'Signature', ref.signature);
      field(dl, 'Instance', ref.instance);
      field(dl, 'Source', (ref.file || '') + ':' + ref.line);
      field(dl, 'First seen', formatTime(live && live.firstSeen ? live.firstSeen : ref.firstSeen));
      field(dl, 'Last seen', formatTime(live && live.lastSeen ? live.lastSeen : ref.lastSeen));
      field(dl, 'Stored occurrences', ref.occurrences == null ? '—' : ref.occurrences);
      if (live) field(dl, 'Occurrences in this report', live.occurrences);
      field(dl, 'Severity', (live && live.severity) || ref.severity || '—');
      item.appendChild(dl);
      var row = el('div', 'event-actions');
      var show = el('button', '', 'Show evidence');
      show.type = 'button';
      show.disabled = !live;
      show.addEventListener('click', function () {
        if (live) showEvidence(live);
      });
      row.appendChild(show);
      if (idx > 0) {
        var unlink = el('button', '', 'Unlink');
        unlink.type = 'button';
        unlink.setAttribute('aria-label', 'Unlink evidence ' + ref.id);
        unlink.addEventListener('click', function () {
          var result = store.unlinkEvidence(issue.id, ref.id);
          if (result.error) {
            showFeedback('issue-feedback', result.error, true);
            return;
          }
          renderAll();
          showFeedback('issue-feedback', 'Unlinked evidence from ' + issue.id + '; state remains ' + issue.state);
        });
        row.appendChild(unlink);
      }
      if (delta) {
        var ack = el('button', '', 'Acknowledge new occurrences');
        ack.type = 'button';
        ack.addEventListener('click', function () {
          var result = store.acknowledgeOccurrences(issue.id, ref.id);
          if (result.error) {
            showFeedback('issue-feedback', result.error, true);
            return;
          }
          renderAll();
          showFeedback('issue-feedback', 'Acknowledged new occurrences on ' + issue.id + '; state remains ' + result.issue.state);
        });
        row.appendChild(ack);
      }
      item.appendChild(row);
      ev.appendChild(item);
    });
    card.appendChild(ev);
    return card;
  }

  function renderReviews() {
    var root = $('match-review');
    var list = $('match-review-list');
    if (!root || !list) return;
    clear(list);
    var reviews = store.listReviews();
    if (!reviews.occurrenceUpdates.length && !reviews.candidates.length) {
      root.hidden = true;
      return;
    }
    root.hidden = false;
    reviews.occurrenceUpdates.forEach(function (u) {
      var card = el('article', 'review-item');
      var lastSeenOnly = !(u.newOccurrences > 0);
      card.appendChild(el('div', 'review-title', lastSeenOnly
        ? 'Newer last-seen time on ' + u.issueId
        : 'Newly observed occurrences on ' + u.issueId));
      var change = lastSeenOnly
        ? formatTime(u.previousLastSeen) + ' → ' + formatTime(u.liveLastSeen)
        : u.previousOccurrences + ' → ' + u.liveOccurrences;
      var detail = 'Evidence ' + u.evidenceId + ': ' + change +
        ' stored vs this report. Issue state remains ' + u.state + ' until you change it.';
      card.appendChild(el('div', 'detail', detail));
      var ack = el('button', '', 'Acknowledge');
      ack.type = 'button';
      ack.addEventListener('click', function () {
        var result = store.acknowledgeOccurrences(u.issueId, u.evidenceId);
        if (result.error) {
          showFeedback('issue-feedback', result.error, true);
          return;
        }
        renderAll();
        showFeedback('issue-feedback', 'Acknowledged new occurrences on ' + u.issueId + '; state remains ' + result.issue.state);
      });
      card.appendChild(ack);
      list.appendChild(card);
    });
    reviews.candidates.forEach(function (c) {
      var card = el('article', 'review-item');
      card.appendChild(el('div', 'review-title', 'Signature match for ' + c.issueId));
      var detail = (c.signature || '') + ' on ' + (c.file || '') + ':' + c.line +
        ' (' + (c.occurrences || 1) + ' occurrence(s); instance ' + (c.instance || '') +
        '). Automatic match has not changed issue state (' + c.state + ').';
      card.appendChild(el('div', 'detail', detail));
      var actions = el('div', 'event-actions');
      var link = el('button', '', 'Link to issue');
      link.type = 'button';
      link.addEventListener('click', function () {
        var live = events.filter(function (e) { return e.evidenceId === c.evidenceId; })[0];
        var result = store.linkEvidence(c.issueId, live);
        if (result.error) {
          showFeedback('issue-feedback', result.error, true);
          return;
        }
        renderAll();
        showFeedback('issue-feedback', 'Linked recurring evidence to ' + c.issueId + '; state remains ' + result.issue.state);
      });
      var dismiss = el('button', '', 'Dismiss');
      dismiss.type = 'button';
      dismiss.addEventListener('click', function () {
        var result = store.dismissCandidate(c.issueId, c.evidenceId);
        if (result.error) {
          showFeedback('issue-feedback', result.error, true);
          return;
        }
        renderAll();
        showFeedback('issue-feedback', 'Dismissed signature match for ' + c.issueId + '; state remains ' + result.issue.state);
      });
      actions.appendChild(link);
      actions.appendChild(dismiss);
      card.appendChild(actions);
      list.appendChild(card);
    });
  }

  function renderIssues(opts) {
    opts = opts || {};
    var root = $('issues');
    clear(root);
    var visible = store.filter(issueFilters());
    var total = store.issueCount();
    var summary = $('issue-summary');
    var summaryText;
    if (!total) {
      summaryText = 'No issues yet. Create one from a timeline event or group.';
    } else {
      summaryText = 'Showing ' + visible.length + ' of ' + total + ' issue(s).';
    }
    if (summary) summary.textContent = summaryText;
    if (!total) {
      root.appendChild(el('div', 'empty', 'No issues yet. Create one from a timeline event or group.'));
      return;
    }
    if (!visible.length) {
      root.appendChild(el('div', 'empty', 'No issues match the current filters.'));
      if (opts.announceCount) showFeedback('issue-feedback', summaryText);
      return;
    }
    var reviews = store.listReviews();
    var reviewIndex = { newOcc: {}, candidates: {} };
    reviews.occurrenceUpdates.forEach(function (u) { reviewIndex.newOcc[u.issueId] = true; });
    reviews.candidates.forEach(function (c) { reviewIndex.candidates[c.issueId] = true; });
    visible.forEach(function (issue) { root.appendChild(renderIssueCard(issue, reviewIndex)); });
    if (opts.announceCount) showFeedback('issue-feedback', summaryText);
    applyPendingFocus();
  }

  function renderWarnings() {
    var section = $('warnings');
    var list = $('warning-list');
    if (!section || !list) return;
    clear(list);
    if (!warnings.length) {
      section.hidden = true;
      return;
    }
    section.hidden = false;
    warnings.forEach(function (w) {
      var item = el('article', 'warning');
      item.appendChild(el('div', 'warn-cat', w.category || 'unknown'));
      item.appendChild(el('div', 'warn-loc', warningLocation(w)));
      item.appendChild(el('div', 'warn-msg', w.message || ''));
      list.appendChild(item);
    });
  }

  function renderAll() {
    renderStats();
    renderWarnings();
    renderStorage();
    renderTimeline();
    renderReviews();
    renderIssues();
  }

  function exportFollowUp() {
    var blob = new Blob([store.exportJSON()], { type: 'application/json' });
    store.markExported();
    var url = URL.createObjectURL(blob);
    var a = document.createElement('a');
    a.href = url;
    a.download = 'logify-follow-up.json';
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
    showFeedback('storage-feedback', 'Exported ' + store.issueCount() + ' issue(s) to logify-follow-up.json');
  }

  function importFollowUp(file) {
    if (!file) return;
    var reader = new FileReader();
    reader.onload = function () {
      var result = store.importJSON(String(reader.result || ''));
      renderAll();
      if (result.error) {
        showFeedback('storage-feedback', result.error, true);
        return;
      }
      var parts = ['Imported ' + result.loaded + ' issue(s)'];
      if (result.invalid.length) parts.push(result.invalid.length + ' invalid record(s) skipped');
      if (result.unmatched.length) parts.push(result.unmatched.length + ' unmatched evidence id(s)');
      if (result.candidates && result.candidates.length) {
        parts.push(result.candidates.length + ' signature match(es) to review');
      }
      if (result.occurrenceUpdates && result.occurrenceUpdates.length) {
        parts.push(result.occurrenceUpdates.length + ' newly observed occurrence group(s)');
      }
      parts.push('Issue states were not changed');
      showFeedback('storage-feedback', parts.join('. '), result.invalid.length > 0);
      showFeedback('issue-feedback', '');
      switchView('issues');
    };
    reader.readAsText(file);
  }

  fillUnique($('sev'), events.map(function (e) { return e.severity; }));
  fillUnique($('inst'), events.map(function (e) { return e.instance; }));
  fillUnique($('src'), events.map(function (e) { return e.sourceType; }));
  fillUnique($('isev'), events.map(function (e) { return e.severity; }));
  fillUnique($('iinst'), events.map(function (e) { return e.instance; }));
  LogifyFollowUp.STATES.forEach(function (s) { option($('istate'), s, stateLabel(s)); });

  $('meta').textContent = (REPORT.root || '.') + ' • generated ' + formatTime(REPORT.generatedAt);
  var redaction = REPORT.redaction || {};
  if (redaction.enabled) {
    $('redaction-status').textContent = ' This report applied ' + (redaction.ruleCount || 0) +
      ' rule(s) (' + (redaction.replacements || 0) + ' replacement(s)).';
  } else {
    $('redaction-status').textContent = ' No redaction rules were applied.';
  }
  ['q', 'sev', 'inst', 'src'].forEach(function (id) {
    $(id).addEventListener('input', renderTimeline);
  });
  ['iq', 'istate', 'itags', 'iflag', 'iowner', 'isev', 'iinst', 'ioverdue'].forEach(function (id) {
    $(id).addEventListener('input', function () { renderIssues({ announceCount: true }); });
    $(id).addEventListener('change', function () { renderIssues({ announceCount: true }); });
  });
  $('tab-timeline').addEventListener('click', function () { switchView('timeline'); });
  $('tab-issues').addEventListener('click', function () { switchView('issues'); });
  [$('tab-timeline'), $('tab-issues')].forEach(function (tab, index, tabs) {
    tab.addEventListener('keydown', function (e) {
      var next = index;
      if (e.key === 'ArrowRight' || e.key === 'ArrowDown') next = (index + 1) % tabs.length;
      else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') next = (index - 1 + tabs.length) % tabs.length;
      else if (e.key === 'Home') next = 0;
      else if (e.key === 'End') next = tabs.length - 1;
      else return;
      e.preventDefault();
      tabs[next].focus();
      tabs[next].click();
    });
  });
  $('btn-export').addEventListener('click', exportFollowUp);
  $('btn-import').addEventListener('click', function () { $('import-file').click(); });
  $('import-file').addEventListener('change', function () {
    importFollowUp($('import-file').files && $('import-file').files[0]);
    $('import-file').value = '';
  });
  $('btn-clear').addEventListener('click', function () {
    if (!window.confirm('Clear local follow-up data for this report? Export first if you need a portable copy. The log bundle is not changed.')) return;
    store.clearLocal();
    renderAll();
    showFeedback('storage-feedback', 'Cleared local follow-up data');
  });
  window.addEventListener('beforeunload', function (event) {
    if (!store.hasUnexportedChanges() || !store.issueCount()) return;
    event.preventDefault();
    event.returnValue = '';
  });

  var loaded = store.loadLocal();
  renderAll();
  if (loaded && loaded.loaded) {
    var extra = [];
    if (loaded.candidates && loaded.candidates.length) extra.push(loaded.candidates.length + ' signature match(es) to review');
    if (loaded.occurrenceUpdates && loaded.occurrenceUpdates.length) extra.push(loaded.occurrenceUpdates.length + ' newly observed occurrence group(s)');
    showFeedback('storage-feedback', 'Restored ' + loaded.loaded + ' issue(s) from local storage' + (extra.length ? '. ' + extra.join('. ') : ''));
  }
})();
