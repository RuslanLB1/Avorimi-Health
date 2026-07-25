(function () {
  "use strict";

  // Ровно та же метка кнопок, что в support_flow.go mainMenuOptions —
  // используется только для синтетического приветствия при пустой истории.
  var mainMenuOptions = [
    "📝 Оставить отзыв",
    "💡 Предложение идеи",
    "📅 Как записаться на приём",
    "🔑 Помощь с регистрацией",
    "💳 Оплата и подписка",
    "🙋 Другой вопрос",
  ];

  var GUEST_ID_KEY = "avorimi_guest_id";
  var GUEST_NAME_KEY = "avorimi_guest_name";
  var GUEST_CONTACT_KEY = "avorimi_guest_contact";

  var scriptEl = document.currentScript;
  var isLoggedIn = scriptEl && scriptEl.getAttribute("data-logged-in") === "1";

  var panel = document.getElementById("supportPanel");
  var toggle = document.getElementById("supportToggle");
  var closeBtn = document.getElementById("supportClose");
  var newChatBtn = document.getElementById("supportNewChat");
  var body = document.getElementById("supportBody");
  var errorEl = document.getElementById("supportError");
  var input = document.getElementById("supportInput");
  var sendBtn = document.getElementById("supportSend");
  var guestFields = document.getElementById("supportGuestFields");
  var guestNameInput = document.getElementById("supportGuestName");
  var guestContactInput = document.getElementById("supportGuestContact");
  var replyPreview = document.getElementById("supportReplyPreview");
  var replyPreviewBody = document.getElementById("supportReplyPreviewBody");
  var replyCancelBtn = document.getElementById("supportReplyCancel");

  var open = false;
  var sending = false;
  var pollTimer = null;
  var guestId = localStorage.getItem(GUEST_ID_KEY) || "";
  var replyTo = null; // {id, body} — сообщение, на которое отвечаем

  function esc(s) {
    var d = document.createElement("div");
    d.textContent = s;
    return d.innerHTML;
  }

  function showError(msg) {
    if (!msg) {
      errorEl.hidden = true;
      errorEl.textContent = "";
      return;
    }
    errorEl.hidden = false;
    errorEl.textContent = msg;
  }

  function scrollToBottom() {
    body.scrollTop = body.scrollHeight;
  }

  function setReplyTo(id, text) {
    replyTo = { id: id, body: text };
    replyPreviewBody.textContent = text.length > 90 ? text.slice(0, 90) + "…" : text;
    replyPreview.hidden = false;
    input.focus();
  }

  function clearReplyTo() {
    replyTo = null;
    replyPreview.hidden = true;
  }

  function renderMessages(messages) {
    body.innerHTML = "";
    if (messages.length === 0) {
      var greet = document.createElement("div");
      greet.className = "support-hint";
      greet.textContent = isLoggedIn
        ? "Здравствуйте! Чем могу помочь?"
        : "Не получается зарегистрироваться или войти? Опишите проблему — оператор ответит прямо здесь.";
      body.appendChild(greet);

      if (isLoggedIn) {
        var opts = document.createElement("div");
        opts.className = "support-options";
        mainMenuOptions.forEach(function (opt) {
          var b = document.createElement("button");
          b.type = "button";
          b.textContent = opt;
          b.onclick = function () {
            send(opt);
          };
          opts.appendChild(b);
        });
        body.appendChild(opts);
      }
      return;
    }

    messages.forEach(function (m) {
      var row = document.createElement("div");
      row.className = "support-msg " + (m.role === "user" ? "user" : "assistant");

      if (m.replyToBody) {
        var quote = document.createElement("div");
        quote.className = "support-quote";
        quote.textContent = m.replyToBody;
        row.appendChild(quote);
      }

      var textEl = document.createElement("div");
      textEl.innerHTML = esc(m.body).replace(/\n/g, "<br>");
      row.appendChild(textEl);

      if (m.options && m.options.length) {
        var optsWrap = document.createElement("div");
        optsWrap.className = "support-options";
        m.options.forEach(function (opt) {
          var b = document.createElement("button");
          b.type = "button";
          b.textContent = opt;
          b.onclick = function () {
            send(opt);
          };
          optsWrap.appendChild(b);
        });
        row.appendChild(optsWrap);
      }

      var replyBtn = document.createElement("button");
      replyBtn.type = "button";
      replyBtn.className = "support-reply-btn";
      replyBtn.textContent = "↩ Ответить";
      replyBtn.onclick = function () {
        setReplyTo(m.id, m.body);
      };
      row.appendChild(replyBtn);

      body.appendChild(row);
    });
    scrollToBottom();
  }

  function loadLoggedIn() {
    fetch("/api/support/messages", { credentials: "same-origin" })
      .then(function (r) {
        return r.ok ? r.json() : [];
      })
      .then(renderMessages)
      .catch(function () {});
  }

  function loadGuest() {
    if (!guestId) {
      renderMessages([]);
      return;
    }
    fetch("/api/support/guest/messages?guestId=" + encodeURIComponent(guestId), { credentials: "same-origin" })
      .then(function (r) {
        return r.ok ? r.json() : [];
      })
      .then(renderMessages)
      .catch(function () {});
  }

  function load() {
    if (isLoggedIn) loadLoggedIn();
    else loadGuest();
  }

  function send(text) {
    text = (text || input.value).trim();
    if (!text || sending) return;
    showError("");
    var replyToId = replyTo ? replyTo.id : undefined;

    if (isLoggedIn) {
      sending = true;
      input.value = "";
      clearReplyTo();
      fetch("/api/support/messages", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ body: text, replyToId: replyToId }),
      })
        .then(function (r) {
          if (!r.ok) return r.json().then(function (e) { throw new Error(e.error || "Ошибка отправки"); });
          return r.json();
        })
        .then(function () {
          sending = false;
          load();
        })
        .catch(function (e) {
          sending = false;
          showError(e.message === "err.supportLimitReached" ? "Дневной лимит сообщений исчерпан — попробуйте завтра." : "Не удалось отправить сообщение");
        });
      return;
    }

    // Гость: имя/контакт нужны только на первое сообщение.
    var name = guestNameInput ? guestNameInput.value.trim() : "";
    var contact = guestContactInput ? guestContactInput.value.trim() : "";
    if (!guestId && (!name || !contact)) {
      showError("Укажите имя и телефон/email, чтобы оператор мог ответить");
      return;
    }
    sending = true;
    input.value = "";
    clearReplyTo();
    fetch("/api/support/guest", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ guestId: guestId || undefined, name: name || undefined, contact: contact || undefined, message: text, replyToId: replyToId }),
    })
      .then(function (r) {
        if (!r.ok) return r.json().then(function (e) { throw new Error(e.error || "Ошибка отправки"); });
        return r.json();
      })
      .then(function (res) {
        sending = false;
        if (res.guestId) {
          guestId = res.guestId;
          localStorage.setItem(GUEST_ID_KEY, guestId);
          localStorage.setItem(GUEST_NAME_KEY, name);
          localStorage.setItem(GUEST_CONTACT_KEY, contact);
          if (guestFields) guestFields.hidden = true;
        }
        load();
      })
      .catch(function (e) {
        sending = false;
        showError(e.message || "Не удалось отправить сообщение");
      });
  }

  function startNewChat() {
    showError("");
    clearReplyTo();
    if (isLoggedIn) {
      fetch("/api/support/reset", { method: "POST", credentials: "same-origin" })
        .then(function (r) {
          if (!r.ok) throw new Error();
          return r.json();
        })
        .then(load)
        .catch(function () {
          showError("Не удалось начать новый чат");
        });
      return;
    }
    localStorage.removeItem(GUEST_ID_KEY);
    localStorage.removeItem(GUEST_NAME_KEY);
    localStorage.removeItem(GUEST_CONTACT_KEY);
    guestId = "";
    if (guestFields) guestFields.hidden = false;
    if (guestNameInput) guestNameInput.value = "";
    if (guestContactInput) guestContactInput.value = "";
    renderMessages([]);
  }

  function startPolling() {
    stopPolling();
    pollTimer = setInterval(load, 5000);
  }
  function stopPolling() {
    if (pollTimer) clearInterval(pollTimer);
    pollTimer = null;
  }

  function openPanel() {
    open = true;
    panel.hidden = false;
    toggle.textContent = "✕";
    if (!isLoggedIn && guestId && guestFields) guestFields.hidden = true;
    load();
    startPolling();
    input.focus();
  }

  function closePanel() {
    open = false;
    panel.hidden = true;
    toggle.textContent = "💬";
    stopPolling();
  }

  toggle.addEventListener("click", function () {
    if (open) closePanel();
    else openPanel();
  });
  closeBtn.addEventListener("click", closePanel);
  newChatBtn.addEventListener("click", startNewChat);
  replyCancelBtn.addEventListener("click", clearReplyTo);
  sendBtn.addEventListener("click", function () {
    send();
  });
  input.addEventListener("keydown", function (e) {
    if (e.key === "Enter") send();
  });
})();
