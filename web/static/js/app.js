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
      '<span class="log-time">'  + entry.timestamp          + '</span>' +
      '<span class="log-level">' + '[' + entry.level + ']'  + '</span>' +
      '<span class="log-msg">'   + escapeHtml(entry.message) + '</span>';

    logOutput.appendChild(line);

    while (logOutput.children.length > MAX_ENTRIES) {
      logOutput.removeChild(logOutput.firstChild);
    }

    if (autoScroll) {
      logOutput.scrollTop = logOutput.scrollHeight;
    }
  }

  function systemEntry(msg) {
    appendEntry({
      timestamp: new Date().toTimeString().slice(0, 8),
      level:     'WARN',
      message:   msg,
    });
  }

  var source = new EventSource('/api/logs/stream');

  source.onmessage = function (e) {
    try {
      appendEntry(JSON.parse(e.data));
    } catch (_) {}
  };

  source.onerror = function () {
    systemEntry('Log stream disconnected. Reconnecting...');
  };
})();
