import './style.css';

// Import Go functions exposed via Wails bindings
import {
    CompareJSONWithOptions,
    FormatJSON,
    ValidateJSON,
    OpenJSONFile,
    GetJSONPaths,
    AnalyzeLogFile
} from '../wailsjs/go/main/App';

// ============================================================
// Tab Navigation
// ============================================================
const tabBtns = document.querySelectorAll('.tab-btn');
const tabContents = document.querySelectorAll('.tab-content');

tabBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        const tabId = btn.dataset.tab;

        // Update button states
        tabBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        // Update content visibility
        tabContents.forEach(content => {
            content.classList.remove('active');
            if (content.id === `tab-${tabId}`) {
                content.classList.add('active');
            }
        });
    });
});

// ============================================================
// Diff Tab - DOM Elements
// ============================================================
const leftTextarea = document.getElementById('left-json');
const rightTextarea = document.getElementById('right-json');
const leftError = document.getElementById('left-error');
const rightError = document.getElementById('right-error');
const compareBtn = document.getElementById('compare-btn');
const formatLeftBtn = document.getElementById('format-left');
const formatRightBtn = document.getElementById('format-right');
const loadLeftBtn = document.getElementById('load-left');
const loadRightBtn = document.getElementById('load-right');
const resultsDiv = document.getElementById('results');
const statsDiv = document.getElementById('stats');

// Normalization option checkboxes
const optSortKeys = document.getElementById('opt-sort-keys');
const optNormalizeNumbers = document.getElementById('opt-normalize-numbers');
const optTrimStrings = document.getElementById('opt-trim-strings');
const optNullEqualsAbsent = document.getElementById('opt-null-equals-absent');

// ============================================================
// Path Explorer Tab - DOM Elements
// ============================================================
const pathsTextarea = document.getElementById('paths-json');
const pathsError = document.getElementById('paths-error');
const extractBtn = document.getElementById('extract-btn');
const formatPathsBtn = document.getElementById('format-paths');
const loadPathsFileBtn = document.getElementById('load-paths-file');
const pathsResultsDiv = document.getElementById('paths-results');
const pathsStatsDiv = document.getElementById('paths-stats');

// ============================================================
// Diff Tab - Event Listeners
// ============================================================
compareBtn.addEventListener('click', handleCompare);
formatLeftBtn.addEventListener('click', () => handleFormat('left'));
formatRightBtn.addEventListener('click', () => handleFormat('right'));
loadLeftBtn.addEventListener('click', () => handleLoadFile('left'));
loadRightBtn.addEventListener('click', () => handleLoadFile('right'));

// Real-time validation as user types (debounced)
let leftTimeout, rightTimeout;
leftTextarea.addEventListener('input', () => {
    clearTimeout(leftTimeout);
    leftTimeout = setTimeout(() => validateInput('left'), 300);
});
rightTextarea.addEventListener('input', () => {
    clearTimeout(rightTimeout);
    rightTimeout = setTimeout(() => validateInput('right'), 300);
});

// ============================================================
// Path Explorer Tab - Event Listeners
// ============================================================
extractBtn.addEventListener('click', handleExtractPaths);
formatPathsBtn.addEventListener('click', () => handleFormatPaths());
loadPathsFileBtn.addEventListener('click', () => handleLoadPathsFile());

let pathsTimeout;
pathsTextarea.addEventListener('input', () => {
    clearTimeout(pathsTimeout);
    pathsTimeout = setTimeout(() => validatePathsInput(), 300);
});

// ============================================================
// Log Analyzer Tab - DOM Elements
// ============================================================
const analyzeFileBtn = document.getElementById('analyze-file-btn');
const logResultsDiv = document.getElementById('log-results');
const logStatsDiv = document.getElementById('log-stats');

// ============================================================
// Log Analyzer Tab - Event Listeners
// ============================================================
analyzeFileBtn.addEventListener('click', handleAnalyzeLogFile);

// ============================================================
// Shared Functions
// ============================================================

/**
 * Validate JSON input and show error if invalid
 */
async function validateInput(side) {
    const textarea = side === 'left' ? leftTextarea : rightTextarea;
    const errorDiv = side === 'left' ? leftError : rightError;

    const value = textarea.value.trim();
    if (!value) {
        errorDiv.textContent = '';
        return true;
    }

    try {
        const error = await ValidateJSON(value);
        if (error) {
            errorDiv.textContent = error;
            return false;
        }
        errorDiv.textContent = '';
        return true;
    } catch (err) {
        errorDiv.textContent = 'Validation error';
        return false;
    }
}

