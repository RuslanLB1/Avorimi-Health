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

  // Ровно та же строка, что closeChatTrigger в support_flow.go.
  var CLOSE_CHAT_TRIGGER = "🔚 Завершить чат";

  var panel = document.getElementById("supportPanel");
  var toggle = document.getElementById("supportToggle");
  var closeBtn = document.getElementById("supportClose");
  var newChatBtn = document.getElementById("supportNewChat");
  var endChatBtn = document.getElementById("supportEndChat");
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

  var attachBtn = document.getElementById("supportAttachBtn");
  var attachMenu = document.getElementById("supportAttachMenu");
  var pickGallery = document.getElementById("supportPickGallery");
  var pickCamera = document.getElementById("supportPickCamera");
  var pickVideo = document.getElementById("supportPickVideo");
  var pickFile = document.getElementById("supportPickFile");
  var fileGallery = document.getElementById("supportFileGallery");
  var fileCamera = document.getElementById("supportFileCamera");
  var fileVideo = document.getElementById("supportFileVideo");
  var fileAny = document.getElementById("supportFileAny");
  var attachPreview = document.getElementById("supportAttachPreview");
  var attachPreviewName = document.getElementById("supportAttachPreviewName");
  var attachCancelBtn = document.getElementById("supportAttachCancel");

  var voiceBtn = document.getElementById("supportVoiceBtn");
  var recordingBar = document.getElementById("supportRecording");
  var recordingTime = document.getElementById("supportRecordingTime");
  var recordingStopBtn = document.getElementById("supportRecordingStop");
  var recordingCancelBtn = document.getElementById("supportRecordingCancel");

  var open = false;
  var sending = false;
  var pollTimer = null;
  var guestId = localStorage.getItem(GUEST_ID_KEY) || "";
  var replyTo = null; // {id, body} — сообщение, на которое отвечаем
  var pendingAttachment = null; // {uploadId, kind, name}

  var mediaRecorder = null;
  var recordedChunks = [];
  var recordingTimer = null;
  var recordingSeconds = 0;

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

  function setPendingAttachment(att) {
    pendingAttachment = att;
    attachPreviewName.textContent = "📎 " + att.name;
    attachPreview.hidden = false;
  }

  function clearPendingAttachment() {
    pendingAttachment = null;
    attachPreview.hidden = true;
  }

  function uploadFile(file) {
    var form = new FormData();
    form.append("file", file);
    return fetch("/api/support/upload", { method: "POST", credentials: "same-origin", body: form }).then(function (r) {
      if (!r.ok) throw new Error("Не удалось загрузить файл");
      return r.json();
    });
  }

  function pickAndUpload(input) {
    var file = input.files && input.files[0];
    input.value = "";
    if (!file) return;
    showError("");
    uploadFile(file)
      .then(function (res) {
        setPendingAttachment({ uploadId: res.uploadId, kind: res.kind, name: res.name || file.name });
      })
      .catch(function () {
        showError("Не удалось загрузить файл");
      });
  }

  function renderAttachment(att, row) {
    var wrap = document.createElement("div");
    wrap.className = "support-attachment";

    if (att.kind === "image") {
      var imgLink = document.createElement("a");
      imgLink.href = att.url;
      imgLink.target = "_blank";
      imgLink.rel = "noopener";
      var img = document.createElement("img");
      img.src = att.url;
      img.alt = att.name || "фото";
      img.className = "support-attachment-img";
      imgLink.appendChild(img);
      wrap.appendChild(imgLink);
    } else if (att.kind === "video") {
      var video = document.createElement("video");
      video.src = att.url;
      video.controls = true;
      video.className = "support-attachment-video";
      wrap.appendChild(video);
    } else if (att.kind === "audio") {
      var audio = document.createElement("audio");
      audio.src = att.url;
      audio.controls = true;
      audio.className = "support-attachment-audio";
      wrap.appendChild(audio);
    } else {
      var fileLink = document.createElement("a");
      fileLink.href = att.url;
      fileLink.target = "_blank";
      fileLink.rel = "noopener";
      fileLink.className = "support-attachment-file";
      fileLink.textContent = "📎 " + (att.name || "файл");
      wrap.appendChild(fileLink);
    }

    var dl = document.createElement("a");
    dl.href = att.url;
    dl.download = att.name || "";
    dl.className = "support-attachment-download";
    dl.textContent = "⬇ Скачать";
    wrap.appendChild(dl);

    row.appendChild(wrap);
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

      if (m.body) {
        var textEl = document.createElement("div");
        textEl.innerHTML = esc(m.body).replace(/\n/g, "<br>");
        row.appendChild(textEl);
      }

      if (m.attachments && m.attachments.length) {
        m.attachments.forEach(function (att) {
          renderAttachment(att, row);
        });
      }

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
        setReplyTo(m.id, m.body || (m.attachments && m.attachments.length ? "[вложение]" : ""));
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
    if ((!text && !pendingAttachment) || sending) return;
    showError("");
    var replyToId = replyTo ? replyTo.id : undefined;
    var att = pendingAttachment;

    if (isLoggedIn) {
      sending = true;
      input.value = "";
      clearReplyTo();
      clearPendingAttachment();
      fetch("/api/support/messages", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          body: text,
          replyToId: replyToId,
          attachmentUploadId: att ? att.uploadId : undefined,
          attachmentKind: att ? att.kind : undefined,
          attachmentName: att ? att.name : undefined,
        }),
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
    clearPendingAttachment();
    fetch("/api/support/guest", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        guestId: guestId || undefined,
        name: name || undefined,
        contact: contact || undefined,
        message: text,
        replyToId: replyToId,
        attachmentUploadId: att ? att.uploadId : undefined,
        attachmentKind: att ? att.kind : undefined,
        attachmentName: att ? att.name : undefined,
      }),
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
    clearPendingAttachment();
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

  // --- Меню вложений (фото / камера / видео / файл) ---

  function toggleAttachMenu(show) {
    attachMenu.hidden = show === undefined ? !attachMenu.hidden : !show;
  }

  attachBtn.addEventListener("click", function (e) {
    e.stopPropagation();
    toggleAttachMenu();
  });
  document.addEventListener("click", function () {
    attachMenu.hidden = true;
  });
  attachMenu.addEventListener("click", function (e) {
    e.stopPropagation();
  });

  pickGallery.addEventListener("click", function () {
    toggleAttachMenu(false);
    fileGallery.click();
  });
  pickCamera.addEventListener("click", function () {
    toggleAttachMenu(false);
    fileCamera.click();
  });
  pickVideo.addEventListener("click", function () {
    toggleAttachMenu(false);
    fileVideo.click();
  });
  pickFile.addEventListener("click", function () {
    toggleAttachMenu(false);
    fileAny.click();
  });

  fileGallery.addEventListener("change", function () { pickAndUpload(fileGallery); });
  fileCamera.addEventListener("change", function () { pickAndUpload(fileCamera); });
  fileVideo.addEventListener("change", function () { pickAndUpload(fileVideo); });
  fileAny.addEventListener("change", function () { pickAndUpload(fileAny); });

  attachCancelBtn.addEventListener("click", clearPendingAttachment);

  // --- Запись голосового сообщения ---

  function formatTime(sec) {
    var m = Math.floor(sec / 60);
    var s = sec % 60;
    return m + ":" + (s < 10 ? "0" : "") + s;
  }

  function startRecording() {
    if (!navigator.mediaDevices || !window.MediaRecorder) {
      showError("Запись голоса не поддерживается этим браузером");
      return;
    }
    showError("");
    navigator.mediaDevices
      .getUserMedia({ audio: true })
      .then(function (stream) {
        recordedChunks = [];
        mediaRecorder = new MediaRecorder(stream);
        mediaRecorder.ondataavailable = function (e) {
          if (e.data && e.data.size > 0) recordedChunks.push(e.data);
        };
        mediaRecorder.onstop = function () {
          stream.getTracks().forEach(function (t) { t.stop(); });
          var wasCancelled = mediaRecorder._cancelled;
          mediaRecorder = null;
          if (wasCancelled) return;
          var blob = new Blob(recordedChunks, { type: "audio/webm" });
          if (blob.size === 0) return;
          var file = new File([blob], "voice-message.webm", { type: "audio/webm" });
          uploadFile(file)
            .then(function (res) {
              setPendingAttachment({ uploadId: res.uploadId, kind: "audio", name: "Голосовое сообщение" });
            })
            .catch(function () {
              showError("Не удалось загрузить голосовое сообщение");
            });
        };
        mediaRecorder.start();
        recordingSeconds = 0;
        recordingTime.textContent = formatTime(0);
        recordingBar.hidden = false;
        recordingTimer = setInterval(function () {
          recordingSeconds++;
          recordingTime.textContent = formatTime(recordingSeconds);
        }, 1000);
      })
      .catch(function () {
        showError("Не удалось получить доступ к микрофону");
      });
  }

  function stopRecording(cancelled) {
    if (recordingTimer) {
      clearInterval(recordingTimer);
      recordingTimer = null;
    }
    recordingBar.hidden = true;
    if (mediaRecorder && mediaRecorder.state !== "inactive") {
      mediaRecorder._cancelled = !!cancelled;
      mediaRecorder.stop();
    }
  }

  voiceBtn.addEventListener("click", startRecording);
  recordingStopBtn.addEventListener("click", function () { stopRecording(false); });
  recordingCancelBtn.addEventListener("click", function () { stopRecording(true); });

  toggle.addEventListener("click", function () {
    if (open) closePanel();
    else openPanel();
  });
  closeBtn.addEventListener("click", closePanel);
  newChatBtn.addEventListener("click", startNewChat);
  if (endChatBtn) {
    endChatBtn.addEventListener("click", function () {
      send(CLOSE_CHAT_TRIGGER);
    });
  }
  replyCancelBtn.addEventListener("click", clearReplyTo);
  sendBtn.addEventListener("click", function () {
    send();
  });
  input.addEventListener("keydown", function (e) {
    if (e.key === "Enter") send();
  });
})();
