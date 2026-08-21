<script lang="ts">
  import {
    settingsOpen,
    opacity, setOpacity,
    readerRadius, setReaderRadius, RADIUS_OPTIONS,
    backgroundMode, setBackgroundMode,
    readingWidth, setReadingWidth
  } from '../stores/ui';
  import { theme, setTheme, type ThemeMode } from '../stores/settings';

  const themes: { key: ThemeMode; label: string }[] = [
    { key: 'light', label: 'Light' },
    { key: 'dark', label: 'Dark' },
    { key: 'system', label: 'System' },
  ];

  const bgModes: { key: 'gradient' | 'solid' | 'frost'; label: string }[] = [
    { key: 'gradient', label: 'Gradient' },
    { key: 'solid', label: 'Solid' },
    { key: 'frost', label: 'Frost' },
  ];

  // AI Configuration state
  let apiKeyInput = $state('');
  let apiKeyConfigured = $state(false);
  let showApiKey = $state(false);
  let selectedModel = $state('claude-sonnet-5');
  let apiKeySaving = $state(false);
  let provider = $state('anthropic');
  let vertexProject = $state('');
  let vertexRegion = $state('us-east5');
  let vertexRegions: string[] = $state([]);

  // Translation state
  let translationLangs: string[] = $state(['es', 'en']);
  let translationDefaultIdx = $state(0);
  let newLangInput = $state('');

  const providers: { key: string; label: string }[] = [
    { key: 'anthropic', label: 'Anthropic' },
    { key: 'vertex', label: 'Vertex AI' },
  ];

  const modelTiers: { id: string; label: string }[] = [
    { id: 'claude-haiku-4-5', label: 'Haiku' },
    { id: 'claude-sonnet-5', label: 'Sonnet' },
    { id: 'claude-opus-5', label: 'Opus' },
  ];

  // Check API key and load model selection when settings panel opens
  $effect(() => {
    if ($settingsOpen) {
      checkApiKey();
      loadSettings();
    }
  });

  async function checkApiKey() {
    try {
      const App = await import('../../../wailsjs/go/main/App');
      apiKeyConfigured = await App.HasAPIKey();
    } catch {
      apiKeyConfigured = false;
    }
  }

  async function loadSettings() {
    try {
      const App = await import('../../../wailsjs/go/main/App');
      const cfg = await App.GetConfig();
      if (cfg.selectedModel) selectedModel = cfg.selectedModel;
      if (cfg.provider) provider = cfg.provider;
      if (cfg.vertexProject) vertexProject = cfg.vertexProject;
      if (cfg.vertexRegion) vertexRegion = cfg.vertexRegion;
      if (cfg.translationLanguages?.length) translationLangs = cfg.translationLanguages;
      if (cfg.translationDefaultIndex != null) translationDefaultIdx = cfg.translationDefaultIndex;
      vertexRegions = await App.GetVertexRegions();
    } catch { /* use defaults */ }
  }

  async function saveApiKey() {
    if (!apiKeyInput.trim()) return;
    apiKeySaving = true;
    try {
      const App = await import('../../../wailsjs/go/main/App');
      await App.SetAPIKey(apiKeyInput.trim());
      apiKeyConfigured = true;
      apiKeyInput = '';
    } catch (err) {
      console.error('Failed to save API key:', err);
    }
    apiKeySaving = false;
  }

  async function clearApiKey() {
    try {
      const App = await import('../../../wailsjs/go/main/App');
      await App.DeleteAPIKey();
      apiKeyConfigured = false;
      apiKeyInput = '';
    } catch (err) {
      console.error('Failed to delete API key:', err);
    }
  }

  async function setModel(modelId: string) {
    selectedModel = modelId;
    try {
      const App = await import('../../../wailsjs/go/main/App');
      const cfg = await App.GetConfig();
      cfg.selectedModel = modelId;
      await App.UpdateConfig(cfg);
    } catch (err) {
      console.error('Failed to save model:', err);
    }
  }

  async function setProvider(p: string) {
    provider = p;
    try {
      const App = await import('../../../wailsjs/go/main/App');
      const cfg = await App.GetConfig();
      cfg.provider = p;
      await App.UpdateConfig(cfg);
    } catch (err) {
      console.error('Failed to save provider:', err);
    }
  }

  async function saveVertexConfig() {
    try {
      const App = await import('../../../wailsjs/go/main/App');
      const cfg = await App.GetConfig();
      cfg.vertexProject = vertexProject.trim();
      cfg.vertexRegion = vertexRegion.trim();
      await App.UpdateConfig(cfg);
    } catch (err) {
      console.error('Failed to save Vertex config:', err);
    }
  }

  async function saveTranslationConfig() {
    try {
      const App = await import('../../../wailsjs/go/main/App');
      const cfg = await App.GetConfig();
      cfg.translationLanguages = translationLangs;
      cfg.translationDefaultIndex = translationDefaultIdx;
      await App.UpdateConfig(cfg);
    } catch (err) {
      console.error('Failed to save translation config:', err);
    }
  }

  function addLanguage() {
    const lang = newLangInput.trim().toLowerCase();
    if (!lang || translationLangs.includes(lang)) return;
    translationLangs = [...translationLangs, lang];
    newLangInput = '';
    saveTranslationConfig();
  }

  function removeLanguage(idx: number) {
    if (translationLangs.length <= 1) return;
    translationLangs = translationLangs.filter((_, i) => i !== idx);
    if (translationDefaultIdx >= translationLangs.length) {
      translationDefaultIdx = 0;
    } else if (translationDefaultIdx > idx) {
      translationDefaultIdx--;
    }
    saveTranslationConfig();
  }

  function moveLanguage(idx: number, dir: -1 | 1) {
    const newIdx = idx + dir;
    if (newIdx < 0 || newIdx >= translationLangs.length) return;
    const copy = [...translationLangs];
    [copy[idx], copy[newIdx]] = [copy[newIdx], copy[idx]];
    translationLangs = copy;
    if (translationDefaultIdx === idx) translationDefaultIdx = newIdx;
    else if (translationDefaultIdx === newIdx) translationDefaultIdx = idx;
    saveTranslationConfig();
  }

  function setDefaultLanguage(idx: number) {
    translationDefaultIdx = idx;
    saveTranslationConfig();
  }

  function handleWidthInput(e: Event) {
    const target = e.target as HTMLInputElement;
    setReadingWidth(parseInt(target.value));
  }

  function handleOpacityInput(e: Event) {
    const target = e.target as HTMLInputElement;
    setOpacity(parseInt(target.value));
  }
