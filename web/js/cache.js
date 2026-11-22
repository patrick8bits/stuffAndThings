
//
// CDN Cache Simulator
//

// -----------------------------------------
// DOM ELEMENTS
// -----------------------------------------

const objectSelect = document.getElementById("cache-object");
const fetchButton = document.getElementById("fetch-object");
const resetButton = document.getElementById("reset-cache");
const cacheStatus = document.getElementById("cache-status");
const historyBody = document.getElementById("cache-request-history");
const cacheOrderContainer = document.getElementById("cache-order");

const capacitySelect = document.getElementById("workload-capacity");
const workloadButton = document.getElementById("run-workload");
const workloadStatus = document.getElementById("workload-status");
const workloadBody = document.getElementById("workload-results");
const hitBar = document.getElementById("workload-hit-bar");

const requestHistory = [];

// -----------------------------------------
// API COMMUNICATION
// -----------------------------------------

async function requestJSON(url, options = {}) {
    const response = await fetch(url, {
        cache: "no-store",
        ...options
    });

    if (!response.ok) {
        const message = (await response.text()).trim();
        throw new Error(`HTTP ${response.status}: ${message}`);
    }

    return response.json();
}

// -----------------------------------------
// DISPLAY HELPERS
// -----------------------------------------

function setText(id, value) {
    document.getElementById(id).textContent = value;
}

function sourceLabel(source) {
    switch (source) {
        case "cache":
            return "Cache HIT";
        case "origin":
            return "Origin (MISS)";
        case "shared":
            return "Shared Fetch";
        default:
            return source;
    }
}

function sourceClass(source) {
    switch (source) {
        case "cache":
            return "cache-source-hit";
        case "origin":
            return "cache-source-miss";
        case "shared":
            return "cache-source-shared";
        default:
            return "";
    }
}

function addCell(row, value, className = "") {
    const cell = document.createElement("td");
    cell.textContent = value;

    if (className) {
        cell.classList.add(className);
    }

    row.appendChild(cell);
}

// -----------------------------------------
// CACHE STATISTICS AND HISTORY
// -----------------------------------------

function setCacheBusy(busy) {
    fetchButton.disabled = busy;
    resetButton.disabled = busy;
    objectSelect.disabled = busy;
}

function renderCacheStats(stats) {
    setText("cache-size", `${stats.size} / ${stats.capacity}`);
    setText("cache-hits", stats.hits);
    setText("cache-misses", stats.misses);
    setText("cache-evictions", stats.evictions);
}

function renderLatestRequest(result) {
    setText("cache-last-source", sourceLabel(result.source));
    setText("cache-last-latency", `${result.latency_ms} ms`);
    setText("cache-last-content", result.content);
}

function renderRequestHistory() {
    historyBody.replaceChildren();

    if (requestHistory.length === 0) {
        const row = document.createElement("tr");
        const cell = document.createElement("td");

        cell.colSpan = 3;
        cell.textContent = "No requests yet.";

        row.appendChild(cell);
        historyBody.appendChild(row);
        return;
    }

    requestHistory.forEach(result => {
        const row = document.createElement("tr");

        addCell(row, result.key);
        addCell(
            row,
            sourceLabel(result.source),
            sourceClass(result.source)
        );
        addCell(row, `${result.latency_ms} ms`);

        historyBody.appendChild(row);
    });
}

// -----------------------------------------
// LRU ORDER VISUALIZATION
// -----------------------------------------

