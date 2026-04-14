(function () {
  'use strict';

  var userMenu     = document.getElementById('user-menu');
  var userTrigger  = document.getElementById('user-menu-trigger');

  if (userTrigger && userMenu) {
    userTrigger.addEventListener('click', function (e) {
      e.stopPropagation();
      var isOpen = userMenu.classList.toggle('open');
      userTrigger.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    });

    document.addEventListener('click', function (e) {
      if (!userMenu.contains(e.target)) {
        userMenu.classList.remove('open');
        userTrigger.setAttribute('aria-expanded', 'false');
      }
    });

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && userMenu.classList.contains('open')) {
        userMenu.classList.remove('open');
        userTrigger.setAttribute('aria-expanded', 'false');
        userTrigger.focus();
      }
    });
  }

  document.querySelectorAll('.alert').forEach(function (el) {
    setTimeout(function () {
      el.style.transition = 'opacity 0.4s ease';
      el.style.opacity = '0';
      setTimeout(function () { el.remove(); }, 400);
    }, 4000);
  });

  var modalOverlay   = document.getElementById('modal-run-test');
  var btnRunTest     = document.getElementById('btn-run-test');
  var btnStopTest    = document.getElementById('btn-stop-test');
  var btnModalClose  = document.getElementById('btn-modal-close');
  var btnModalCancel = document.getElementById('btn-modal-cancel');
  var btnModalSubmit = document.getElementById('btn-modal-submit');
  var inputApiUrl    = document.getElementById('input-api-doc-url');
  var runErrorEl     = document.getElementById('run-error');
  var inputSource    = document.getElementById('input-source');

  function openModal() {
    if (!modalOverlay) return;
    modalOverlay.classList.add('open');
    if (inputApiUrl) inputApiUrl.focus();
    showRunError('');
  }

  function closeModal() {
    if (!modalOverlay) return;
    modalOverlay.classList.remove('open');
    modalOverlay.querySelectorAll('input:not([type=file]), select').forEach(function (el) {
      if (el.defaultValue !== undefined) el.value = el.defaultValue;
    });
    showRunError('');
  }

  function showRunError(msg) {
    if (!runErrorEl) return;
    if (msg) {
      runErrorEl.textContent = msg;
      runErrorEl.style.display = 'block';
    } else {
      runErrorEl.textContent = '';
      runErrorEl.style.display = 'none';
    }
  }

  var sourceTabs = document.querySelectorAll('.source-tab:not([disabled])');
  sourceTabs.forEach(function (tab) {
    tab.addEventListener('click', function () {
      sourceTabs.forEach(function (t) {
        t.classList.remove('active');
        t.setAttribute('aria-selected', 'false');
      });
      tab.classList.add('active');
      tab.setAttribute('aria-selected', 'true');

      var source = tab.dataset.source;
      if (inputSource) inputSource.value = source;

      document.querySelectorAll('.source-panel').forEach(function (panel) {
        panel.hidden = true;
      });
      var activePanel = document.getElementById('panel-' + source);
      if (activePanel) activePanel.hidden = false;
    });
  });

  if (btnRunTest)     btnRunTest.addEventListener('click', openModal);
  if (btnModalClose)  btnModalClose.addEventListener('click', closeModal);
  if (btnModalCancel) btnModalCancel.addEventListener('click', closeModal);

  if (modalOverlay) {
    modalOverlay.addEventListener('click', function (e) {
      if (e.target === modalOverlay) closeModal();
    });
  }

  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && modalOverlay && modalOverlay.classList.contains('open')) {
      closeModal();
    }
  });

  if (btnModalSubmit) {
    btnModalSubmit.addEventListener('click', function () {
      var source = inputSource ? inputSource.value : 'openapi';

      if (source === 'openapi') {
        if (!inputApiUrl || !inputApiUrl.value.trim()) {
          if (inputApiUrl) inputApiUrl.focus();
          showRunError('API Doc URL is required.');
          return;
        }
      }

      var formData = new FormData();
      formData.append('source', source);

      if (inputApiUrl && inputApiUrl.value.trim()) formData.append('spec_url', inputApiUrl.value.trim());

      var token = document.getElementById('input-jwt-token');
      if (token && token.value.trim()) formData.append('token', token.value.trim());

      var workers = document.getElementById('input-workers');
      if (workers && workers.value) formData.append('workers', workers.value);

      var duration = document.getElementById('input-duration');
      if (duration && duration.value.trim()) formData.append('duration', duration.value.trim());

      var profile = document.getElementById('input-profile');
      if (profile && profile.value) formData.append('profile', profile.value);

      if (source === 'postman') {
        var fileInput = document.getElementById('input-postman-file');
        if (fileInput && fileInput.files.length > 0) {
          formData.append('postman_file', fileInput.files[0]);
        }
      }

      btnModalSubmit.disabled = true;
      btnModalSubmit.textContent = 'Starting…';

      fetch('/api/run', { method: 'POST', body: formData })
        .then(function (res) { return res.json().then(function (body) { return { ok: res.ok, body: body }; }); })
        .then(function (r) {
          if (!r.ok) {
            showRunError(r.body.error || 'Failed to start test.');
          } else {
            closeModal();
            setTimeout(function () { window.location.reload(); }, 500);
          }
        })
        .catch(function () { showRunError('Network error. Please try again.'); })
        .finally(function () {
          btnModalSubmit.disabled = false;
          btnModalSubmit.textContent = 'Start Test';
        });
    });
  }

  if (btnStopTest) {
    btnStopTest.addEventListener('click', function () {
      btnStopTest.disabled = true;
      btnStopTest.textContent = 'Stopping…';
      fetch('/api/run', { method: 'DELETE' })
        .then(function () { setTimeout(function () { window.location.reload(); }, 1500); })
        .catch(function () {
          btnStopTest.disabled = false;
          btnStopTest.textContent = 'Stop Test';
        });
    });
  }

  var statActiveEl = document.querySelector('[data-stat="active-tests"]');
  var statStatusEl = document.querySelector('[data-stat="last-status"]');

  if (statActiveEl || statStatusEl) {
    setInterval(function () {
      fetch('/api/status')
        .then(function (r) { return r.json(); })
        .then(function (data) {
          if (statActiveEl) statActiveEl.textContent = data.ActiveTests;
          if (statStatusEl) {
            statStatusEl.textContent = data.LastStatus;
            statStatusEl.className = 'status-badge status-' + data.LastStatus;
          }
        })
        .catch(function () {});
    }, 3000);
  }

  var logOutput = document.getElementById('log-output');
  if (!logOutput) return;

  var MAX_ENTRIES = 500;
  var autoScroll  = true;

  logOutput.addEventListener('scroll', function () {
    autoScroll = logOutput.scrollTop + logOutput.clientHeight >= logOutput.scrollHeight - 40;
  });

  var clearBtn = document.getElementById('btn-clear-log');
  if (clearBtn) {
    clearBtn.addEventListener('click', function () {
      fetch('/api/logs', { method: 'DELETE' })
        .then(function (res) {
          if (res.ok) logOutput.innerHTML = '';
        })
        .catch(function () {});
    });
  }

  function escapeHtml(str) {
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  function appendEntry(entry) {
    var empty = logOutput.querySelector('.log-empty');
    if (empty) empty.remove();

    var line = document.createElement('div');
    line.className = 'log-line log-' + entry.level.toLowerCase();
    line.innerHTML =
      '<span class="log-time">'  + entry.timestamp           + '</span>' +
      '<span class="log-level">' + '[' + entry.level + ']'   + '</span>' +
      '<span class="log-msg">'   + escapeHtml(entry.message) + '</span>';

    logOutput.appendChild(line);

    while (logOutput.children.length > MAX_ENTRIES) {
      logOutput.removeChild(logOutput.firstChild);
    }

    if (autoScroll) {
      logOutput.scrollTop = logOutput.scrollHeight;
    }
  }

  var source = new EventSource('/api/logs/stream');

  source.onmessage = function (e) {
    try { appendEntry(JSON.parse(e.data)); } catch (_) {}
  };

  source.onerror = function () {
    appendEntry({
      timestamp: new Date().toTimeString().slice(0, 8),
      level:     'WARN',
      message:   'Log stream disconnected. Reconnecting...',
    });
  };
})();

