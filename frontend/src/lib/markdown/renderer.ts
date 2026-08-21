import MarkdownIt from 'markdown-it';
import hljs from 'highlight.js/lib/core';
import go from 'highlight.js/lib/languages/go';
import typescript from 'highlight.js/lib/languages/typescript';
import javascript from 'highlight.js/lib/languages/javascript';
import python from 'highlight.js/lib/languages/python';
import bash from 'highlight.js/lib/languages/bash';
import json from 'highlight.js/lib/languages/json';
import yaml from 'highlight.js/lib/languages/yaml';
import css from 'highlight.js/lib/languages/css';
import xml from 'highlight.js/lib/languages/xml';
import sql from 'highlight.js/lib/languages/sql';
import markdown from 'highlight.js/lib/languages/markdown';
import dockerfile from 'highlight.js/lib/languages/dockerfile';
import rust from 'highlight.js/lib/languages/rust';
import java from 'highlight.js/lib/languages/java';
import makefile from 'highlight.js/lib/languages/makefile';

hljs.registerLanguage('go', go);
hljs.registerLanguage('typescript', typescript);
hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('python', python);
hljs.registerLanguage('bash', bash);
hljs.registerLanguage('shell', bash);
hljs.registerLanguage('json', json);
hljs.registerLanguage('yaml', yaml);
hljs.registerLanguage('yml', yaml);
hljs.registerLanguage('css', css);
hljs.registerLanguage('html', xml);
hljs.registerLanguage('xml', xml);
hljs.registerLanguage('sql', sql);
hljs.registerLanguage('markdown', markdown);
hljs.registerLanguage('dockerfile', dockerfile);
hljs.registerLanguage('rust', rust);
hljs.registerLanguage('java', java);
hljs.registerLanguage('makefile', makefile);

const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  highlight: (str: string, lang: string): string => {
    const copyBtn = `<button class="code-copy-btn" aria-label="Copy code block" title="Copy code"><svg viewBox="0 0 20 20"><rect x="6" y="6" width="10" height="12" rx="1.5"/><path d="M4 14V4a1.5 1.5 0 011.5-1.5H13"/></svg></button>`;
    if (lang && hljs.getLanguage(lang)) {
      try {
        const result = hljs.highlight(str, { language: lang, ignoreIllegals: true });
        const escapedLang = md.utils.escapeHtml(lang);
        return `<pre class="code-block"><div class="code-lang">${escapedLang}</div>${copyBtn}<code class="hljs language-${escapedLang}">${result.value}</code></pre>`;
      } catch (_) { /* fall through */ }
    }
    const escaped = md.utils.escapeHtml(str);
    return `<pre class="code-block">${copyBtn}<code class="hljs">${escaped}</code></pre>`;
  }
});

// ── Heading IDs (Task 1.5b) ──
function slugify(text: string): string {
  return text.toLowerCase().replace(/[^\w\s-]/g, '').replace(/\s+/g, '-').replace(/-+/g, '-').trim();
}

md.core.ruler.push('heading_ids', (state) => {
  for (const token of state.tokens) {
    if (token.type === 'heading_open') {
      const nextToken = state.tokens[state.tokens.indexOf(token) + 1];
      if (nextToken && nextToken.type === 'inline' && nextToken.content) {
        token.attrSet('id', slugify(nextToken.content));
      }
    }
  }
});

// ── External Link Indicators (Task 1.6) ──
const defaultLinkOpen = md.renderer.rules.link_open || function(tokens: any[], idx: number, options: any, _env: any, self: any) {
  return self.renderToken(tokens, idx, options);
};

md.renderer.rules.link_open = function(tokens: any[], idx: number, options: any, env: any, self: any) {
  const href = String(tokens[idx].attrGet('href') ?? '');
  if (href && (href.startsWith('http://') || href.startsWith('https://'))) {
    tokens[idx].attrSet('data-external', 'true');
    tokens[idx].attrSet('target', '_blank');
    tokens[idx].attrSet('rel', 'noopener noreferrer');
  }
  return defaultLinkOpen(tokens, idx, options, env, self);
};

// ── Image URL Rewriting (Task 2.2) ──
const defaultImageRule = md.renderer.rules.image || function(tokens: any[], idx: number, options: any, _env: any, self: any) {
  return self.renderToken(tokens, idx, options);
};

md.renderer.rules.image = function(tokens: any[], idx: number, options: any, env: any, self: any) {
  const token = tokens[idx];
  const src = String(token.attrGet('src') ?? '');

  if (src && !src.startsWith('http://') && !src.startsWith('https://') && !src.startsWith('data:')) {
    const basePath: string = (env && env.basePath) ? String(env.basePath) : '';
    let resolvedPath: string;
    if (src.startsWith('/')) {
      resolvedPath = src;
    } else if (basePath) {
      const dir = basePath.includes('/') ? basePath.substring(0, basePath.lastIndexOf('/')) : '';
      resolvedPath = dir ? `${dir}/${src}` : src;
    } else {
      resolvedPath = src;
    }
    // Normalize path segments
    const parts = resolvedPath.split('/');
    const normalized: string[] = [];
    for (const p of parts) {
      if (p === '..') normalized.pop();
      else if (p !== '.' && p !== '') normalized.push(p);
    }
    token.attrSet('src', `/local/${normalized.join('/')}`);
  }

  return defaultImageRule(tokens, idx, options, env, self);
};

export function renderMarkdown(source: string, basePath?: string): string {
  return md.render(source, { basePath });
}
