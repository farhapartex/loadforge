(function () {
  'use strict';

  // ── Alert auto-dismiss ────────────────────────────────────────
  document.querySelectorAll('.alert').forEach(function (el) {
    setTimeout(function () {
      el.style.transition = 'opacity 0.4s ease';
      el.style.opacity = '0';
      setTimeout(function () { el.remove(); }, 400);
    }, 4000);
  });

  // ── Run Test modal ────────────────────────────────────────────
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

  // Source tab switching
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

  // ── Stop test ─────────────────────────────────────────────────
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

  // ── Live stats polling ────────────────────────────────────────
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

  // ── Activity log stream ───────────────────────────────────────
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

// ── History detail modal ──────────────────────────────────────
(function () {
  'use strict';

  var detailOverlay = document.getElementById('modal-run-detail');
  var detailBody    = document.getElementById('detail-body');
  var btnClose      = document.getElementById('btn-detail-close');

  if (!detailOverlay) return;

  function openDetail() {
    detailOverlay.classList.add('open');
  }

  function closeDetail() {
    detailOverlay.classList.remove('open');
    if (detailBody) detailBody.innerHTML = '<div class="detail-loading">Loading...</div>';
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
        .then(function (d) { renderDetail(d); })
        .catch(function () {
          detailBody.innerHTML = '<p class="detail-error">Failed to load run details.</p>';
        });
    });
  });

  function esc(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  function renderDetail(d) {
    var html = '';

    // Overview
    html += '<div class="detail-section">';
    html += '<h4 class="detail-section-title">Overview</h4>';
    html += '<div class="detail-grid">';
    html += detailCell('Status', '<span class="status-badge status-' + esc(d.Status) + '">' + esc(d.Status) + '</span>');
    html += detailCell('Spec URL', '<span class="detail-url" title="' + esc(d.SpecURL) + '">' + esc(d.SpecURL) + '</span>');
    html += detailCell('Profile', esc(d.Profile));
    html += detailCell('Workers', esc(d.Workers));
    html += detailCell('Config Duration', esc(d.Duration));
    html += detailCell('Started', esc(d.StartedAt));
    html += detailCell('Ended', esc(d.EndedAt) || '—');
    html += detailCell('Elapsed', esc(d.Elapsed));
    if (d.Error) html += detailCell('Error', '<span class="detail-err">' + esc(d.Error) + '</span>');
    html += '</div></div>';

    // Metrics
    html += '<div class="detail-section">';
    html += '<h4 class="detail-section-title">Metrics</h4>';
    html += '<div class="detail-grid">';
    html += detailCell('Total Requests', esc(d.Requests));
    html += detailCell('Successes', esc(d.Successes));
    html += detailCell('Failures', esc(d.Failures));
    html += detailCell('Error Rate', esc(d.ErrorRate));
    html += detailCell('Avg RPS', esc(d.RPS));
    html += detailCell('Data Received', esc(d.DataBytes));
    html += '</div></div>';

    // Latency
    if (d.P50 || d.P90 || d.P95 || d.P99) {
      html += '<div class="detail-section">';
      html += '<h4 class="detail-section-title">Latency Percentiles</h4>';
      html += '<div class="detail-grid">';
      html += detailCell('p50', esc(d.P50) || '—');
      html += detailCell('p90', esc(d.P90) || '—');
      html += detailCell('p95', esc(d.P95) || '—');
      html += detailCell('p99', esc(d.P99) || '—');
      html += '</div></div>';
    }

    // Status codes
    if (d.StatusCodes && d.StatusCodes.length > 0) {
      html += '<div class="detail-section">';
      html += '<h4 class="detail-section-title">Status Codes</h4>';
      html += '<table class="detail-table"><thead><tr><th>Code</th><th>Count</th></tr></thead><tbody>';
      d.StatusCodes.forEach(function (sc) {
        html += '<tr><td>HTTP ' + esc(sc.Code) + '</td><td>' + esc(sc.Count) + '</td></tr>';
      });
      html += '</tbody></table></div>';
    }

    // Errors
    if (d.Errors && d.Errors.length > 0) {
      html += '<div class="detail-section">';
      html += '<h4 class="detail-section-title">Errors</h4>';
      html += '<table class="detail-table"><thead><tr><th>Count</th><th>Message</th></tr></thead><tbody>';
      d.Errors.forEach(function (e) {
        html += '<tr><td>' + esc(e.Count) + '</td><td class="detail-err-msg">' + esc(e.Message) + '</td></tr>';
      });
      html += '</tbody></table></div>';
    }

    detailBody.innerHTML = html;
  }

  function detailCell(label, value) {
    return '<div class="detail-cell"><span class="detail-label">' + label + '</span><span class="detail-value">' + value + '</span></div>';
  }
}());
