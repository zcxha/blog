(() => {
  "use strict";

  const storageKey = (path) => `folio:liked:${path}`;
  let counterSequence = 0;

  const appendCount = (element, path) => {
    const id = `folio-counter-${counterSequence++}`;
    element.id = id;
    element.textContent = "";
    window.goatcounter.visit_count({ append: `#${id}`, type: "html", ...(path ? { path } : {}) });
  };

  const loadVisibleCounts = (attempt = 0) => {
    if (!window.goatcounter || typeof window.goatcounter.visit_count !== "function") {
      if (attempt < 40) window.setTimeout(() => loadVisibleCounts(attempt + 1), 125);
      return;
    }
    document.querySelectorAll("[data-site-view-count]").forEach((element) => appendCount(element, "TOTAL"));
    document.querySelectorAll("[data-page-view-count]").forEach((element) => appendCount(element));
    document.querySelectorAll("[data-like-count]").forEach((element) => appendCount(element, `like:${element.dataset.likePath}`));
  };

  loadVisibleCounts();

  document.querySelectorAll("[data-like-button]").forEach((button) => {
    const path = button.dataset.likePath || window.location.pathname;
    let liked = false;

    try {
      liked = window.localStorage.getItem(storageKey(path)) === "1";
    } catch (_) {
      // Private browsing modes may deny storage; likes still remain clickable.
    }

    const render = () => {
      button.classList.toggle("is-liked", liked);
      button.setAttribute("aria-pressed", String(liked));
      button.querySelector("[data-like-icon]").textContent = liked ? "♥" : "♡";
      button.querySelector("[data-like-label]").textContent = liked ? "已喜欢" : "喜欢这篇文章";
    };
    render();

    button.addEventListener("click", () => {
      if (liked) return;
      liked = true;
      try {
        window.localStorage.setItem(storageKey(path), "1");
      } catch (_) {
        // The anonymous event below is the durable source of truth.
      }
      render();

      if (window.goatcounter && typeof window.goatcounter.count === "function") {
        window.goatcounter.count({
          path: `like:${path}`,
          title: `Like: ${document.title}`,
          event: true,
        });
      }
      document.querySelectorAll(`[data-like-count][data-like-path="${CSS.escape(path)}"]`).forEach((element) => {
        const current = Number.parseInt(element.textContent.replace(/\D/g, ""), 10);
        if (Number.isFinite(current)) element.textContent = String(current + 1);
      });
    });
  });
})();