function renderCacheOrder(keys) {
    const previousPositions = new Map();
    const existingCards = new Map();

    cacheOrderContainer
        .querySelectorAll(".cache-order-card")
        .forEach(card => {
            previousPositions.set(
                card.dataset.key,
                card.getBoundingClientRect()
            );

            existingCards.set(card.dataset.key, card);
        });

    cacheOrderContainer
        .querySelector(".cache-order-empty")
        ?.remove();

    // Remove cards for evicted objects.
    for (const [key, card] of existingCards) {
        if (!keys.includes(key)) {
            card.remove();
        }
    }

    // Create or move cards into the actual LRU order.
    keys.forEach((key, index) => {
        let card = existingCards.get(key);

        if (!card) {
            card = document.createElement("div");
            card.className = "cache-order-card";
            card.dataset.key = key;

            const name = document.createElement("strong");
            name.textContent = key;

            const label = document.createElement("small");
            card.append(name, label);
        }

        card.classList.toggle("lru-oldest", index === 0);
        card.classList.toggle(
            "lru-newest",
            index === keys.length - 1
        );

        const label = card.querySelector("small");

        if (keys.length === 1) {
            label.textContent = "Only cached object";
        } else if (index === 0) {
            label.textContent = "Evicted next";
        } else if (index === keys.length - 1) {
            label.textContent = "Most recently used";
        } else {
            label.textContent = "Cached";
        }

        cacheOrderContainer.appendChild(card);
    });

    if (keys.length === 0) {
        const empty = document.createElement("p");
        empty.className = "cache-order-empty";
        empty.textContent =
            "Cache is empty. Request an object to begin.";

        cacheOrderContainer.appendChild(empty);
    }

    // Animate existing cards as their positions change.
    const reducedMotion = window.matchMedia(
        "(prefers-reduced-motion: reduce)"
    ).matches;

    if (reducedMotion) {
        return;
    }

    cacheOrderContainer
        .querySelectorAll(".cache-order-card")
        .forEach(card => {
            const previous = previousPositions.get(
                card.dataset.key
            );

            if (!previous) {
                card.animate(
                    [
                        { opacity: 0, transform: "scale(0.9)" },
                        { opacity: 1, transform: "scale(1)" }
                    ],
                    {
                        duration: 350,
                        easing: "ease-out"
                    }
                );
                return;
            }

            const current = card.getBoundingClientRect();
            const deltaX = previous.left - current.left;
            const deltaY = previous.top - current.top;

            if (Math.abs(deltaX) > 1 || Math.abs(deltaY) > 1) {
                card.animate(
                    [
                        {
                            transform:
                                `translate(${deltaX}px, ${deltaY}px)`
                        },
                        {
                            transform: "translate(0, 0)"
                        }
                    ],
                    {
                        duration: 350,
                        easing: "ease-out"
                    }
                );
            }
        });
}

async function refreshCacheOrder() {
    const data = await requestJSON("/api/cache/order");
    renderCacheOrder(data.order);
}

// -----------------------------------------
// REQUEST CONTENT
// -----------------------------------------

async function fetchObject() {
    setCacheBusy(true);

    const key = objectSelect.value;
    cacheStatus.textContent = `Requesting ${key}...`;

    try {
        const data = await requestJSON(
            `/api/cache?key=${encodeURIComponent(key)}`
        );

        renderCacheStats(data.stats);
        renderLatestRequest(data.result);

        requestHistory.unshift(data.result);

        if (requestHistory.length > 15) {
            requestHistory.pop();
        }

        renderRequestHistory();
        await refreshCacheOrder();

        cacheStatus.textContent =
            `${key} served from ${data.result.source} ` +
            `in ${data.result.latency_ms} ms.`;

    } catch (error) {
        console.error("Cache request failed:", error);
        cacheStatus.textContent = error.message;

    } finally {
        setCacheBusy(false);
    }
}

// -----------------------------------------
// RESET CACHE
// -----------------------------------------

async function resetCache() {
    setCacheBusy(true);
    cacheStatus.textContent = "Resetting cache...";

    try {
        const data = await requestJSON(
            "/api/cache/reset",
            { method: "POST" }
        );

        renderCacheStats(data.stats);

        requestHistory.length = 0;
        renderRequestHistory();

        await refreshCacheOrder();

        setText("cache-last-source", "—");
        setText("cache-last-latency", "—");
        setText("cache-last-content", "—");

        cacheStatus.textContent =
            "Cache reset successfully. All objects removed.";

    } catch (error) {
        console.error("Cache reset failed:", error);
        cacheStatus.textContent = error.message;

    } finally {
        setCacheBusy(false);
    }
}

// -----------------------------------------
// WORKLOAD SIMULATOR
// -----------------------------------------

function renderWorkload(report) {
    const stats = report.stats;

    setText(
        "workload-hit-rate",
        `${report.hit_rate_pct.toFixed(1)}%`
    );

    setText(
        "workload-origin-requests",
        report.requests.filter(
            result => result.source === "origin"
        ).length
    );

    setText("workload-evictions", stats.evictions);

    setText(
        "workload-duration",
        `${report.total_ms.toLocaleString()} ms`
    );

    workloadBody.replaceChildren();

    report.requests.forEach((result, index) => {
        const row = document.createElement("tr");

        addCell(row, index + 1);
        addCell(row, result.key);
        addCell(
            row,
            sourceLabel(result.source),
            sourceClass(result.source)
        );
        addCell(row, `${result.latency_ms} ms`);

        workloadBody.appendChild(row);
    });

    const percentage = Math.max(
        0,
        Math.min(100, report.hit_rate_pct)
    );

    requestAnimationFrame(() => {
        requestAnimationFrame(() => {
            hitBar.style.width = `${percentage}%`;
        });
    });
}

