const sequentialButton = document.getElementById("run-sequential");
const concurrentButton = document.getElementById("run-concurrent");
const compareButton = document.getElementById("compare-both");

const healthStatus = document.getElementById("health-status");
const healthResults = document.getElementById("health-results");

const buttons = [
    sequentialButton,
    concurrentButton,
    compareButton
];

const executionTimes = {
    sequential: null,
    concurrent: null
};

// Enable or disable monitoring controls.
function setRunning(running) {
    buttons.forEach(button => {
        button.disabled = running;
    });
}

// Request a health report from the Go backend.
async function fetchHealthReport(mode) {
    const response = await fetch(`/api/health?mode=${mode}`);

    if (!response.ok) {
        throw new Error(await response.text());
    }

    return await response.json();
}

// Update the dashboard summary cards.
function renderSummary(report) {
    document.getElementById("total-servers")
        .textContent = report.total;

    document.getElementById("healthy-servers")
        .textContent = report.healthy;

    document.getElementById("unhealthy-servers")
        .textContent = report.unhealthy;
}

// Populate the server health table.
function renderResults(report) {
    healthResults.replaceChildren();

    const maxLatency = Math.max(
        1,
        ...report.results.map(result => result.latency_ms)
    );

    report.results.forEach(result => {
        const row = document.createElement("tr");

        const values = [
            result.name,
            result.status,
            result.http_status || "—",
            `${result.latency_ms} ms`
        ];

        values.forEach((value, index) => {
            const cell = document.createElement("td");

            cell.textContent = value;

            if (index === 1) {
                cell.classList.add(`status-${result.status}`);
            }

            if (index === 3) {
                const track = document.createElement("div");
                track.className = "latency-track";

                const bar = document.createElement("div");
                bar.className = "latency-bar";

                if (result.status !== "healthy") {
                    bar.classList.add("failed");
                }

                track.appendChild(bar);
                cell.appendChild(track);

                requestAnimationFrame(() => {
                    requestAnimationFrame(() => {
                        bar.style.width =
                            `${result.latency_ms / maxLatency * 100}%`;
                    });
                });
            }

            row.appendChild(cell);
        });

        healthResults.appendChild(row);
    });
}

// Clear previous performance measurements.
function resetPerformance() {
    executionTimes.sequential = null;
    executionTimes.concurrent = null;

    for (const mode of ["sequential", "concurrent"]) {
        document.getElementById(`${mode}-time`)
            .textContent = "—";

        document.getElementById(`${mode}-bar`)
            .style.width = "0%";
    }

    document.getElementById("speedup")
        .textContent = "—";

    document.getElementById("time-reduction")
        .textContent = "—";
}

// Display execution times and update comparison bars.
function renderExecutionTime(report) {
    executionTimes[report.mode] = report.total_ms;

    document.getElementById(`${report.mode}-time`)
        .textContent = `${report.total_ms.toLocaleString()} ms`;

    const maxTime = Math.max(
        executionTimes.sequential ?? 0,
        executionTimes.concurrent ?? 0,
        1
    );

    for (const mode of ["sequential", "concurrent"]) {
        const elapsed = executionTimes[mode];

        const percentage = elapsed === null
            ? 0
            : elapsed / maxTime * 100;

        document.getElementById(`${mode}-bar`)
            .style.width = `${percentage}%`;
    }

    renderSpeedup();
}

// Calculate the performance improvement.
function renderSpeedup() {
    const sequential = executionTimes.sequential;
    const concurrent = executionTimes.concurrent;

    if (
        sequential === null ||
        concurrent === null ||
        sequential <= 0 ||
        concurrent <= 0
    ) {
        return;
    }

    const speedup = sequential / concurrent;

    const reduction =
        (1 - concurrent / sequential) * 100;

    document.getElementById("speedup")
        .textContent = `${speedup.toFixed(2)}x`;

    document.getElementById("time-reduction")
        .textContent = `${reduction.toFixed(1)}%`;
}

// Execute one or both monitoring modes.
async function runMonitoring(modes) {
    setRunning(true);
    resetPerformance();

    try {
        for (const mode of modes) {
            healthStatus.textContent =
                `Running ${mode} health checks...`;

            const report = await fetchHealthReport(mode);

            renderSummary(report);
            renderResults(report);
            renderExecutionTime(report);

            console.log(`${mode} health report:`, report);
        }

        healthStatus.textContent =
            "Health checks complete!";

    } catch (error) {
        console.error(error);

        healthStatus.textContent =
            `Health check failed: ${error.message}`;

    } finally {
        setRunning(false);
    }
}

// Event listeners.
sequentialButton.addEventListener("click", () => {
    runMonitoring(["sequential"]);
});

concurrentButton.addEventListener("click", () => {
    runMonitoring(["concurrent"]);
});

compareButton.addEventListener("click", () => {
    runMonitoring(["sequential", "concurrent"]);
});

// Apply basic Go syntax highlighting.
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

        } else if (/^(int|int64|string|bool|error|float64|byte|rune)$/.test(token)) {
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

// Initialize expandable health source panels.
function initializeHealthSourcePanels() {
    const panels = document.querySelectorAll(".health-source");

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

        panel.addEventListener("toggle", async () => {
            if (!panel.open || loaded || loading) {
                return;
            }

            loading = true;
            code.textContent = "Loading Go source...";

            try {
                const response = await fetch(
                    `/api/health/source?component=${encodeURIComponent(component)}`
                );

                if (!response.ok) {
                    throw new Error(await response.text());
                }

                source = await response.text();

                highlightGo(source, code);

                loaded = true;
                copyButton.disabled = false;

            } catch (error) {
                code.textContent =
                    `Failed to load source: ${error.message}`;

            } finally {
                loading = false;
            }
        });

        copyButton.addEventListener("click", async () => {
            try {
                await navigator.clipboard.writeText(source);

                copyButton.textContent = "Copied!";

                setTimeout(() => {
                    copyButton.textContent = "Copy Source";
                }, 1500);

            } catch (error) {
                copyButton.textContent = "Copy failed";
                console.error(error);
            }
        });
    });
}

initializeHealthSourcePanels();