/**
 * Validate paths input
 */
async function validatePathsInput() {
    const value = pathsTextarea.value.trim();
    if (!value) {
        pathsError.textContent = '';
        return true;
    }

    try {
        const error = await ValidateJSON(value);
        if (error) {
            pathsError.textContent = error;
            return false;
        }
        pathsError.textContent = '';
        return true;
    } catch (err) {
        pathsError.textContent = 'Validation error';
        return false;
    }
}

/**
 * Format JSON in the specified textarea
 */
async function handleFormat(side) {
    const textarea = side === 'left' ? leftTextarea : rightTextarea;
    const errorDiv = side === 'left' ? leftError : rightError;

    const value = textarea.value.trim();
    if (!value) return;

    try {
        const formatted = await FormatJSON(value);
        textarea.value = formatted;
        errorDiv.textContent = '';
    } catch (err) {
        errorDiv.textContent = err.message || 'Invalid JSON';
    }
}

/**
 * Format JSON in paths textarea
 */
async function handleFormatPaths() {
    const value = pathsTextarea.value.trim();
    if (!value) return;

    try {
        const formatted = await FormatJSON(value);
        pathsTextarea.value = formatted;
        pathsError.textContent = '';
    } catch (err) {
        pathsError.textContent = err.message || 'Invalid JSON';
    }
}

/**
 * Load a JSON file into the specified textarea (diff tab)
 */
async function handleLoadFile(side) {
    const textarea = side === 'left' ? leftTextarea : rightTextarea;
    const errorDiv = side === 'left' ? leftError : rightError;

    try {
        const content = await OpenJSONFile();
        if (!content) return;

        textarea.value = content;
        errorDiv.textContent = '';
        validateInput(side);
    } catch (err) {
        errorDiv.textContent = err.message || 'Error loading file';
    }
}

/**
 * Load a JSON file into paths textarea
 */
async function handleLoadPathsFile() {
    try {
        const content = await OpenJSONFile();
        if (!content) return;

        pathsTextarea.value = content;
        pathsError.textContent = '';
        validatePathsInput();
    } catch (err) {
        pathsError.textContent = err.message || 'Error loading file';
    }
}

// ============================================================
// Diff Tab - Core Functions
// ============================================================

/**
 * Get current normalization options from UI checkboxes
 */
function getNormalizeOptions() {
    return {
        sortKeys: optSortKeys.checked,
        normalizeNumbers: optNormalizeNumbers.checked,
        trimStrings: optTrimStrings.checked,
        nullEqualsAbsent: optNullEqualsAbsent.checked,
        sortArrays: false,
        sortArraysByKey: '',
    };
}

/**
 * Compare the two JSON inputs and display results
 */
async function handleCompare() {
    const leftValue = leftTextarea.value.trim();
    const rightValue = rightTextarea.value.trim();

    resultsDiv.innerHTML = '';
    statsDiv.textContent = '';

    if (!leftValue || !rightValue) {
        resultsDiv.innerHTML = '<p class="error">Please enter JSON in both panels</p>';
        return;
    }

    try {
        const options = getNormalizeOptions();
        const result = await CompareJSONWithOptions(leftValue, rightValue, options);

        displayStats(result.stats);
        displayDiff(result.root);
    } catch (err) {
        resultsDiv.innerHTML = `<p class="error">${escapeHtml(err.message || 'Comparison failed')}</p>`;
    }
}

/**
 * Display diff statistics
 */
function displayStats(stats) {
    const parts = [];
    if (stats.added > 0) parts.push(`<span class="stat-added">+${stats.added} added</span>`);
    if (stats.removed > 0) parts.push(`<span class="stat-removed">-${stats.removed} removed</span>`);
    if (stats.changed > 0) parts.push(`<span class="stat-changed">~${stats.changed} changed</span>`);
    if (stats.equal > 0) parts.push(`<span class="stat-equal">${stats.equal} equal</span>`);

    if (parts.length === 0) {
        statsDiv.innerHTML = '<span class="stat-equal">No differences found</span>';
    } else {
        statsDiv.innerHTML = parts.join(' | ');
    }
}

/**
 * Display the diff tree recursively
 */