async function runWorkload() {
    workloadButton.disabled = true;
    capacitySelect.disabled = true;

    const capacity = capacitySelect.value;

    workloadStatus.textContent =
        `Running 12 requests with capacity ${capacity}...`;

    hitBar.style.width = "0%";

    try {
        const report = await requestJSON(
            `/api/cache/workload?capacity=${capacity}`
        );

        renderWorkload(report);

        workloadStatus.textContent =
            `Completed 12 requests with ` +
            `${report.stats.hits} hits and ` +
            `${report.stats.misses} misses.`;

        console.log("Cache workload report:", report);

    } catch (error) {
        console.error("Workload failed:", error);
        workloadStatus.textContent = error.message;

    } finally {
        workloadButton.disabled = false;
        capacitySelect.disabled = false;
    }
}

// -----------------------------------------
// GO SOURCE CODE VIEWER
// -----------------------------------------

function highlightGo(source, container) {
    const pattern = /(\/\/[^\n]*|\/\*[\s\S]*?\*\/|"(?:\\.|[^"\\])*"|`[^`]*`|'(?:\\.|[^'\\])*'|\b(?:package|import|func|var|const|type|struct|interface|return|if|else|for|range|switch|case|default|select|go|defer|chan|map|break|continue)\b|\b(?:int|int64|string|bool|error|float64|byte|rune)\b|\b(?:true|false|nil)\b)/g;

    container.replaceChildren();

    let lastIndex = 0;

    for (const match of source.matchAll(pattern)) {
        container.appendChild(
            document.createTextNode(
                source.slice(lastIndex, match.index)
            )
        );

        const token = match[0];
        const span = document.createElement("span");

        span.textContent = token;

        if (token.startsWith("//") || token.startsWith("/*")) {
            span.className = "go-comment";

        } else if (/^["'`]/.test(token)) {
            span.className = "go-string";

        } else if (/^(true|false|nil)$/.test(token)) {
            span.className = "go-value";

        } else if (
            /^(int|int64|string|bool|error|float64|byte|rune)$/
                .test(token)
        ) {
            span.className = "go-type";

        } else {
            span.className = "go-keyword";
        }

        container.appendChild(span);
        lastIndex = match.index + token.length;
    }

    container.appendChild(
        document.createTextNode(source.slice(lastIndex))
    );
}

function initializeCacheSourcePanels() {
    const panels = document.querySelectorAll(".cache-source");

    panels.forEach(panel => {
        const component = panel.dataset.component;
        const code = panel.querySelector("code");

        let loaded = false;
        let loading = false;
        let source = "";

        const copyButton = document.createElement("button");
        copyButton.textContent = "Copy Source";
        copyButton.className = "copy-source";
        copyButton.disabled = true;

        panel.appendChild(copyButton);

        // Load each file only when its panel is first opened.
        panel.addEventListener("toggle", async () => {
            if (!panel.open || loaded || loading) {
                return;
            }

            loading = true;
            code.textContent = "Loading Go source...";

            try {
                const response = await fetch(
                    `/api/cache/source?component=${encodeURIComponent(component)}`,
                    { cache: "no-store" }
                );

                if (!response.ok) {
                    throw new Error(
                        `HTTP ${response.status}`
                    );
                }

                source = await response.text();

                highlightGo(source, code);

                loaded = true;
                copyButton.disabled = false;

            } catch (error) {
                code.textContent =
                    `Failed to load source: ${error.message}`;

                console.error("Source loading failed:", error);

            } finally {
                loading = false;
            }
        });

        copyButton.addEventListener("click", async () => {
            if (!loaded) {
                return;
            }

            try {
                await navigator.clipboard.writeText(source);

                copyButton.textContent = "Copied!";

                setTimeout(() => {
                    copyButton.textContent = "Copy Source";
                }, 1500);

            } catch (error) {
                copyButton.textContent = "Copy failed";
                console.error("Copy failed:", error);
            }
        });
    });
}

// -----------------------------------------
// INITIALIZATION
// -----------------------------------------

async function initializeCache() {
    try {
        const data = await requestJSON("/api/cache/stats");

        renderCacheStats(data.stats);
        renderRequestHistory();

        await refreshCacheOrder();

        cacheStatus.textContent =
            "Ready. Request an object to begin.";

    } catch (error) {
        console.error("Failed to initialize cache:", error);

        cacheStatus.textContent =
            `Failed to initialize cache: ${error.message}`;
    }
}

// -----------------------------------------
// EVENT LISTENERS
// -----------------------------------------

fetchButton.addEventListener("click", fetchObject);
resetButton.addEventListener("click", resetCache);
workloadButton.addEventListener("click", runWorkload);

initializeCacheSourcePanels();
initializeCache();
