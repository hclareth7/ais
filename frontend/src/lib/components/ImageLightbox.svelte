<script lang="ts">
  let { src = '', alt = '', open = false, onclose }: {
    src?: string;
    alt?: string;
    open?: boolean;
    onclose?: () => void;
  } = $props();

  function handleClose() {
    onclose?.();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      handleClose();
    }
  }

  function handleScroll() {
    handleClose();
  }

  $effect(() => {
    if (open) {
      document.addEventListener('keydown', handleKeydown, true);
      window.addEventListener('scroll', handleScroll, true);
      return () => {
        document.removeEventListener('keydown', handleKeydown, true);
        window.removeEventListener('scroll', handleScroll, true);
      };
    }
  });
</script>

{#if open}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="lightbox-overlay"
    onclick={handleClose}
    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleClose(); }}
    role="dialog"
    aria-label="Image preview: {alt}"
    aria-modal="true"
    tabindex="-1"
  >
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <img
      src={src}
      alt={alt}
      class="lightbox-img"
      onclick={(e) => { e.stopPropagation(); handleClose(); }}
      onkeydown={() => {}}
    />
  </div>
{/if}

<style>
  .lightbox-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.85);
    z-index: 300;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    animation: lightbox-in 120ms ease-out;
  }

  .lightbox-img {
    max-width: 90vw;
    max-height: 85vh;
    border-radius: 16px;
    object-fit: contain;
    cursor: pointer;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  @keyframes lightbox-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .lightbox-overlay {
      animation: none;
    }
  }
</style>
