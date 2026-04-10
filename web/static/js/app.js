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
      logOutput.innerHTML = '';
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