(function () {
  'use strict';

  var tbody      = document.getElementById('threshold-tbody');
  var btnAdd     = document.getElementById('btn-add-threshold');
  var btnSave    = document.getElementById('btn-save-thresholds');
  var saveStatus = document.getElementById('save-status');

  if (!tbody) return;

  var METRICS = [
    { value: 'p50_latency',    label: 'P50 Latency',    unit: 'ms'    },
    { value: 'p90_latency',    label: 'P90 Latency',    unit: 'ms'    },
    { value: 'p95_latency',    label: 'P95 Latency',    unit: 'ms'    },
    { value: 'p99_latency',    label: 'P99 Latency',    unit: 'ms'    },
    { value: 'avg_latency',    label: 'Avg Latency',    unit: 'ms'    },
    { value: 'max_latency',    label: 'Max Latency',    unit: 'ms'    },
    { value: 'error_rate',     label: 'Error Rate',     unit: '%'     },
    { value: 'success_rate',   label: 'Success Rate',   unit: '%'     },
    { value: 'rps',            label: 'RPS',            unit: 'req/s' },
    { value: 'total_requests', label: 'Total Requests', unit: ''      },
    { value: 'total_errors',   label: 'Total Errors',   unit: ''      },
  ];

  var OPERATORS = [
    { value: 'less_than',             label: '<' },
    { value: 'less_than_or_equal',    label: '≤' },
    { value: 'greater_than',          label: '>' },
    { value: 'greater_than_or_equal', label: '≥' },
    { value: 'equal',                 label: '=' },
  ];

  function unitForMetric(metric) {
    var m = METRICS.find(function (x) { return x.value === metric; });
    return m ? m.unit : '';
  }

  function buildSelect(options, selectedValue, className) {
    var sel = document.createElement('select');
    sel.className = className;
    options.forEach(function (opt) {
      var el = document.createElement('option');
      el.value = opt.value;
      el.textContent = opt.label;
      if (opt.value === selectedValue) el.selected = true;
      sel.appendChild(el);
    });
    return sel;
  }

  function createThresholdRow(assertion) {
    var tr = document.createElement('tr');
    tr.className = 'threshold-row' + (assertion.enabled ? '' : ' disabled-row');

    var tdEnabled = document.createElement('td');
    tdEnabled.className = 'td-enabled';
    var checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.className = 'threshold-checkbox';
    checkbox.checked = assertion.enabled;
    checkbox.addEventListener('change', function () {
      tr.classList.toggle('disabled-row', !checkbox.checked);
    });
    tdEnabled.appendChild(checkbox);

    var tdMetric = document.createElement('td');
    var metricSel = buildSelect(METRICS, assertion.metric, 'threshold-select threshold-metric');
    metricSel.addEventListener('change', function () {
      unitCell.textContent = unitForMetric(metricSel.value);
    });
    tdMetric.appendChild(metricSel);

    var tdOperator = document.createElement('td');
    tdOperator.appendChild(buildSelect(OPERATORS, assertion.operator, 'threshold-select threshold-operator'));

    var tdValue = document.createElement('td');
    var valueInput = document.createElement('input');
    valueInput.type = 'number';
    valueInput.className = 'threshold-input';
    valueInput.value = assertion.value;
    valueInput.min = '0';
    valueInput.step = 'any';
    tdValue.appendChild(valueInput);

    var unitCell = document.createElement('td');
    unitCell.className = 'threshold-unit';
    unitCell.textContent = unitForMetric(assertion.metric);

    var tdRemove = document.createElement('td');
    var removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'threshold-remove';
    removeBtn.title = 'Remove';
    removeBtn.textContent = '×';
    removeBtn.addEventListener('click', function () { tr.remove(); });
    tdRemove.appendChild(removeBtn);

    tr.appendChild(tdEnabled);
    tr.appendChild(tdMetric);
    tr.appendChild(tdOperator);
    tr.appendChild(tdValue);
    tr.appendChild(unitCell);
    tr.appendChild(tdRemove);

    return tr;
  }

  function renderThresholds(assertions) {
    tbody.innerHTML = '';
    if (!assertions || assertions.length === 0) {
      var emptyRow = document.createElement('tr');
      emptyRow.className = 'threshold-empty-row';
      var emptyCell = document.createElement('td');
      emptyCell.colSpan = 6;
      emptyCell.textContent = 'No thresholds configured. Click "+ Add Threshold" to add one.';
      emptyRow.appendChild(emptyCell);
      tbody.appendChild(emptyRow);
      return;
    }
    assertions.forEach(function (a) {
      tbody.appendChild(createThresholdRow(a));
    });
  }

  function collectThresholds() {
    var rows = tbody.querySelectorAll('.threshold-row');
    var result = [];
    rows.forEach(function (tr) {
      var enabled  = tr.querySelector('.threshold-checkbox').checked;
      var metric   = tr.querySelector('.threshold-metric').value;
      var operator = tr.querySelector('.threshold-operator').value;
      var value    = parseFloat(tr.querySelector('.threshold-input').value) || 0;
      result.push({ metric: metric, operator: operator, value: value, enabled: enabled });
    });
    return result;
  }

  function showSaveStatus(msg, isError) {
    if (!saveStatus) return;
    saveStatus.textContent = msg;
    saveStatus.className = 'save-status ' + (isError ? 'status-error' : 'status-ok');
    setTimeout(function () {
      saveStatus.textContent = '';
      saveStatus.className = 'save-status';
    }, 3000);
  }

  fetch('/api/assertions')
    .then(function (r) { return r.json(); })
    .then(function (data) { renderThresholds(data); })
    .catch(function () { renderThresholds([]); });

  if (btnAdd) {
    btnAdd.addEventListener('click', function () {
      var emptyRow = tbody.querySelector('.threshold-empty-row');
      if (emptyRow) emptyRow.remove();
      tbody.appendChild(createThresholdRow({
        metric: 'p95_latency', operator: 'less_than', value: 500, enabled: true,
      }));
    });
  }

  if (btnSave) {
    btnSave.addEventListener('click', function () {
      btnSave.disabled = true;
      btnSave.textContent = 'Saving…';

      var payload = collectThresholds();

      fetch('/api/assertions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
        .then(function (r) { return r.json().then(function (b) { return { ok: r.ok, body: b }; }); })
        .then(function (r) {
          if (r.ok) {
            showSaveStatus('Changes saved successfully.', false);
          } else {
            showSaveStatus(r.body.error || 'Save failed.', true);
          }
        })
        .catch(function () { showSaveStatus('Network error. Please try again.', true); })
        .finally(function () {
          btnSave.disabled = false;
          btnSave.textContent = 'Save Changes';
        });
    });
  }
}());

(function () {
  'use strict';

  var detailOverlay = document.getElementById('modal-run-detail');
  var detailBody    = document.getElementById('detail-body');
  var btnClose      = document.getElementById('btn-detail-close');
  var headerBadge   = document.getElementById('detail-header-badge');
  var headerTitle   = document.getElementById('detail-modal-title');

  if (!detailOverlay) return;

  function openDetail() {
    detailOverlay.classList.add('open');
  }

  function closeDetail() {
    detailOverlay.classList.remove('open');
    if (detailBody)  detailBody.innerHTML = '<div class="detail-loading">Loading...</div>';
    if (headerBadge) headerBadge.innerHTML = '';
    if (headerTitle) headerTitle.textContent = 'Run Details';
  }

  if (btnClose) btnClose.addEventListener('click', closeDetail);

  detailOverlay.addEventListener('click', function (e) {
    if (e.target === detailOverlay) closeDetail();
  });

  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && detailOverlay.classList.contains('open')) closeDetail();
  });

  document.querySelectorAll('.history-row').forEach(function (row) {
    row.addEventListener('click', function () {
      var id = row.dataset.id;
      if (!id) return;
      openDetail();
      fetch('/api/history?id=' + encodeURIComponent(id))
        .then(function (res) { return res.json(); })
        .then(function (d)   { renderDetail(d); })
        .catch(function ()   {
          if (detailBody) detailBody.innerHTML = '<p class="detail-error">Failed to load run details.</p>';
        });
    });
  });

  function esc(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  function heroStat(value, label, mod) {
    return '<div class="detail-hero-stat' + (mod ? ' ' + mod : '') + '">' +
      '<span class="detail-hero-value">' + esc(String(value)) + '</span>' +
      '<span class="detail-hero-label">' + label + '</span>' +
      '</div>';
  }

  function metricCard(value, label, mod) {
    return '<div class="detail-metric-card' + (mod ? ' ' + mod : '') + '">' +
      '<span class="detail-metric-value">' + esc(String(value)) + '</span>' +
      '<span class="detail-metric-label">' + label + '</span>' +
      '</div>';
  }

  function latencyCard(pct, value) {
    return '<div class="detail-latency-card">' +
      '<span class="detail-latency-pct">' + pct + '</span>' +
      '<span class="detail-latency-val">' + esc(value) + '</span>' +
      '</div>';
  }

  function infoRow(label, value) {
    return '<div class="detail-info-row">' +
      '<span class="detail-info-label">' + label + '</span>' +
      '<span class="detail-info-value">' + value + '</span>' +
      '</div>';
  }

  function emptyPanel(msg) {
    return '<div class="detail-empty-panel">' + esc(msg) + '</div>';
  }

  function buildOverviewPanel(d) {
    var html = '<div class="detail-tab-panel" data-panel="overview">';
    html += '<div class="detail-info-list">';
    html += infoRow('Status', '<span class="status-badge status-' + esc(d.Status) + '">' + esc(d.Status) + '</span>');
    html += infoRow('Spec URL', '<span class="detail-url-full" title="' + esc(d.SpecURL) + '">' + esc(d.SpecURL) + '</span>');
    html += infoRow('Profile', esc(d.Profile));
    html += infoRow('Workers', esc(d.Workers));
    html += infoRow('Config Duration', esc(d.Duration));
    html += infoRow('Started', esc(d.StartedAt));
    html += infoRow('Ended', esc(d.EndedAt) || '—');
    html += infoRow('Elapsed', esc(d.Elapsed));
    html += '</div>';
    if (d.Error) {
      html += '<div class="detail-error-banner">' +
        '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>' +
        '<span>' + esc(d.Error) + '</span></div>';
    }
    html += '</div>';
    return html;
  }

  function buildMetricsPanel(d) {
    var html = '<div class="detail-tab-panel hidden" data-panel="metrics">';
    html += '<div class="detail-metric-cards">';
    html += metricCard(d.Requests  || '—', 'Total Requests', '');
    html += metricCard(d.Successes || '—', 'Successes',      'card-success');
    html += metricCard(d.Failures  || '—', 'Failures',       Number(d.Failures) > 0 ? 'card-danger' : '');
    html += metricCard(d.RPS       || '—', 'Avg RPS',        '');
    html += metricCard(d.ErrorRate || '—', 'Error Rate',     '');
    html += metricCard(d.DataBytes || '—', 'Data Received',  '');
    html += '</div>';
    if (d.P50 || d.P90 || d.P95 || d.P99) {
      html += '<div class="detail-latency-section">';
      html += '<p class="detail-sub-title">Latency Percentiles</p>';
      html += '<div class="detail-latency-cards">';
      html += latencyCard('P50', d.P50 || '—');
      html += latencyCard('P90', d.P90 || '—');
      html += latencyCard('P95', d.P95 || '—');
      html += latencyCard('P99', d.P99 || '—');
      html += '</div></div>';
    }
    html += '</div>';
    return html;
  }

  function buildStatusCodesPanel(d) {
    var html = '<div class="detail-tab-panel hidden" data-panel="status-codes">';
    if (d.StatusCodes && d.StatusCodes.length > 0) {
      var total = Number(d.Requests) || 1;
      html += '<table class="detail-table"><thead><tr><th>Status</th><th>Count</th><th>Share</th></tr></thead><tbody>';
      d.StatusCodes.forEach(function (sc) {
        var share = total > 0 ? ((sc.Count / total) * 100).toFixed(1) + '%' : '—';
        var mod   = sc.Code >= 500 ? 'code-5xx' : sc.Code >= 400 ? 'code-4xx' : 'code-2xx';
        html += '<tr><td><span class="http-code-badge ' + mod + '">HTTP ' + esc(sc.Code) + '</span></td>' +
                '<td>' + esc(sc.Count) + '</td><td>' + share + '</td></tr>';
      });
      html += '</tbody></table>';
    } else {
      html += emptyPanel('No status code data available for this run.');
    }
    html += '</div>';
    return html;
  }

  function buildErrorsPanel(d) {
    var html = '<div class="detail-tab-panel hidden" data-panel="errors">';
    if (d.Errors && d.Errors.length > 0) {
      html += '<table class="detail-table"><thead><tr><th>Count</th><th>Error Message</th></tr></thead><tbody>';
      d.Errors.forEach(function (e) {
        html += '<tr><td><span class="error-count-badge">' + esc(e.Count) + '</span></td>' +
                '<td class="detail-err-msg">' + esc(e.Message) + '</td></tr>';
      });
      html += '</tbody></table>';
    } else {
      html += '<div class="detail-no-errors">' +
        '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>' +
        '<span>No errors recorded</span></div>';
    }
    html += '</div>';
    return html;
  }

  function buildSlaPanel(d) {
    var html = '<div class="detail-tab-panel hidden" data-panel="sla">';
    if (d.AssertionResults && d.AssertionResults.length > 0) {
      var passed    = d.AssertionsPassed;
      var badgeMod  = passed ? 'all-passed' : 'has-failures';
      var badgeText = passed ? 'All thresholds passed' : 'Some thresholds failed';
      html += '<div class="sla-panel-header"><span class="assertion-summary-badge ' + badgeMod + '">' + badgeText + '</span></div>';
      html += '<table class="detail-table"><thead><tr><th>Metric</th><th>Expected</th><th>Actual</th><th>Result</th></tr></thead><tbody>';
      d.AssertionResults.forEach(function (ar) {
        var rowMod = ar.Passed ? 'sla-row-pass' : 'sla-row-fail';
        html += '<tr class="' + rowMod + '">' +
          '<td>' + esc(ar.Metric) + '</td>' +
          '<td>' + esc(ar.Expected) + '</td>' +
          '<td><strong>' + esc(ar.Actual) + '</strong></td>' +
          '<td><span class="sla-result-badge ' + (ar.Passed ? 'sla-pass' : 'sla-fail') + '">' + (ar.Passed ? 'PASS' : 'FAIL') + '</span></td>' +
          '</tr>';
      });
      html += '</tbody></table>';
    } else {
      html += emptyPanel('No SLA thresholds were configured for this run.');
    }
    html += '</div>';
    return html;
  }

  function renderDetail(d) {
    if (headerBadge) {
      headerBadge.innerHTML = '<span class="status-badge status-' + esc(d.Status) + '">' + esc(d.Status) + '</span>';
    }
    if (headerTitle) {
      headerTitle.textContent = d.SpecURL || 'Run Details';
      headerTitle.title = d.SpecURL || '';
    }

    var errorRateNum = parseFloat(d.ErrorRate) || 0;
    var heroHTML =
      heroStat(d.Requests  || '—', 'Requests',   '') +
      heroStat(d.RPS       || '—', 'Avg RPS',    '') +
      heroStat(d.ErrorRate || '—', 'Error Rate',  errorRateNum > 0 ? 'hero-danger' : 'hero-success') +
      heroStat(d.Elapsed   || '—', 'Elapsed',    '');

    var errorCount = d.Errors ? d.Errors.length : 0;
    var slaLabel   = d.AssertionResults && d.AssertionResults.length > 0
      ? 'SLA ' + (d.AssertionsPassed ? '✓' : '✗')
      : 'SLA';

    var tabs = [
      { id: 'overview',     label: 'Overview'     },
      { id: 'metrics',      label: 'Metrics'      },
      { id: 'status-codes', label: 'Status Codes' },
      { id: 'errors',       label: errorCount > 0 ? 'Errors (' + errorCount + ')' : 'Errors' },
      { id: 'sla',          label: slaLabel },
    ];

    var tabNavHTML = tabs.map(function (t, i) {
      return '<button class="detail-tab' + (i === 0 ? ' active' : '') + '" data-tab="' + t.id + '">' + t.label + '</button>';
    }).join('');

    var panelsHTML =
      buildOverviewPanel(d) +
      buildMetricsPanel(d) +
      buildStatusCodesPanel(d) +
      buildErrorsPanel(d) +
      buildSlaPanel(d);

    detailBody.innerHTML =
      '<div class="detail-hero">'    + heroHTML   + '</div>' +
      '<div class="detail-tab-nav">' + tabNavHTML + '</div>' +
      '<div class="detail-panels">'  + panelsHTML + '</div>';

    var allTabs   = detailBody.querySelectorAll('.detail-tab');
    var allPanels = detailBody.querySelectorAll('.detail-tab-panel');

    allTabs.forEach(function (tab) {
      tab.addEventListener('click', function () {
        allTabs.forEach(function (t)   { t.classList.remove('active'); });
        allPanels.forEach(function (p) { p.classList.add('hidden'); });
        tab.classList.add('active');
        var target = detailBody.querySelector('[data-panel="' + tab.dataset.tab + '"]');
        if (target) target.classList.remove('hidden');
      });
    });
  }
}());

