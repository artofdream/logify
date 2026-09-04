(function () {
  'use strict';

  var events = Array.isArray(REPORT.events) ? REPORT.events : [];
  var warnings = Array.isArray(REPORT.warnings) ? REPORT.warnings : [];
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
  function showFeedback(id, message, isError) {
    var node = $(id);
    node.textContent = message || '';
    node.className = isError ? 'feedback error' : 'feedback';
  }
  function switchView(name) {
    var timeline = name === 'timeline';
    $('view-timeline').hidden = !timeline;
    $('view-issues').hidden = timeline;
    $('tab-timeline').setAttribute('aria-selected', timeline ? 'true' : 'false');
    $('tab-issues').setAttribute('aria-selected', timeline ? 'false' : 'true');
  }

  function renderStats() {
    var items = [
      [store.observedRecords(), 'observed records'],
      [store.eventGroupCount(), 'event groups'],
      [store.issueCount(), 'tracked issues'],
      [REPORT.filesScanned || 0, 'files'],
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
      var existing = store.get(LogifyFollowUp.issueIDFromEvidence(event.evidenceId));
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
      body.appendChild(actions);
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

  function renderIssueCard(issue) {
    var card = el('article', 'issue' + (issue.flagged ? ' flagged' : ''));
    card.id = issueAnchor(issue);
    card.tabIndex = -1;
    var heading = el('div', 'issue-heading');
    var left = el('div');
    left.appendChild(el('div', 'issue-id', issue.id));
    if (issue.flagged) left.appendChild(el('span', 'flag-badge', 'Flagged'));
    left.appendChild(el('span', 'state-badge', issue.state));
    if (!issue.evidenceMatched) left.appendChild(el('span', 'unmatched', 'Evidence not in this report'));
    heading.appendChild(left);
    var tools = el('div', 'event-actions');
    var flagBtn = el('button', '', issue.flagged ? 'Unflag' : 'Flag for attention');
    flagBtn.type = 'button';
    flagBtn.setAttribute('aria-pressed', issue.flagged ? 'true' : 'false');
    flagBtn.addEventListener('click', function () {
      var next = !issue.flagged;
      store.setFlagged(issue.id, next);
      renderAll();
      showFeedback('issue-feedback', (next ? 'Flagged ' : 'Unflagged ') + issue.id);
    });
    var evBtn = el('button', '', 'Show evidence');
    evBtn.type = 'button';
    evBtn.disabled = !issue.evidenceMatched;
    evBtn.addEventListener('click', function () {
      var live = events.filter(function (e) { return e.evidenceId === issue.evidence.id; })[0];
      if (live) showEvidence(live);
    });
    tools.appendChild(flagBtn);
    tools.appendChild(evBtn);
    heading.appendChild(tools);
    card.appendChild(heading);

    var titleLabel = el('label', '', 'Title');
    var title = document.createElement('input');
    title.type = 'text';
    title.maxLength = LogifyFollowUp.LIMITS.maxTitle;
    title.value = issue.title;
    title.addEventListener('input', function () {
      store.updateTitle(issue.id, title.value);
      showFeedback('issue-feedback', 'Updated title for ' + issue.id);
    });
    titleLabel.appendChild(title);
    card.appendChild(titleLabel);

    var stateLabel = el('label', '', 'Workflow state');
    var state = document.createElement('select');
    LogifyFollowUp.STATES.forEach(function (s) { option(state, s, s); });
    state.value = issue.state;
    state.addEventListener('change', function () {
      store.setState(issue.id, state.value);
      renderAll();
      showFeedback('issue-feedback', 'State for ' + issue.id + ' is now ' + state.value);
    });
    stateLabel.appendChild(state);
    card.appendChild(stateLabel);

    var op = el('div', 'box');
    op.appendChild(el('h3', '', 'Operator metadata'));
    var tags = el('div', 'tag-list');
    if (!issue.tags.length) tags.appendChild(el('span', 'muted', 'No tags'));
    issue.tags.forEach(function (tag) {
      var chip = el('span', 'tag', tag);
      var rm = el('button', '', 'Remove');
      rm.type = 'button';
      rm.setAttribute('aria-label', 'Remove tag ' + tag);
      rm.addEventListener('click', function () {
        store.removeTag(issue.id, tag);
        renderAll();
        showFeedback('issue-feedback', 'Removed tag from ' + issue.id);
      });
      chip.appendChild(rm);
      tags.appendChild(chip);
    });
    op.appendChild(tags);
    var add = el('div', 'tag-add');
    var tagInput = document.createElement('input');
    tagInput.type = 'text';
    tagInput.maxLength = LogifyFollowUp.LIMITS.maxTagLength;
    tagInput.placeholder = 'Add tag';
    var addBtn = el('button', '', 'Add tag');
    addBtn.type = 'button';
    addBtn.addEventListener('click', function () {
      var result = store.addTag(issue.id, tagInput.value);
      if (result.error) {
        showFeedback('issue-feedback', result.error, true);
        return;
      }
      renderAll();
      showFeedback('issue-feedback', result.added ? 'Tagged ' + issue.id : 'Tag already present');
    });
    add.appendChild(tagInput);
    add.appendChild(addBtn);
    op.appendChild(add);
    var meta = el('dl', 'meta-grid');
    field(meta, 'Owner', issue.owner || '—');
    field(meta, 'Due', issue.due ? (store.isOverdue(issue) ? issue.due + ' (overdue)' : issue.due) : '—');
    field(meta, 'Notes', issue.notes || '—');
    field(meta, 'Last modified', formatTime(issue.modifiedAt));
    op.appendChild(meta);
    card.appendChild(op);

    var ev = el('div', 'box');
    ev.appendChild(el('h3', '', 'Observed evidence'));
    var dl = el('dl', 'evidence');
    field(dl, 'Evidence ID', issue.evidence.id);
    field(dl, 'Signature', issue.evidence.signature);
    field(dl, 'Instance', issue.evidence.instance);
    field(dl, 'Source', (issue.evidence.file || '') + ':' + issue.evidence.line);
    field(dl, 'First seen', formatTime(issue.evidence.firstSeen));
    field(dl, 'Last seen', formatTime(issue.evidence.lastSeen));
    field(dl, 'Occurrences', issue.evidence.occurrences);
    field(dl, 'Severity', issue.evidence.severity || '—');
    ev.appendChild(dl);
    card.appendChild(ev);
    return card;
  }

  function renderIssues() {
    var root = $('issues');
    clear(root);
    var visible = store.filter(issueFilters());
    if (!store.issueCount()) {
      root.appendChild(el('div', 'empty', 'No issues yet. Create one from a timeline event or group.'));
      return;
    }
    if (!visible.length) {
      root.appendChild(el('div', 'empty', 'No issues match the current filters.'));
      return;
    }
    visible.forEach(function (issue) { root.appendChild(renderIssueCard(issue)); });
  }

  function renderAll() {
    renderStats();
    renderStorage();
    renderTimeline();
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
      showFeedback('storage-feedback', parts.join('. '), result.invalid.length > 0);
      switchView('issues');
    };
    reader.readAsText(file);
  }

  fillUnique($('sev'), events.map(function (e) { return e.severity; }));
  fillUnique($('inst'), events.map(function (e) { return e.instance; }));
  fillUnique($('src'), events.map(function (e) { return e.sourceType; }));
  fillUnique($('isev'), events.map(function (e) { return e.severity; }));
  fillUnique($('iinst'), events.map(function (e) { return e.instance; }));
  LogifyFollowUp.STATES.forEach(function (s) { option($('istate'), s, s); });

  $('meta').textContent = (REPORT.root || '.') + ' • generated ' + formatTime(REPORT.generatedAt);
  ['q', 'sev', 'inst', 'src'].forEach(function (id) {
    $(id).addEventListener('input', renderTimeline);
  });
  ['iq', 'istate', 'itags', 'iflag', 'iowner', 'isev', 'iinst', 'ioverdue'].forEach(function (id) {
    $(id).addEventListener('input', renderIssues);
    $(id).addEventListener('change', renderIssues);
  });
  $('tab-timeline').addEventListener('click', function () { switchView('timeline'); });
  $('tab-issues').addEventListener('click', function () { switchView('issues'); });
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
    showFeedback('storage-feedback', 'Restored ' + loaded.loaded + ' issue(s) from local storage');
  }
})();
