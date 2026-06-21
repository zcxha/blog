(() => {
  "use strict";

  const storageKey = (path) => `folio:liked:${path}`;
  const countRequests = new Map();

  const goatCounterBaseURL = () => {
    const script = document.querySelector("script[data-goatcounter]");
    const endpoint = script?.dataset.goatcounter || window.goatcounter?.endpoint || "";
    return endpoint.replace(/\/count\/?$/, "/counter/");
  };

  const currentPagePath = () => {
    const canonical = document.querySelector('link[rel="canonical"][href]');
    return canonical ? new URL(canonical.href, window.location.href).pathname : window.location.pathname;
  };

  const fetchCount = (path) => {
    if (countRequests.has(path)) return countRequests.get(path);

    const baseURL = goatCounterBaseURL();
    const request = baseURL
      ? fetch(`${baseURL}${encodeURIComponent(path)}.json`, { credentials: "omit" }).then(async (response) => {
          if (response.status === 404) return "0";
          if (!response.ok) throw new Error(`GoatCounter returned ${response.status}`);
          const data = await response.json();
          return String(data.count ?? data.count_unique ?? "0");
        })
      : Promise.reject(new Error("GoatCounter endpoint is not configured"));

    countRequests.set(path, request);
    return request;
  };

  const showCount = async (element, path) => {
    element.textContent = "…";
    try {
      element.textContent = await fetchCount(path);
    } catch (error) {
      console.warn("folio: unable to load GoatCounter count", error);
      element.textContent = "—";
    }
  };

  const loadVisibleCounts = () => {
    document.querySelectorAll("[data-site-view-count]").forEach((element) => showCount(element, "TOTAL"));
    document.querySelectorAll("[data-page-view-count]").forEach((element) => showCount(element, currentPagePath()));
    document.querySelectorAll("[data-like-count]").forEach((element) => showCount(element, `like:${element.dataset.likePath}`));
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