function displayDiff(node) {
    const container = document.createElement('div');
    container.className = 'diff-tree';

    renderNode(node, container, 0);

    resultsDiv.appendChild(container);
}

/**
 * Render a single diff node and its children
 */
function renderNode(node, container, depth) {
    if (node.path === '$' && node.children && node.children.length > 0) {
        for (const child of node.children) {
            renderNode(child, container, depth);
        }
        return;
    }

    const div = document.createElement('div');
    div.className = `diff-node diff-${node.type}`;
    div.style.paddingLeft = `${depth * 16}px`;

    let content = `<span class="diff-path">${escapeHtml(node.path)}</span>`;

    if (node.type === 'added') {
        content += ` <span class="diff-badge badge-added">added</span>`;
        content += ` <span class="diff-value">${formatValue(node.right)}</span>`;
    } else if (node.type === 'removed') {
        content += ` <span class="diff-badge badge-removed">removed</span>`;
        content += ` <span class="diff-value">${formatValue(node.left)}</span>`;
    } else if (node.type === 'changed') {
        content += ` <span class="diff-badge badge-changed">changed</span>`;
        if (node.left !== undefined && node.right !== undefined) {
            content += ` <span class="diff-value diff-old">${formatValue(node.left)}</span>`;
            content += ` → `;
            content += `<span class="diff-value diff-new">${formatValue(node.right)}</span>`;
        }
    }

    div.innerHTML = content;

    if (node.type !== 'equal' || (node.children && node.children.some(c => c.type !== 'equal'))) {
        container.appendChild(div);
    }

    if (node.children && node.children.length > 0) {
        for (const child of node.children) {
            renderNode(child, container, depth + 1);
        }
    }
}

/**
 * Format a value for display
 */
function formatValue(value) {
    if (value === null) return 'null';
    if (value === undefined) return 'undefined';
    if (typeof value === 'string') return `"${escapeHtml(value)}"`;
    if (typeof value === 'object') {
        const str = JSON.stringify(value);
        if (str.length > 100) {
            return escapeHtml(str.substring(0, 100) + '...');
        }
        return escapeHtml(str);
    }
    return String(value);
}

/**
 * Escape HTML entities to prevent XSS
 */
function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

// ============================================================
// Path Explorer Tab - Core Functions
// ============================================================

/**
 * Extract and display all JSON paths
 */
async function handleExtractPaths() {
    const value = pathsTextarea.value.trim();

    pathsResultsDiv.innerHTML = '';
    pathsStatsDiv.textContent = '';

    if (!value) {
        pathsResultsDiv.innerHTML = '<p class="error">Please enter JSON</p>';
        return;
    }

    try {
        const result = await GetJSONPaths(value);

        displayPathsStats(result);
        displayPaths(result.paths);
    } catch (err) {
        pathsResultsDiv.innerHTML = `<p class="error">${escapeHtml(err.message || 'Path extraction failed')}</p>`;
    }
}

/**
 * Display path extraction statistics
 */
function displayPathsStats(result) {
    pathsStatsDiv.innerHTML = `
        <span class="stat-equal">${result.totalPaths} unique paths</span> |
        <span class="stat-changed">${result.totalLeafs} total values</span>
    `;
}

/**
 * Display extracted paths as a table
 */
