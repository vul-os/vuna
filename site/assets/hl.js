// ============================================================================
// hl.js — the site's own syntax highlighter. ~4 KB, no dependencies, no fetch.
// ----------------------------------------------------------------------------
// The footer of both pages claims that nothing here is fetched from a third
// party, so a CDN highlighter is not an option — it would make the page a liar.
// This covers only the languages the site actually contains: shell, TOML, Rust
// and JSON, plus a deliberate plain-text mode.
//
// COLOUR. brand/tokens.css is emphatic that colour is load-bearing on this site
// (gold = your own shard, slate = peer, green = running, ember = stub) and that
// "if two things on a screen are gold, at least one of them is wrong". So this
// is NOT a rainbow theme. It draws on four already-measured text tokens only —
// --ink, --ink-2, --ink-3 and --grain-text, each >= 4.5:1 on --sunk in both
// themes — and it spends the single gold accent sparsely, on the imperative
// word alone: the shell command name, the Rust keyword, the TOML table header.
// Never on strings, flags, keys or punctuation. Light and dark are one drawing
// because every colour is a custom property that the theme already flips.
//
// ASCII DIAGRAMS. docs/architecture.md fences two box-drawing diagrams with no
// language tag. Tokenising those would be vandalism, so a block containing box
// characters is forced to plain text no matter what it is labelled.
// ============================================================================
(function () {
  "use strict";

  function esc(s) {
    return s.replace(/[&<>]/g, function (c) {
      return c === "&" ? "&amp;" : c === "<" ? "&lt;" : "&gt;";
    });
  }

  /** `cls` of "" means "emit as plain text". */
  function tok(cls, text) {
    return cls ? '<span class="hl-' + cls + '">' + esc(text) + "</span>" : esc(text);
  }

  // ── Shell ────────────────────────────────────────────────────────────────
  // Hand-rolled rather than regex-listed because two things need position:
  // a word is a command only in command position, and a leading dash is a flag
  // only at the start of a word (otherwise `vul-os` in a URL reads as a flag).
  var SH_OP = /^(?:\|\||&&|[|;&])/;
  var SH_REDIR = /^(?:>>|<<|[<>])/;
  var SH_STR = /^(?:"(?:[^"\\]|\\[\s\S])*"|'[^']*')/;
  var SH_WORD = /^[^\s|;&<>()'"]+/;

  function shell(src) {
    var out = "";
    var i = 0;
    var atCmd = true; // start of a line, or just after | && ; ( — a command may begin
    var atWordStart = true; // preceded by whitespace or nothing: where # and - are special

    while (i < src.length) {
      var rest = src.slice(i);
      var ch = src[i];
      var m;

      if (ch === "\n") {
        out += "\n";
        i++;
        atCmd = true;
        atWordStart = true;
        continue;
      }
      if (ch === " " || ch === "\t") {
        m = /^[ \t]+/.exec(rest)[0];
        out += m;
        i += m.length;
        atWordStart = true;
        continue;
      }
      // A `#` only opens a comment at the start of a word.
      if (ch === "#" && atWordStart) {
        m = /^#[^\n]*/.exec(rest)[0];
        out += tok("c", m);
        i += m.length;
        atWordStart = false;
        continue;
      }
      if ((m = SH_STR.exec(rest))) {
        out += tok("s", m[0]);
        i += m[0].length;
        atCmd = false;
        atWordStart = false;
        continue;
      }
      if ((m = SH_OP.exec(rest))) {
        out += tok("p", m[0]);
        i += m[0].length;
        atCmd = true; // a new command may follow a pipe or &&
        atWordStart = true;
        continue;
      }
      if ((m = SH_REDIR.exec(rest))) {
        out += tok("p", m[0]);
        i += m[0].length;
        atCmd = false;
        atWordStart = true;
        continue;
      }
      if ((m = SH_WORD.exec(rest))) {
        var w = m[0];
        var cls = "";
        if (atWordStart && /^--?[A-Za-z]/.test(w)) cls = "f"; // a flag, not a command
        else if (atCmd && !/^[A-Z_]+=/.test(w)) cls = "k"; // the imperative word
        out += tok(cls, w);
        i += w.length;
        if (cls !== "f") atCmd = false;
        atWordStart = false;
        continue;
      }
      out += tok("p", ch); // ( ) and anything else structural
      i++;
      atWordStart = false;
    }
    return out;
  }

  // ── Rule-list languages ──────────────────────────────────────────────────
  // Ordered [class, regex] pairs, each anchored at the cursor. First match wins,
  // so order is precedence. A position that matches nothing emits one character.
  var RUST_KW =
    "as|async|await|break|const|continue|crate|dyn|else|enum|extern|false|fn|for|" +
    "if|impl|in|let|loop|match|mod|move|mut|pub|ref|return|self|static|struct|" +
    "super|trait|true|type|unsafe|use|where|while";

  var GRAMMARS = {
    rust: [
      ["c", /^(?:\/\/[^\n]*|\/\*[\s\S]*?\*\/)/],
      ["c", /^#!?\[[^\]]*\]/], // attributes read as annotation, not as code
      ["s", /^(?:r#*"[\s\S]*?"#*|"(?:[^"\\]|\\[\s\S])*")/],
      ["s", /^'(?:\\[\s\S]|[^'\\])'/], // char literal, before the lifetime rule
      ["p", /^'[A-Za-z_]\w*/], // lifetime
      ["k", new RegExp("^(?:" + RUST_KW + ")\\b")],
      ["t", /^[A-Za-z_]\w*!/], // macro invocation
      ["t", /^(?:[A-Z]\w*|Self)\b/], // type
      ["n", /^\b\d[\d_]*(?:\.\d+)?(?:[eE][+-]?\d+)?(?:[iuf](?:8|16|32|64|128|size))?/],
      ["", /^[A-Za-z_]\w*/],
      ["p", /^[{}()[\];,.:<>&|+\-*/=!?#]+/],
    ],
    toml: [
      ["c", /^#[^\n]*/],
      ["k", /^\[\[?[^\]\n]*\]\]?/], // [table] / [[array of tables]]
      ["s", /^(?:"""[\s\S]*?"""|'''[\s\S]*?'''|"(?:[^"\\]|\\[\s\S])*"|'[^'\n]*')/],
      ["n", /^\b(?:true|false)\b/],
      ["n", /^\d{4}-\d{2}-\d{2}(?:[Tt ][\d:.+\-Zz]+)?/], // date-time, before the number rule
      ["n", /^[+-]?\b\d[\d_]*(?:\.\d[\d_]*)?(?:[eE][+-]?\d+)?/],
      ["t", /^[A-Za-z_][\w-]*(?=[ \t]*=)/], // bare key
      ["", /^[A-Za-z_][\w-]*/],
      ["p", /^[=[\]{},.]+/],
    ],
    json: [
      ["t", /^"(?:[^"\\]|\\[\s\S])*"(?=[ \t]*:)/], // key
      ["s", /^"(?:[^"\\]|\\[\s\S])*"/],
      ["n", /^\b(?:true|false|null)\b/],
      ["n", /^-?\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?/],
      ["p", /^[{}[\],:]+/],
    ],
  };

  function byRules(src, rules) {
    var out = "";
    var i = 0;
    outer: while (i < src.length) {
      var rest = src.slice(i);
      // Whitespace is never a token; passing it through keeps the rules simple.
      var ws = /^\s+/.exec(rest);
      if (ws) {
        out += ws[0];
        i += ws[0].length;
        continue;
      }
      for (var r = 0; r < rules.length; r++) {
        var m = rules[r][1].exec(rest);
        if (m && m[0]) {
          out += tok(rules[r][0], m[0]);
          i += m[0].length;
          continue outer;
        }
      }
      out += esc(src[i]);
      i++;
    }
    return out;
  }

  // ── Language resolution ──────────────────────────────────────────────────
  var ALIAS = {
    sh: "shell",
    bash: "shell",
    zsh: "shell",
    shell: "shell",
    console: "shell",
    "shell-session": "shell",
    toml: "toml",
    rust: "rust",
    rs: "rust",
    json: "json",
  };

  // Box-drawing and arrow glyphs mean the block is a diagram. Absolute override.
  var DIAGRAM = /[\u2500-\u257F\u25B2\u25B6\u25BC\u25C0]/;

  /** A tagless fence is only guessed at when the guess is unambiguous. */
  function sniff(src) {
    if (/^\s*[{[]/.test(src) && /"\s*:/.test(src)) return "json";
    if (/^\s*\[[\w.-]+\]\s*$/m.test(src)) return "toml";
    if (/\b(?:fn|impl|pub struct|pub fn|let mut)\b/.test(src)) return "rust";
    if (/^\s*(?:\$ |git |cargo |npm |cd |curl |sudo |docker )/m.test(src)) return "shell";
    return "plain";
  }

  function highlight(src, lang) {
    if (DIAGRAM.test(src)) return esc(src);
    var l = ALIAS[(lang || "").toLowerCase()] || (lang ? "plain" : sniff(src));
    if (l === "shell") return shell(src);
    if (GRAMMARS[l]) return byRules(src, GRAMMARS[l]);
    return esc(src);
  }

  /** Highlights every `pre code` under `root` exactly once. */
  function apply(root) {
    var blocks = (root || document).querySelectorAll("pre code");
    for (var i = 0; i < blocks.length; i++) {
      var el = blocks[i];
      if (el.dataset.hl) continue;
      var lang = (el.className.match(/(?:language|lang)-([\w-]+)/) || [])[1] || "";
      el.innerHTML = highlight(el.textContent, lang);
      el.dataset.hl = lang || "auto";
    }
  }

  window.vunaHighlight = { apply: apply, highlight: highlight };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function () {
      apply(document);
    });
  } else {
    apply(document);
  }
})();
