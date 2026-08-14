<script lang="ts">
</script>

<div class="welcome" role="article" aria-label="Welcome">
  <div class="welcome-content">
    <h1 class="welcome-title">ais</h1>
    <p class="welcome-subtitle">Markdown visualizer</p>

    <div class="shortcuts">
      <h2 class="shortcuts-title">Keyboard Shortcuts</h2>
      <div class="shortcut-list">
        <div class="shortcut">
          <kbd>Ctrl</kbd> + <kbd>B</kbd>
          <span>Toggle sidebar</span>
        </div>
        <div class="shortcut">
          <kbd>Ctrl</kbd> + <kbd>Tab</kbd>
          <span>Next tab</span>
        </div>
        <div class="shortcut">
          <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Tab</kbd>
          <span>Previous tab</span>
        </div>
        <div class="shortcut">
          <kbd>Ctrl</kbd> + <kbd>W</kbd>
          <span>Close tab</span>
        </div>
      </div>
    </div>

    <p class="welcome-hint">Open a file from the sidebar to begin reading</p>
  </div>
</div>

<style>
  .welcome {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    padding: 32px;
  }

  .welcome-content {
    text-align: center;
    max-width: 400px;
  }

  .welcome-title {
    font-size: 48px;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0 0 8px;
    letter-spacing: -1px;
  }

  .welcome-subtitle {
    font-size: 16px;
    color: var(--text-muted);
    margin: 0 0 48px;
  }

  .shortcuts {
    text-align: left;
    margin-bottom: 48px;
  }

  .shortcuts-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 0 0 16px;
  }

  .shortcut-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .shortcut {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: var(--text-secondary);
  }

  .shortcut span {
    margin-left: auto;
  }

  kbd {
    display: inline-block;
    padding: 2px 8px;
    font-size: 12px;
    font-family: 'Inter', sans-serif;
    background: var(--bg-elevated);
    border: 1px solid var(--border-strong);
    border-radius: 4px;
    color: var(--text-secondary);
  }

  .welcome-hint {
    font-size: 14px;
    color: var(--text-muted);
    margin: 0;
  }
</style>