function displayPaths(paths) {
    if (paths.length === 0) {
        pathsResultsDiv.innerHTML = '<p class="placeholder">No paths found (empty JSON?)</p>';
        return;
    }

    const table = document.createElement('table');
    table.className = 'path-table';

    // Header
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th>Path</th>
            <th>Count</th>
        </tr>
    `;
    table.appendChild(thead);

    // Body
    const tbody = document.createElement('tbody');
    for (const item of paths) {
        const tr = document.createElement('tr');
        const countClass = item.count > 1 ? 'count-cell multiple' : 'count-cell';
        tr.innerHTML = `
            <td class="path-cell">${escapeHtml(item.path)}</td>
            <td class="${countClass}">${item.count}</td>
        `;
        tbody.appendChild(tr);
    }
    table.appendChild(tbody);

    pathsResultsDiv.appendChild(table);
}

// ============================================================
// Log Analyzer Tab - Core Functions
// ============================================================

/**
 * Analyze a log file for JSON paths
 */
async function handleAnalyzeLogFile() {
    logResultsDiv.innerHTML = '<p class="placeholder">Analyzing file...</p>';
    logStatsDiv.textContent = '';

    try {
        const result = await AnalyzeLogFile();

        // User cancelled file dialog
        if (!result) {
            logResultsDiv.innerHTML = '<p class="placeholder">Click "Select Log File" to analyze JSON paths in a log file</p>';
            return;
        }

        displayLogStats(result);
        displayLogPaths(result.paths);
    } catch (err) {
        logResultsDiv.innerHTML = `<p class="error">${escapeHtml(err.message || 'Analysis failed')}</p>`;
    }
}

/**
 * Display log analysis statistics
 */
function displayLogStats(result) {
    logStatsDiv.innerHTML = `
        <span class="stat-equal">${result.jsonLines.toLocaleString()} JSON lines</span> |
        <span class="stat-removed">${result.skippedLines.toLocaleString()} skipped</span> |
        <span class="stat-changed">${result.totalPaths.toLocaleString()} unique paths</span>
    `;
}

/**
 * Display log analysis paths as a table
 */
function displayLogPaths(paths) {
    if (paths.length === 0) {
        logResultsDiv.innerHTML = '<p class="placeholder">No JSON paths found in file</p>';
        return;
    }

    const table = document.createElement('table');
    table.className = 'path-table';

    // Header - includes objectHits and distinctCount columns
    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th>Path</th>
            <th>Count</th>
            <th>Objects</th>
            <th>Distinct</th>
        </tr>
    `;
    table.appendChild(thead);

    // Body
    const tbody = document.createElement('tbody');
    for (const item of paths) {
        const tr = document.createElement('tr');
        const countClass = item.count > 1 ? 'count-cell multiple' : 'count-cell';
        // Highlight if all values are distinct (likely an ID field)
        // Only consider it meaningfully unique if we have enough samples (n >= 10)
        const isUnique = item.distinctCount === item.count;
        const isStatisticallyMeaningful = item.count >= 10;
        const distinctClass = isUnique ? 'count-cell unique clickable' : 'count-cell clickable';
        const distinctTitle = isUnique
            ? 'Every value is unique (likely an ID/key field). Click to see values.'
            : 'Click to see top values';
        // Only show key icon if uniqueness is statistically meaningful
        const keyIcon = (isUnique && isStatisticallyMeaningful) ? '<span class="key-icon" title="All values unique">🔑</span>' : '';
        tr.innerHTML = `
            <td class="path-cell">${escapeHtml(item.path)}</td>
            <td class="${countClass}">${item.count.toLocaleString()}</td>
            <td class="objects-cell">${item.objectHits.toLocaleString()}</td>
            <td class="${distinctClass}" title="${distinctTitle}">${item.distinctCount.toLocaleString()}${keyIcon}</td>
        `;

        // Add click handler for distinct count
        const distinctCell = tr.querySelector('.clickable');
        distinctCell.addEventListener('click', (e) => {
            e.stopPropagation();
            showTopValues(item, tr);
        });

        tbody.appendChild(tr);
    }
    table.appendChild(tbody);

    logResultsDiv.appendChild(table);
}

/**
 * Show top values for a path in an expandable row
 */
function showTopValues(item, row) {
    // Check if already expanded
    const existingDetail = row.nextElementSibling;
    if (existingDetail && existingDetail.classList.contains('value-detail-row')) {
        existingDetail.remove();
        return;
    }

    // Remove any other open detail rows
    document.querySelectorAll('.value-detail-row').forEach(r => r.remove());

    if (!item.topValues || item.topValues.length === 0) {
        return;
    }

    // Create detail row
    const detailRow = document.createElement('tr');
    detailRow.className = 'value-detail-row';

    const detailCell = document.createElement('td');
    detailCell.colSpan = 4;

    // Build value list
    let html = '<div class="value-detail"><strong>Top values:</strong><ul>';
    for (const v of item.topValues) {
        const displayValue = v.value.length > 50 ? v.value.substring(0, 50) + '...' : v.value;
        html += `<li><span class="value-text">${escapeHtml(displayValue)}</span> <span class="value-count">(${v.count.toLocaleString()})</span></li>`;
    }
    if (item.distinctCount > item.topValues.length) {
        html += `<li class="more-values">... and ${(item.distinctCount - item.topValues.length).toLocaleString()} more</li>`;
    }
    html += '</ul></div>';

    detailCell.innerHTML = html;
    detailRow.appendChild(detailCell);

    // Insert after current row
    row.after(detailRow);
}