(function () {
  'use strict';

  var btnChange  = document.getElementById('btn-change-password');
  var pwCurrent  = document.getElementById('pw-current');
  var pwNew      = document.getElementById('pw-new');
  var pwConfirm  = document.getElementById('pw-confirm');
  var pwAlert    = document.getElementById('pw-alert');

  if (!btnChange) return;

  function showAlert(msg, isError) {
    pwAlert.textContent = msg;
    pwAlert.className = 'pw-alert ' + (isError ? 'pw-alert-error' : 'pw-alert-success');
    pwAlert.style.display = 'block';
  }

  btnChange.addEventListener('click', function () {
    var current = pwCurrent.value.trim();
    var newPw   = pwNew.value.trim();
    var confirm = pwConfirm.value.trim();

    if (!current || !newPw || !confirm) {
      showAlert('All fields are required.', true);
      return;
    }

    if (newPw !== confirm) {
      showAlert('New password and confirm password do not match.', true);
      return;
    }

    if (newPw.length < 6) {
      showAlert('New password must be at least 6 characters.', true);
      return;
    }

    btnChange.disabled = true;
    btnChange.textContent = 'Saving…';

    fetch('/api/settings/password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ current: current, new: newPw, confirm: confirm }),
    })
      .then(function (r) { return r.json().then(function (b) { return { ok: r.ok, body: b }; }); })
      .then(function (r) {
        if (r.ok) {
          showAlert('Password changed successfully.', false);
          pwCurrent.value = '';
          pwNew.value = '';
          pwConfirm.value = '';
        } else {
          showAlert(r.body.error || 'Failed to change password.', true);
        }
      })
      .catch(function () { showAlert('Network error. Please try again.', true); })
      .finally(function () {
        btnChange.disabled = false;
        btnChange.textContent = 'Change Password';
      });
  });
}());