</script>

{#if $settingsOpen}
  <div class="settings" role="dialog" aria-label="Appearance settings">
    <div class="sp-title">
      <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="2.5"/><path d="M10 2v2m0 12v2M3.5 5l1.5 1m10 8l1.5 1M2 10h2m12 0h2M3.5 15l1.5-1m10-8l1.5-1"/></svg>
      Appearance
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="7"/><path d="M10 3a7 7 0 000 14" fill="var(--icon-stroke)" opacity=".15"/></svg>
        Theme
      </span>
      <div class="theme-sel" role="radiogroup" aria-label="Theme selection">
        {#each themes as t (t.key)}
          <button
            class="th-opt"
            class:on={$theme === t.key}
            role="radio"
            aria-checked={$theme === t.key}
            onclick={() => setTheme(t.key)}
          >{t.label}</button>
        {/each}
      </div>
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><rect x="3" y="5" width="14" height="10" rx="1"/><line x1="7" y1="5" x2="7" y2="15" opacity=".3"/><line x1="13" y1="5" x2="13" y2="15" opacity=".3"/></svg>
        Reading Width
      </span>
      <input
        type="range"
        class="w-slider"
        min="600"
        max="1000"
        step="20"
        value={$readingWidth}
        oninput={handleWidthInput}
        aria-label="Reading width"
      />
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="7"/><circle cx="10" cy="10" r="4" opacity=".3"/></svg>
        Opacity
      </span>
      <div class="sr-val-group">
        <input
          type="range"
          class="w-slider"
          min="40"
          max="100"
          step="5"
          value={$opacity}
          oninput={handleOpacityInput}
          aria-label="Surface opacity"
        />
        <span class="cb-val">{$opacity}%</span>
      </div>
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><rect x="3" y="3" width="14" height="14" rx="4"/></svg>
        Window Radius
      </span>
      <div class="radius-sel" role="radiogroup" aria-label="Window corner radius">
        {#each RADIUS_OPTIONS as r (r)}
          <button
            class="rad-opt"
            class:on={$readerRadius === r}
            role="radio"
            aria-checked={$readerRadius === r}
            onclick={() => setReaderRadius(r)}
          >{r}</button>
        {/each}
      </div>
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><path d="M2 14l4-5 3 3 4-6 5 8"/><rect x="2" y="2" width="16" height="16" rx="2"/></svg>
        Background
      </span>
      <div class="bg-sel" role="radiogroup" aria-label="Background mode">
        {#each bgModes as bg (bg.key)}
          <button
            class="bg-opt"
            class:on={$backgroundMode === bg.key}
            role="radio"
            aria-checked={$backgroundMode === bg.key}
            onclick={() => setBackgroundMode(bg.key)}
          >{bg.label}</button>
        {/each}
      </div>
    </div>

    <div class="settings-divider" aria-hidden="true"></div>

    <div class="sp-title">
      <svg viewBox="0 0 20 20"><path d="M10 2l1.5 3.5L15 7l-3.5 1.5L10 12l-1.5-3.5L5 7l3.5-1.5z"/><path d="M15 12l1 2 2 1-2 1-1 2-1-2-2-1 2-1z" opacity=".5"/></svg>
      AI
      <span
        class="status-dot"
        class:configured={provider === 'vertex' || apiKeyConfigured}
        title={provider === 'vertex' ? 'Vertex AI (ADC)' : apiKeyConfigured ? 'Connected' : 'No API key'}
      ></span>
    </div>

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><path d="M3 10a7 7 0 1114 0 7 7 0 01-14 0z"/><path d="M10 3v14M3 10h14" opacity=".3"/></svg>
        Provider
      </span>
      <div class="provider-sel" role="radiogroup" aria-label="Provider selection">
        {#each providers as p (p.key)}
          <button
            class="prov-opt"
            class:on={provider === p.key}
            role="radio"
            aria-checked={provider === p.key}
            onclick={() => setProvider(p.key)}
          >{p.label}</button>
        {/each}
      </div>
    </div>

    {#if provider === 'anthropic'}
      <div class="sr">
        <span class="sr-label">
          <svg viewBox="0 0 20 20"><rect x="3" y="9" width="14" height="8" rx="2"/><path d="M6 9V6a4 4 0 018 0v3"/></svg>
          API Key
        </span>
        <div class="api-key-group">
          {#if apiKeyConfigured && !apiKeyInput}
            <span class="api-key-status">Configured</span>
            <button class="api-key-clear" onclick={clearApiKey} title="Remove API key" aria-label="Remove API key">
              <svg viewBox="0 0 14 14"><line x1="3" y1="3" x2="11" y2="11"/><line x1="11" y1="3" x2="3" y2="11"/></svg>
            </button>
          {:else}
            <input
              class="api-key-input"
              type={showApiKey ? 'text' : 'password'}
              bind:value={apiKeyInput}
              placeholder="sk-ant-..."
              autocomplete="off"
              spellcheck="false"
              aria-label="Anthropic API key"
            />
            <button class="api-key-toggle" onclick={() => showApiKey = !showApiKey} title={showApiKey ? 'Hide key' : 'Show key'} aria-label={showApiKey ? 'Hide API key' : 'Show API key'}>
              <svg viewBox="0 0 14 14">
                {#if showApiKey}
                  <path d="M1 7s2.5-4 6-4 6 4 6 4-2.5 4-6 4-6-4-6-4z"/><circle cx="7" cy="7" r="2"/>
                {:else}
                  <path d="M1 7s2.5-4 6-4 6 4 6 4-2.5 4-6 4-6-4-6-4z"/><circle cx="7" cy="7" r="2"/><line x1="2" y1="12" x2="12" y2="2"/>
                {/if}
              </svg>
            </button>
            <button class="api-key-save" onclick={saveApiKey} disabled={!apiKeyInput.trim() || apiKeySaving} aria-label="Save API key">
              Save
            </button>
          {/if}
        </div>
      </div>
    {:else}
      <div class="sr vertex-field">
        <span class="sr-label">
          <svg viewBox="0 0 20 20"><path d="M4 4h12v12H4z" fill="none"/><path d="M10 4v12M4 10h12" opacity=".3"/></svg>
          Project ID
        </span>
        <input
          class="vertex-input"
          type="text"
          bind:value={vertexProject}
          placeholder="my-gcp-project"
          onblur={saveVertexConfig}
          aria-label="GCP Project ID"
        />
      </div>
      <div class="sr vertex-field">
        <span class="sr-label">
          <svg viewBox="0 0 20 20"><circle cx="10" cy="10" r="7" fill="none"/><path d="M10 3v14" opacity=".3"/></svg>
          Region
        </span>
        <select
          class="vertex-select"
          bind:value={vertexRegion}
          onchange={saveVertexConfig}
          aria-label="Vertex AI Region"
        >
          {#each vertexRegions as r (r)}
            <option value={r}>{r}</option>
          {/each}
        </select>
      </div>
    {/if}

    <div class="sr">
      <span class="sr-label">
        <svg viewBox="0 0 20 20"><rect x="3" y="3" width="14" height="14" rx="3"/><circle cx="7" cy="10" r="1.5"/><circle cx="13" cy="10" r="1.5"/><path d="M7 7h6"/></svg>
        Model
      </span>
      <div class="model-sel" role="radiogroup" aria-label="Model selection">
        {#each modelTiers as m (m.id)}
          <button
            class="model-opt"
            class:on={selectedModel === m.id}
            role="radio"
            aria-checked={selectedModel === m.id}
            onclick={() => setModel(m.id)}
            title={m.id}
          >{m.label}</button>
        {/each}
      </div>
    </div>

    <div class="ai-help-text">
      {#if provider === 'vertex'}
        Uses Google Cloud Application Default Credentials. Run <code>gcloud auth application-default login</code>.
      {:else}
        Your API key is stored securely in your OS keychain. It is never saved to the config file.
      {/if}
    </div>

    <div class="settings-divider" aria-hidden="true"></div>

    <div class="sp-title">
      <svg viewBox="0 0 20 20"><path d="M3 5h14M3 10h10M3 15h6"/></svg>
      Translation
    </div>

    <div class="tl-langs" role="list" aria-label="Translation languages">
      {#each translationLangs as lang, i (lang)}
        <div class="tl-lang" class:tl-default={i === translationDefaultIdx} role="listitem">
          <button
            class="tl-star"
            class:active={i === translationDefaultIdx}
            onclick={() => setDefaultLanguage(i)}
            title={i === translationDefaultIdx ? 'Default language' : 'Set as default'}
            aria-label={i === translationDefaultIdx ? `${lang} is default` : `Set ${lang} as default`}
          >
            <svg viewBox="0 0 12 12">
              {#if i === translationDefaultIdx}
                <circle cx="6" cy="6" r="4" fill="var(--accent-solid)" stroke="none"/>
              {:else}
                <circle cx="6" cy="6" r="3.5" fill="none"/>
              {/if}
            </svg>
          </button>
          <span class="tl-code">{lang.toUpperCase()}</span>
          <div class="tl-actions">
            <button
              class="tl-move"
              onclick={() => moveLanguage(i, -1)}
              disabled={i === 0}
              title="Move up"
              aria-label={`Move ${lang} up`}
            >
              <svg viewBox="0 0 10 10"><path d="M2 6l3-3 3 3"/></svg>
            </button>
            <button
              class="tl-move"
              onclick={() => moveLanguage(i, 1)}
              disabled={i === translationLangs.length - 1}
              title="Move down"
              aria-label={`Move ${lang} down`}
            >
              <svg viewBox="0 0 10 10"><path d="M2 4l3 3 3-3"/></svg>
            </button>
            <button
              class="tl-remove"
              onclick={() => removeLanguage(i)}
              disabled={translationLangs.length <= 1}
              title="Remove"
              aria-label={`Remove ${lang}`}
            >
              <svg viewBox="0 0 10 10"><line x1="2" y1="2" x2="8" y2="8"/><line x1="8" y1="2" x2="2" y2="8"/></svg>
            </button>
          </div>
        </div>
      {/each}
    </div>

    <div class="tl-add">
      <input
        class="tl-add-input"
        type="text"
        bind:value={newLangInput}
        placeholder="Add language (e.g. fr)"
        maxlength="10"
        onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter') addLanguage(); }}
        aria-label="New language code"
      />
      <button
        class="tl-add-btn"
        onclick={addLanguage}
        disabled={!newLangInput.trim() || translationLangs.includes(newLangInput.trim().toLowerCase())}
        aria-label="Add language"
      >+</button>
    </div>

    <div class="ai-help-text">
      Select text and press <code>T</code> to translate. Press <code>T</code> again to cycle languages. <code>Ctrl+Space</code> to switch default.
    </div>
  </div>
{/if}

<style>
  .settings {
    position: absolute;
    bottom: 60px;
    left: 50%;
    transform: translateX(-50%);
    background: var(--surface-elevated);
    backdrop-filter: blur(40px);
    -webkit-backdrop-filter: blur(40px);
    border: 1px solid var(--border);
    border-radius: 18px;
    padding: 20px;
    width: 360px;
    z-index: 35;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
    user-select: none;
  }

  .sp-title {
    font-size: 13px;
    font-weight: 600;
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .sp-title svg {
    width: 16px;
    height: 16px;
    stroke: var(--icon-stroke);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .sr {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 9px 0;
    border-bottom: 1px solid var(--border);
  }

  .sr:last-child {
    border-bottom: 0;
  }

  .sr-label {
    font-size: 13px;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: 7px;
  }

  .sr-label svg {
    width: 15px;
    height: 15px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .sr-val-group {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .cb-val {
    font-size: 11px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-secondary);
    min-width: 32px;
    text-align: right;
  }

  .theme-sel {
    display: flex;
    gap: 4px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .th-opt {
    padding: 5px 12px;
    font-size: 12px;
    color: var(--text-tertiary);
    border-radius: 7px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: inherit;
    transition: background 0.12s, color 0.12s;
  }

  .th-opt:hover {
    color: var(--text-secondary);
  }

  .th-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  .radius-sel {
    display: flex;
    gap: 3px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .rad-opt {
    padding: 4px 8px;
    font-size: 11px;
    color: var(--text-tertiary);
    border-radius: 6px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: 'JetBrains Mono', monospace;
    transition: background 0.12s, color 0.12s;
  }

  .rad-opt:hover {
    color: var(--text-secondary);
  }

  .rad-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  .bg-sel {
    display: flex;
    gap: 3px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .bg-opt {
    padding: 4px 8px;
    font-size: 11px;
    color: var(--text-tertiary);
    border-radius: 6px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: inherit;
    transition: background 0.12s, color 0.12s;
  }

  .bg-opt:hover {
    color: var(--text-secondary);
  }

  .bg-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  .w-slider {
    width: 80px;
    height: 4px;
    -webkit-appearance: none;
    appearance: none;
    background: var(--border);
    border-radius: 2px;
    outline: 0;
    cursor: pointer;
  }

  .w-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 14px;
    height: 14px;
    background: var(--accent-solid);
    border-radius: 50%;
    cursor: pointer;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.3);
  }

  .w-slider::-moz-range-thumb {
    width: 14px;
    height: 14px;
    background: var(--accent-solid);
    border-radius: 50%;
    cursor: pointer;
    border: 0;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.3);
  }

  /* Section divider (Design.md) */
  .settings-divider {
    height: 1px;
    background: var(--border);
    margin: 12px 0;
  }

  /* Status dot (Design.md) */
  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--warning);
    flex-shrink: 0;
  }

  .status-dot.configured {
    background: var(--success);
  }

  /* API Key group */
  .api-key-group {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .api-key-input {
    width: 140px;
    padding: 5px 10px;
    font-size: 12px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-primary);
    background: var(--hover-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    outline: none;
    transition: border-color 0.12s;
  }

  .api-key-input:focus {
    border-color: var(--border-focus);
  }

  .api-key-input::placeholder {
    color: var(--text-ghost);
  }

  .api-key-toggle {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: none;
    cursor: pointer;
    border-radius: 4px;
    padding: 0;
    transition: background 0.12s;
  }

  .api-key-toggle:hover {
    background: var(--hover-bg);
  }

  .api-key-toggle svg {
    width: 14px;
    height: 14px;
    stroke: var(--text-tertiary);
    stroke-width: 1.2;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .api-key-save {
    padding: 4px 10px;
    font-size: 11px;
    font-family: inherit;
    color: var(--text-primary);
    background: var(--accent-dim);
    border: 1px solid var(--border);
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.12s;
  }

  .api-key-save:hover:not(:disabled) {
    background: var(--active-bg);
  }

  .api-key-save:disabled {
    opacity: 0.35;
    cursor: default;
  }

  .api-key-status {
    font-size: 11px;
    color: var(--success);
    font-family: 'JetBrains Mono', monospace;
  }

  .api-key-clear {
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: none;
    cursor: pointer;
    border-radius: 4px;
    padding: 0;
    transition: background 0.12s;
  }

  .api-key-clear:hover {
    background: rgba(239, 68, 68, 0.12);
  }

  .api-key-clear svg {
    width: 10px;
    height: 10px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    fill: none;
  }

  .api-key-clear:hover svg {
    stroke: var(--danger);
  }

  /* Provider selector — reuses theme selector pattern */
  .provider-sel {
    display: flex;
    gap: 4px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .prov-opt {
    padding: 5px 12px;
    font-size: 12px;
    color: var(--text-tertiary);
    border-radius: 7px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: inherit;
    transition: background 0.12s, color 0.12s;
  }

  .prov-opt:hover {
    color: var(--text-secondary);
  }

  .prov-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  /* Vertex AI config inputs */
  .vertex-input {
    width: 140px;
    padding: 5px 10px;
    font-size: 12px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-primary);
    background: var(--hover-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    outline: none;
    transition: border-color 0.12s;
  }

  .vertex-input:focus {
    border-color: var(--border-focus);
  }

  .vertex-input::placeholder {
    color: var(--text-ghost);
  }

  .vertex-select {
    width: 140px;
    padding: 5px 10px;
    font-size: 12px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-primary);
    background: var(--hover-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    outline: none;
    cursor: pointer;
    transition: border-color 0.12s;
    -webkit-appearance: none;
    appearance: none;
  }

  .vertex-select:focus {
    border-color: var(--border-focus);
  }

  .ai-help-text code {
    font-family: 'JetBrains Mono', monospace;
    font-size: 10px;
    background: var(--hover-bg);
    padding: 1px 4px;
    border-radius: 3px;
  }

  /* Model selector — reuses theme selector pattern */
  .model-sel {
    display: flex;
    gap: 4px;
    background: var(--hover-bg);
    border-radius: 9px;
    padding: 3px;
  }

  .model-opt {
    padding: 5px 12px;
    font-size: 12px;
    color: var(--text-tertiary);
    border-radius: 7px;
    cursor: pointer;
    border: none;
    background: 0;
    font-family: inherit;
    transition: background 0.12s, color 0.12s;
  }

  .model-opt:hover {
    color: var(--text-secondary);
  }

  .model-opt.on {
    background: var(--surface-solid);
    color: var(--text-primary);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }

  .ai-help-text {
    font-size: 11px;
    color: var(--text-ghost);
    padding: 8px 0 0;
    line-height: 1.4;
  }

  /* Translation language list */
  .tl-langs {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-bottom: 8px;
  }

  .tl-lang {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 5px 8px;
    border-radius: 8px;
    transition: background 0.12s;
  }

  .tl-lang:hover {
    background: var(--hover-bg);
  }

  .tl-lang.tl-default {
    background: var(--accent-dim);
  }

  .tl-star {
    width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: none;
    cursor: pointer;
    padding: 0;
    flex-shrink: 0;
  }

  .tl-star svg {
    width: 12px;
    height: 12px;
    stroke: var(--text-tertiary);
    stroke-width: 1.2;
    fill: none;
  }

  .tl-star.active svg {
    stroke: var(--accent-solid);
  }

  .tl-code {
    font-size: 12px;
    font-family: 'JetBrains Mono', monospace;
    font-weight: 600;
    color: var(--text-primary);
    flex: 1;
  }

  .tl-actions {
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.12s;
  }

  .tl-lang:hover .tl-actions {
    opacity: 1;
  }

  .tl-move, .tl-remove {
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: none;
    cursor: pointer;
    border-radius: 4px;
    padding: 0;
    transition: background 0.12s;
  }

  .tl-move:hover:not(:disabled) {
    background: var(--hover-bg);
  }

  .tl-remove:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.12);
  }

  .tl-move:disabled, .tl-remove:disabled {
    opacity: 0.25;
    cursor: default;
  }

  .tl-move svg {
    width: 10px;
    height: 10px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    fill: none;
  }

  .tl-remove svg {
    width: 8px;
    height: 8px;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
    fill: none;
  }

  .tl-remove:hover:not(:disabled) svg {
    stroke: var(--danger);
  }

  .tl-add {
    display: flex;
    gap: 4px;
  }

  .tl-add-input {
    flex: 1;
    padding: 5px 10px;
    font-size: 12px;
    font-family: 'JetBrains Mono', monospace;
    color: var(--text-primary);
    background: var(--hover-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    outline: none;
    transition: border-color 0.12s;
  }

  .tl-add-input:focus {
    border-color: var(--border-focus);
  }

  .tl-add-input::placeholder {
    color: var(--text-ghost);
    font-family: inherit;
  }

  .tl-add-btn {
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-family: inherit;
    color: var(--text-secondary);
    background: var(--hover-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.12s, color 0.12s;
  }

  .tl-add-btn:hover:not(:disabled) {
    background: var(--accent-dim);
    color: var(--text-primary);
  }

  .tl-add-btn:disabled {
    opacity: 0.3;
    cursor: default;
  }
</style>
