/* try GUI — same-origin /api/* client */
(() => {
  "use strict";

  const SOURCES = ["all", "tries", "ship", "bug"];
  // key 与后端 bootstrapMessages 保持一致（camelCase），bootstrap 下发时覆盖此表
  const FB = {
    title: "Try Directory Selection", searchPrefix: "Search: ",
    hintBar: "Ctrl-T: New  Ctrl-D: Delete  Ctrl-R: Rename  Ctrl-G: Ship  Tab: Filter  Esc: Quit",
    emptyStateHint: "No directories yet", noMatchesHint: 'No matches for "%s"',
    createNew: "Create new: ", renamePrompt: "New name: ",
    deleteTitle: "Delete %d directories?", deleteOptionNo: "NO", deleteOptionYes: "YES",
    deleteCancelled: "Delete cancelled", shipTitle: "Ship try to project", shipMoveLabel: "Move to: ",
    deleteModeLabel: " DELETE MODE ", markedCount: "%d marked",
    timeJustNow: "just now", timeMinAgo: "%dm ago", timeHourAgo: "%dh ago", timeDayAgo: "%dd ago",
    filterAll: "all",
  };

  const S = {
    view: "selector", theme: "dark", messages: { ...FB },
    paths: { tries: [], ships: [] }, source: "all", query: "",
    entries: [], counts: {}, selected: 0, marked: new Set(),
    filesPath: "", filesRoot: "", files: [], fileSelected: 0, fileMarked: new Set(),
    inline: null, modal: null, toastTimer: 0,
  };

  const $ = (id) => document.getElementById(id);
  const el = {
    app: $("app"), title: $("title"), themeBtn: $("theme-btn"),
    selector: $("view-selector"), files: $("view-files"),
    search: $("search"), searchPrefix: $("search-prefix"), tabs: $("source-tabs"),
    entryList: $("entry-list"), hintBar: $("hint-bar"),
    inlineBox: $("inline-input"), inlineLabel: $("inline-label"), inlineField: $("inline-field"),
    crumb: $("breadcrumb"), fileList: $("file-list"),
    filesBack: $("files-back"), filesDelete: $("files-delete"),
    modal: $("modal"), modalTitle: $("modal-title"), modalBody: $("modal-body"),
    modalExtra: $("modal-extra"), modalCancel: $("modal-cancel"), modalOk: $("modal-ok"),
    toast: $("toast"),
  };

  const msg = (k, ...a) => {
    let s = S.messages[k] ?? FB[k] ?? k;
    a.forEach((v) => { s = s.replace(/%[ds]/, v); });
    return s;
  };
  const pick = (o, ...ks) => { for (const k of ks) if (o && o[k] != null) return o[k]; };
  const esc = (s) => String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");

  async function api(method, path, body) {
    const opts = { method, headers: {} };
    if (body !== undefined) {
      opts.headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(body);
    }
    const res = await fetch(path, opts);
    let data = null;
    try { data = JSON.parse(await res.text() || "null"); } catch { /* ignore */ }
    if (!res.ok) throw new Error(pick(data, "error", "message") || res.statusText || "failed");
    return data;
  }

  function toast(text) {
    el.toast.textContent = text;
    el.toast.classList.remove("hidden");
    clearTimeout(S.toastTimer);
    S.toastTimer = setTimeout(() => el.toast.classList.add("hidden"), 2000);
  }

  function setTheme(t) {
    S.theme = t === "light" ? "light" : "dark";
    document.documentElement.classList.toggle("dark", S.theme === "dark");
    document.documentElement.classList.toggle("light", S.theme === "light");
  }

  function highlight(text, positions, query) {
    const set = new Set(Array.isArray(positions) ? positions : []);
    if (!set.size && query) {
      const L = text.toLowerCase(), q = query.toLowerCase();
      let qi = 0;
      for (let i = 0; i < L.length && qi < q.length; i++) if (L[i] === q[qi]) { set.add(i); qi++; }
      if (qi < q.length) set.clear();
    }
    if (!set.size) return esc(text);
    let out = "";
    for (let i = 0; i < text.length; i++) {
      const c = esc(text[i]);
      out += set.has(i) ? `<span class="match">${c}</span>` : c;
    }
    return out;
  }

  function relTime(raw) {
    const t = new Date(raw);
    if (Number.isNaN(+t)) return "";
    const sec = Math.floor((Date.now() - t) / 1000);
    if (sec < 60) return msg("timeJustNow");
    const m = Math.floor(sec / 60);
    if (m < 60) return msg("timeMinAgo", m);
    const h = Math.floor(m / 60);
    return h < 24 ? msg("timeHourAgo", h) : msg("timeDayAgo", Math.floor(h / 24));
  }

  function fmtSize(kb) {
    const n = +kb || 0;
    if (n < 1) return `${Math.round(n * 1024)}B`;
    return n < 1024 ? `${n.toFixed(1)}KB` : `${(n / 1024).toFixed(1)}MB`;
  }

  function normEntry(e) {
    const id = pick(e, "id", "path", "Path", "name", "Name");
    return {
      id: String(id), path: pick(e, "path", "Path", "id") || id,
      name: pick(e, "name", "Name") || "",
      baseName: pick(e, "baseName", "BaseName", "name") || "",
      date: pick(e, "date", "Date") || "",
      source: pick(e, "source", "Source") || "tries",
      score: +pick(e, "score", "Score") || 0,
      lastModified: pick(e, "lastModified", "LastModified", "modified") || "",
      highlightPositions: pick(e, "highlightPositions", "HighlightPositions", "highlights", "Highlights", "matches") || [],
    };
  }

  function normFile(e) {
    const name = pick(e, "name", "Name") || "";
    const isDir = !!(pick(e, "isDir", "IsDir") || pick(e, "type", "Type") === "dir");
    return {
      id: String(pick(e, "id", "path", "Path") || name), name, isDir,
      sizeKB: +pick(e, "sizeKB", "SizeKB", "size") || 0,
      modified: pick(e, "modified", "Modified", "lastModified") || "",
      path: pick(e, "path", "Path") || "",
    };
  }

  async function loadBootstrap() {
    try {
      const d = await api("GET", "/api/bootstrap");
      const m = pick(d, "messages", "Messages");
      if (m) S.messages = { ...FB, ...m };
      setTheme(pick(d, "theme", "Theme") || S.theme);
      S.paths = pick(d, "paths", "Paths") || S.paths;
      el.title.textContent = msg("title");
      el.searchPrefix.textContent = msg("searchPrefix");
    } catch (e) { toast(e.message); }
  }

  async function loadEntries() {
    try {
      const d = await api("GET", `/api/entries?q=${encodeURIComponent(S.query)}&source=${encodeURIComponent(S.source)}`);
      S.entries = (pick(d, "entries", "Entries") || []).map(normEntry);
      S.counts = pick(d, "counts", "Counts") || S.counts;
      if (S.selected >= S.entries.length) S.selected = Math.max(0, S.entries.length - 1);
    } catch (e) { toast(e.message); S.entries = []; }
    renderSelector();
  }

  async function loadFiles(path) {
    try {
      const d = await api("GET", `/api/files?path=${encodeURIComponent(path)}`);
      S.files = (pick(d, "files", "Files", "entries", "Entries") || []).map(normFile);
      S.filesPath = pick(d, "path", "Path") || path;
      S.fileSelected = 0;
      S.fileMarked.clear();
    } catch (e) { toast(e.message); }
    renderFiles();
  }

  function cur() { return S.entries[S.selected] || null; }
  function move(key, len, d) { if (len) S[key] = (S[key] + d + len) % len; }

  function renderSelector() {
    el.selector.classList.toggle("hidden", S.view !== "selector");
    el.files.classList.toggle("hidden", S.view !== "files");
    el.tabs.querySelectorAll(".tab").forEach((b) => {
      b.classList.toggle("active", b.dataset.source === S.source);
      const lab = b.dataset.source === "all" ? msg("filterAll") : b.dataset.source;
      // 后端来源计数以 "" 表示全部
      const c = S.counts[b.dataset.source === "all" ? "" : b.dataset.source];
      b.textContent = c != null ? `${lab} ${c}` : lab;
    });
    if (!S.entries.length) {
      el.entryList.innerHTML = `<div class="empty">${esc(S.query ? msg("noMatchesHint", S.query) : msg("emptyStateHint"))}</div>`;
    } else {
      el.entryList.innerHTML = S.entries.map((e, i) => {
        const mk = S.marked.has(e.id), sel = i === S.selected;
        const arrow = mk ? `<span class="mark-icon">✕</span>` : `<span class="arrow">${sel ? "›" : ""}</span>`;
        const nm = highlight(e.baseName || e.name, e.highlightPositions, S.query);
        const dt = e.date ? ` <span class="muted">${esc(e.date)}</span>` : "";
        return `<button type="button" class="row${sel ? " selected" : ""}${mk ? " marked" : ""}" data-i="${i}" role="option">${arrow}<span class="name">${nm}${dt}</span><span class="meta">${esc(relTime(e.lastModified))}</span><span class="source-badge">${esc(e.source)}</span></button>`;
      }).join("");
      el.entryList.querySelector(".row.selected")?.scrollIntoView({ block: "nearest" });
    }
    if (S.inline) {
      el.inlineBox.classList.remove("hidden");
      el.inlineLabel.textContent = S.inline.mode === "rename" ? msg("renamePrompt") : msg("createNew");
      if (document.activeElement !== el.inlineField) el.inlineField.value = S.inline.value || "";
    } else el.inlineBox.classList.add("hidden");
    if (S.marked.size) {
      el.hintBar.className = "hint danger-mode";
      el.hintBar.textContent = `${msg("deleteModeLabel").trim()}  ${msg("markedCount", S.marked.size)}  |  Ctrl-D · Enter · Esc`;
    } else {
      el.hintBar.className = "hint muted";
      el.hintBar.textContent = msg("hintBar");
    }
  }

  function rootName(p) {
    const parts = String(p || "").replace(/\\/g, "/").split("/").filter(Boolean);
    return parts.at(-1) || "root";
  }

  function pathParts(full, root) {
    const n = (p) => String(p || "").replace(/\\/g, "/").replace(/\/+$/, "");
    const f = n(full), r = n(root);
    let rel = (r && (f === r || f.startsWith(r + "/"))) ? f.slice(r.length).replace(/^\//, "") : f;
    const segs = rel ? rel.split("/").filter(Boolean) : [];
    const parts = [{ label: rootName(r || f), path: r || f }];
    let acc = r || "";
    segs.forEach((s) => { acc = acc ? `${acc}/${s}` : s; parts.push({ label: s, path: acc }); });
    return parts;
  }

  function renderFiles() {
    el.selector.classList.toggle("hidden", S.view !== "selector");
    el.files.classList.toggle("hidden", S.view !== "files");
    const parts = pathParts(S.filesPath, S.filesRoot);
    el.crumb.innerHTML = parts.map((p, i) => {
      const sep = i ? `<span class="sep">/</span>` : "";
      if (i === parts.length - 1) return `${sep}<span class="cur">${esc(p.label)}</span>`;
      return `${sep}<button type="button" data-path="${esc(p.path)}">${esc(p.label)}</button>`;
    }).join("");
    el.filesDelete.classList.toggle("hidden", !S.fileMarked.size);
    if (!S.files.length) {
      el.fileList.innerHTML = `<div class="empty">This directory is empty</div>`;
    } else {
      el.fileList.innerHTML = S.files.map((f, i) => {
        const sel = i === S.fileSelected, mk = S.fileMarked.has(f.id);
        return `<button type="button" class="row${f.isDir ? " isdir" : ""}${sel ? " selected" : ""}${mk ? " marked" : ""}" data-i="${i}"><span class="arrow">${sel ? "›" : ""}</span><span class="name"><span class="file-icon">${f.isDir ? "📁" : "📄"}</span> ${esc(f.name)}</span><span class="meta">${f.isDir ? "" : esc(fmtSize(f.sizeKB))}</span><span class="meta">${esc(relTime(f.modified))}</span></button>`;
      }).join("");
      el.fileList.querySelector(".row.selected")?.scrollIntoView({ block: "nearest" });
    }
  }

  function cycleSource(dir) {
    S.source = SOURCES[(SOURCES.indexOf(S.source) + dir + 4) % 4];
    S.selected = 0;
    loadEntries();
  }

  function toggleMark() {
    const e = cur();
    if (!e) return;
    S.marked.has(e.id) ? S.marked.delete(e.id) : S.marked.add(e.id);
    renderSelector();
  }

  function clearEsc() {
    if (S.inline) { S.inline = null; renderSelector(); return true; }
    if (S.modal) { closeModal(false); return true; }
    if (S.marked.size) { S.marked.clear(); renderSelector(); return true; }
    if (S.query) { S.query = ""; el.search.value = ""; loadEntries(); return true; }
    return false;
  }

  function openInline(mode, value) {
    S.inline = { mode, value: value || "" };
    renderSelector();
    requestAnimationFrame(() => { el.inlineField.focus(); el.inlineField.select(); });
  }

  async function submitInline() {
    if (!S.inline) return;
    const name = el.inlineField.value.trim(), mode = S.inline.mode;
    S.inline = null;
    if (!name) { renderSelector(); return; }
    try {
      if (mode === "create") {
        await api("POST", "/api/entries/create", { name });
        toast(`Created: ${name}`);
      } else {
        const e = cur();
        if (!e) return;
        await api("POST", "/api/entries/rename", { path: e.path || e.id, newName: name });
        toast(`Renamed → ${name}`);
      }
      await loadEntries();
    } catch (e) { toast(e.message); renderSelector(); }
  }

  function showModal(opts) {
    S.modal = opts;
    el.modalTitle.textContent = opts.title || "";
    el.modalBody.textContent = opts.body || "";
    el.modalExtra.classList.toggle("hidden", !opts.extraHtml);
    el.modalExtra.innerHTML = opts.extraHtml || "";
    el.modalCancel.textContent = opts.cancelLabel || "Cancel";
    el.modalOk.textContent = opts.okLabel || "OK";
    el.modalOk.className = opts.danger ? "btn danger" : "btn";
    el.modal.classList.remove("hidden");
    requestAnimationFrame(() => el.modalCancel.focus());
  }

  function closeModal(ok) {
    const m = S.modal;
    S.modal = null;
    el.modal.classList.add("hidden");
    if (ok && m?.onOk) m.onOk();
    else if (!ok && m?.onCancel) m.onCancel();
  }

  function confirmDeleteEntries() {
    const paths = [...S.marked];
    if (!paths.length) return;
    const names = S.entries.filter((e) => S.marked.has(e.id)).map((e) => e.name || e.baseName);
    showModal({
      title: msg("deleteTitle", paths.length),
      body: names.slice(0, 5).join("\n") + (names.length > 5 ? `\n+${names.length - 5} more` : ""),
      danger: true, cancelLabel: msg("deleteOptionNo"), okLabel: msg("deleteOptionYes"),
      onOk: async () => {
        try {
          await api("POST", "/api/entries/delete", { paths });
          S.marked.clear();
          toast(`Deleted ${paths.length}`);
          await loadEntries();
        } catch (e) { toast(e.message); }
      },
      onCancel: () => toast(msg("deleteCancelled")),
    });
  }

  function openShip() {
    const e = cur();
    if (!e) return;
    const ships = [].concat(pick(S.paths, "ships", "Ships") || []);
    const extra = ships.length
      ? `<select id="ship-dest">${ships.map((p, i) => `<option value="${i}">${esc(p)}</option>`).join("")}</select>`
      : `<input id="ship-dest" type="number" min="0" value="0" />`;
    showModal({
      title: msg("shipTitle"), body: `${msg("shipMoveLabel")}${e.name || e.baseName}`,
      extraHtml: extra, cancelLabel: "Cancel", okLabel: "Ship",
      onOk: async () => {
        try {
          await api("POST", "/api/entries/ship", {
            path: e.path || e.id,
            destIndex: +($("ship-dest")?.value) || 0,
          });
          toast("Shipped");
          await loadEntries();
        } catch (err) { toast(err.message); }
      },
    });
  }

  async function enterFiles() {
    const e = cur();
    if (!e) { if (S.query) openInline("create", S.query); return; }
    S.view = "files";
    S.filesRoot = e.path || e.id;
    await loadFiles(S.filesRoot);
  }

  function backToSelector() {
    S.view = "selector";
    S.fileMarked.clear();
    renderSelector();
  }

  function filePath(f) {
    if (f.path) return f.path;
    return `${S.filesPath.replace(/\/+$/, "")}/${f.name}`;
  }

  function parentPath() {
    const p = S.filesPath.replace(/\\/g, "/").replace(/\/+$/, "").replace(/\/[^/]+$/, "") || S.filesRoot;
    return p.length < S.filesRoot.length ? S.filesRoot : p;
  }

  async function activateFile() {
    const f = S.files[S.fileSelected];
    if (!f) return;
    if (f.isDir) { await loadFiles(filePath(f)); return; }
    try {
      await api("POST", "/api/files/open", { path: filePath(f) });
      toast("Opened");
    } catch (e) { toast(e.message); }
  }

  function confirmDeleteFiles() {
    const paths = S.files.filter((f) => S.fileMarked.has(f.id)).map(filePath);
    if (!paths.length) return;
    showModal({
      title: msg("deleteTitle", paths.length),
      body: paths.slice(0, 5).join("\n") + (paths.length > 5 ? `\n+${paths.length - 5} more` : ""),
      danger: true, cancelLabel: msg("deleteOptionNo"), okLabel: msg("deleteOptionYes"),
      onOk: async () => {
        try {
          await api("POST", "/api/files/delete", { paths });
          S.fileMarked.clear();
          toast(`Deleted ${paths.length}`);
          await loadFiles(S.filesPath);
        } catch (e) { toast(e.message); }
      },
      onCancel: () => toast(msg("deleteCancelled")),
    });
  }

  function onKey(ev) {
    const tag = ev.target.tagName;
    const inInput = tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";
    const mod = ev.ctrlKey || ev.metaKey;
    const k = ev.key;
    const kl = k.toLowerCase();

    if (S.modal) {
      if (k === "Escape") { ev.preventDefault(); closeModal(false); }
      else if (k === "Enter" && !inInput) { ev.preventDefault(); closeModal(true); }
      else if (k === "ArrowLeft" || k === "ArrowRight") {
        ev.preventDefault();
        (document.activeElement === el.modalOk ? el.modalCancel : el.modalOk).focus();
      }
      return;
    }

    if (S.inline && ev.target === el.inlineField) {
      if (k === "Enter") { ev.preventDefault(); submitInline(); }
      if (k === "Escape") { ev.preventDefault(); clearEsc(); }
      return;
    }

    if (mod && kl === "f") { ev.preventDefault(); el.search.focus(); return; }
    if (!inInput && k === "/") { ev.preventDefault(); el.search.focus(); return; }

    const navUp = k === "ArrowUp" || (mod && kl === "p");
    const navDn = k === "ArrowDown" || (mod && kl === "n");

    if (S.view === "selector") {
      if (navUp) { ev.preventDefault(); move("selected", S.entries.length, -1); renderSelector(); return; }
      if (navDn) { ev.preventDefault(); move("selected", S.entries.length, 1); renderSelector(); return; }
      if (k === "Tab") { ev.preventDefault(); cycleSource(ev.shiftKey ? -1 : 1); return; }
      if (k === "Enter") {
        ev.preventDefault();
        S.marked.size ? confirmDeleteEntries() : enterFiles();
        return;
      }
      if (k === " " && !inInput) { ev.preventDefault(); toggleMark(); return; }
      if (mod && kl === "d") { ev.preventDefault(); toggleMark(); return; }
      if (k === "Escape") { ev.preventDefault(); clearEsc(); if (inInput) el.app.focus(); return; }
      if (mod && kl === "t") { ev.preventDefault(); openInline("create", S.query); return; }
      if (mod && kl === "r") {
        ev.preventDefault();
        const e = cur();
        if (e) openInline("rename", e.baseName || e.name);
        return;
      }
      if (mod && kl === "g") { ev.preventDefault(); openShip(); return; }
      return;
    }

    // files
    if (k === "Escape") {
      ev.preventDefault();
      if (S.fileMarked.size) { S.fileMarked.clear(); renderFiles(); }
      else if (S.filesPath !== S.filesRoot) loadFiles(parentPath());
      else backToSelector();
      return;
    }
    if (navUp) { ev.preventDefault(); move("fileSelected", S.files.length, -1); renderFiles(); return; }
    if (navDn) { ev.preventDefault(); move("fileSelected", S.files.length, 1); renderFiles(); return; }
    if (k === "Enter") { ev.preventDefault(); activateFile(); return; }
    if (k === " " || (mod && kl === "d")) {
      ev.preventDefault();
      const f = S.files[S.fileSelected];
      if (!f) return;
      S.fileMarked.has(f.id) ? S.fileMarked.delete(f.id) : S.fileMarked.add(f.id);
      renderFiles();
      return;
    }
    if (k === "Delete") { ev.preventDefault(); confirmDeleteFiles(); }
  }

  el.search.addEventListener("input", () => { S.query = el.search.value; S.selected = 0; loadEntries(); });
  el.tabs.addEventListener("click", (ev) => {
    const b = ev.target.closest(".tab");
    if (!b) return;
    S.source = b.dataset.source; S.selected = 0; loadEntries();
  });
  el.entryList.addEventListener("click", (ev) => {
    const r = ev.target.closest(".row");
    if (!r) return;
    S.selected = +r.dataset.i; renderSelector();
  });
  el.entryList.addEventListener("dblclick", (ev) => {
    const r = ev.target.closest(".row");
    if (!r) return;
    S.selected = +r.dataset.i; enterFiles();
  });
  el.fileList.addEventListener("click", (ev) => {
    const r = ev.target.closest(".row");
    if (!r) return;
    S.fileSelected = +r.dataset.i; renderFiles();
  });
  el.fileList.addEventListener("dblclick", (ev) => {
    const r = ev.target.closest(".row");
    if (!r) return;
    S.fileSelected = +r.dataset.i; activateFile();
  });
  el.crumb.addEventListener("click", (ev) => {
    const b = ev.target.closest("button[data-path]");
    if (b) loadFiles(b.dataset.path);
  });
  el.filesBack.addEventListener("click", () => {
    S.filesPath !== S.filesRoot ? loadFiles(parentPath()) : backToSelector();
  });
  el.filesDelete.addEventListener("click", confirmDeleteFiles);
  el.themeBtn.addEventListener("click", () => setTheme(S.theme === "dark" ? "light" : "dark"));
  el.modalCancel.addEventListener("click", () => closeModal(false));
  el.modalOk.addEventListener("click", () => closeModal(true));
  el.modal.addEventListener("click", (ev) => { if (ev.target === el.modal) closeModal(false); });
  document.addEventListener("keydown", onKey);

  (async () => { await loadBootstrap(); await loadEntries(); el.app.focus(); })();
})();
