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
  var modalOverlay  = document.getElementById('modal-run-test');
  var btnRunTest    = document.getElementById('btn-run-test');
  var btnModalClose = document.getElementById('btn-modal-close');
  var btnModalCancel= document.getElementById('btn-modal-cancel');
  var btnModalSubmit= document.getElementById('btn-modal-submit');
  var inputApiUrl   = document.getElementById('input-api-doc-url');

  function openModal() {
    if (!modalOverlay) return;
    modalOverlay.classList.add('open');
    if (inputApiUrl) inputApiUrl.focus();
  }

  function closeModal() {
    if (!modalOverlay) return;
    modalOverlay.classList.remove('open');
    if (inputApiUrl) inputApiUrl.value = '';
    var jwtInput = document.getElementById('input-jwt-token');
    if (jwtInput) jwtInput.value = '';
  }

  if (btnRunTest)     btnRunTest.addEventListener('click', openModal);
  if (btnModalClose)  btnModalClose.addEventListener('click', closeModal);
  if (btnModalCancel) btnModalCancel.addEventListener('click', closeModal);

  if (modalOverlay) {
    modalOverlay.addEventListener('click', function (e) {
      if (e.target === modalOverlay) closeModal();
    });
  }

  if (btnModalSubmit) {
    btnModalSubmit.addEventListener('click', function () {
      if (!inputApiUrl || !inputApiUrl.value.trim()) {
        inputApiUrl.focus();
        inputApiUrl.reportValidity();
        return;
      }
      // TODO: wire to POST /api/run when backend is ready
    });
  }

  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && modalOverlay && modalOverlay.classList.contains('open')) {
      closeModal();
    }
  });

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
    try {
      appendEntry(JSON.parse(e.data));
    } catch (_) {}
  };

  source.onerror = function () {
    appendEntry({
      timestamp: new Date().toTimeString().slice(0, 8),
      level:     'WARN',
      message:   'Log stream disconnected. Reconnecting...',
    });
  };
})();
