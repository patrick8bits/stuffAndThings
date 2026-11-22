const algorithms = ["bubble", "insertion", "quick"];

const arraySize = document.getElementById("array-size");
const animationSpeed = document.getElementById("animation-speed");

const sizeValue = document.getElementById("size-value");
const speedValue = document.getElementById("speed-value");

const startButton = document.getElementById("start-sort");
const randomizeButton = document.getElementById("randomize");
const resetButton = document.getElementById("reset");
const sortStatus = document.getElementById("sort-status");

let numbers = [];
let runId = 0;

// Generate randomized integers between 1 and 100.
function generateNumbers() {
    const size = Number(arraySize.value);

    numbers = Array.from({ length: size }, () =>
        Math.floor(Math.random() * 100) + 1
    );

    resetCharts();
}

// Render bars and highlight sorting operations.
function renderBars(algorithm, values, action = "", indices = []) {
    const chart = document.querySelector(
        `#${algorithm} .bar-chart`
    );

    chart.replaceChildren();

    values.forEach((value, index) => {
        const bar = document.createElement("div");

        bar.className = "bar";
        bar.style.height = `${value}%`;
        bar.title = value;

        if (action === "done") {
            bar.classList.add("sorted");
        } else if (indices.includes(index)) {
            if (action === "compare") {
                bar.classList.add("compare");
            } else if (["swap", "shift", "insert"].includes(action)) {
                bar.classList.add("swap");
            }
        }

        chart.appendChild(bar);
    });
}

// Update an algorithm's displayed statistics.
function updateStats(algorithm, stats, duration = "—") {
    const card = document.getElementById(algorithm);

    card.querySelector(".comparisons").textContent =
        stats.comparisons;

    card.querySelector(".swaps").textContent =
        stats.swaps;

    card.querySelector(".shifts").textContent =
        stats.shifts;

    card.querySelector(".duration").textContent =
        duration;
}

// Enable or disable controls during sorting.
function setRunning(running) {
    startButton.disabled = running;
    randomizeButton.disabled = running;
    arraySize.disabled = running;
}

// Restore the original dataset and cancel any animation.
function resetCharts() {
    runId++;

    algorithms.forEach(algorithm => {
        renderBars(algorithm, numbers);

        updateStats(algorithm, {
            comparisons: 0,
            swaps: 0,
            shifts: 0
        });
    });

    setRunning(false);
    sortStatus.textContent = "Ready to sort.";
}

// Request sorting traces from the Go backend.
async function fetchSortingTraces() {
    const response = await fetch("/api/sort", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            numbers: numbers
        })
    });

    if (!response.ok) {
        throw new Error(await response.text());
    }

    return await response.json();
}

// Pause execution between animation frames.
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

// Animate all three sorting algorithms.
async function startSorting() {
    resetCharts();

    const currentRun = runId;

    setRunning(true);
    sortStatus.textContent = "Requesting sorting traces from Go...";

    try {
        const traces = await fetchSortingTraces();

        if (currentRun !== runId) {
            return;
        }

        sortStatus.textContent = "Animating sorting algorithms...";

        const counters = {};

        algorithms.forEach(algorithm => {
            counters[algorithm] = {
                comparisons: 0,
                swaps: 0,
                shifts: 0
            };
        });

        const maxSteps = Math.max(
            ...algorithms.map(
                algorithm => traces[algorithm].steps.length
            )
        );

        for (let frame = 0; frame < maxSteps; frame++) {
            if (currentRun !== runId) {
                return;
            }

            algorithms.forEach(algorithm => {
                const trace = traces[algorithm];
                const stats = counters[algorithm];

                if (frame >= trace.steps.length) {
                    return;
                }

                const step = trace.steps[frame];

                if (step.action === "compare") {
                    stats.comparisons++;
                }

                if (step.action === "swap") {
                    stats.swaps++;
                }

                if (step.action === "shift") {
                    stats.shifts++;
                }

                renderBars(
                    algorithm,
                    step.values,
                    step.action,
                    step.indices
                );

                if (step.action === "done") {
                    const duration = trace.duration_ns === 0
                        ? "Below timer resolution"
                        : `${trace.duration_ns.toLocaleString()} ns`;

                    updateStats(algorithm, trace, duration);
                } else {
                    updateStats(algorithm, stats);
                }
            });

            if (frame < maxSteps - 1) {
                await sleep(Number(animationSpeed.value));
            }
        }

        sortStatus.textContent = "Sorting complete!";

    } catch (error) {
        if (currentRun === runId) {
            console.error(error);

            sortStatus.textContent =
                `Sorting failed: ${error.message}`;
        }
    } finally {
        if (currentRun === runId) {
            setRunning(false);
        }
    }
}

// Event listeners.
arraySize.addEventListener("input", () => {
    sizeValue.textContent = arraySize.value;
    generateNumbers();
});

animationSpeed.addEventListener("input", () => {
    speedValue.textContent = animationSpeed.value;
});

randomizeButton.addEventListener("click", generateNumbers);

resetButton.addEventListener("click", resetCharts);

startButton.addEventListener("click", startSorting);

// Initialize the page.
generateNumbers();

// Highlight Go source code.
function highlightGo(codeElement, source) {
    codeElement.replaceChildren();

    const pattern = /\/\/[^\n]*|\/\*[\s\S]*?\*\/|"(?:\\.|[^"\\])*"|`[^`]*`|'(?:\\.|[^'\\])*'|\b(?:package|import|func|return|if|else|for|range|var|const|type|struct|switch|case|default|break|continue)\b|\b(?:int|int64|string|bool|error)\b|\b(?:nil|true|false|\d+)\b/g;

    let previousIndex = 0;

    for (const match of source.matchAll(pattern)) {
        codeElement.append(
            document.createTextNode(
                source.slice(previousIndex, match.index)
            )
        );

        const token = match[0];
        const span = document.createElement("span");

        span.textContent = token;

        if (token.startsWith("//") || token.startsWith("/*")) {
            span.className = "go-comment";
        } else if (
            token.startsWith('"') ||
            token.startsWith("'") ||
            token.startsWith("`")
        ) {
            span.className = "go-string";
        } else if (/^(int|int64|string|bool|error)$/.test(token)) {
            span.className = "go-type";
        } else if (/^(nil|true|false|\d+)$/.test(token)) {
            span.className = "go-value";
        } else {
            span.className = "go-keyword";
        }

        codeElement.appendChild(span);

        previousIndex = match.index + token.length;
    }

    codeElement.append(
        document.createTextNode(source.slice(previousIndex))
    );
}

// Initialize expandable source code panels.
function initializeSourcePanels() {
    algorithms.forEach(algorithm => {
        const details = document.querySelector(
            `#${algorithm} details`
        );

        const code = details.querySelector("code");

        const copyButton = document.createElement("button");

        copyButton.textContent = "Copy Source";
        copyButton.className = "copy-source";
        copyButton.disabled = true;

        details.insertBefore(
            copyButton,
            details.querySelector("pre")
        );

        let source = null;
        let loading = false;

        details.addEventListener("toggle", async () => {
            if (!details.open || source !== null || loading) {
                return;
            }

            loading = true;
            code.textContent = "Loading Go source...";

            try {
                const response = await fetch(
                    `/api/sort/source?algorithm=${algorithm}`
                );

                if (!response.ok) {
                    throw new Error("Unable to retrieve source");
                }

                source = await response.text();

                highlightGo(code, source);

                copyButton.disabled = false;

            } catch (error) {
                code.textContent =
                    `${error.message}. Close and reopen to retry.`;
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
            }
        });
    });
}

initializeSourcePanels();