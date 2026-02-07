import './style.css';

// Import Go functions exposed via Wails bindings
import {
    CompareJSONWithOptions,
    FormatJSON,
    ValidateJSON,
    OpenJSONFile,
    OpenJSONFileWithPath,
    ReadFilePath,
    GetJSONPaths,
    GetJSONPathsWithContainers,
    SelectAndAnalyzeLogFile,
    AnalyzeLogFilePath,
    CompareLogAnalyses,
    CompareLogFiles,
    GetAllFileHistory,
    SaveFilePathToHistory
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
// File Path History Management
// ============================================================

// Map of history keys to datalist IDs
const historyConfig = {
    'diff-left': 'left-file-history',
    'diff-right': 'right-file-history',
    'paths': 'paths-file-history',
    'logs': 'logs-file-history',
    'compare-left': 'compare-left-history',
    'compare-right': 'compare-right-history'
};

/**
 * Update a datalist with history items
 */
function updateDatalist(datalistId, paths) {
    const datalist = document.getElementById(datalistId);
    if (!datalist) return;

    // Clear existing options
    datalist.innerHTML = '';

    // Add new options from history (newest first)
    paths.forEach(path => {
        const option = document.createElement('option');
        option.value = path;
        datalist.appendChild(option);
    });
}

/**
 * Save a file path to history and update the corresponding datalist
 */
async function saveToHistory(key, path) {
    if (!path) return;

    try {
        await SaveFilePathToHistory(key, path);

        // Update the datalist
        const datalistId = historyConfig[key];
        if (datalistId) {
            // Get the updated history and refresh the datalist
            const history = await GetAllFileHistory();
            if (history[key]) {
                updateDatalist(datalistId, history[key]);
            }
        }
    } catch (err) {
        console.error('Error saving to history:', err);
    }
}

/**
 * Load file path history on startup and populate all datalists
 */
async function loadFilePathHistory() {
    try {
        const history = await GetAllFileHistory();

        // Update each datalist
        for (const [key, datalistId] of Object.entries(historyConfig)) {
            const paths = history[key] || [];
            updateDatalist(datalistId, paths);
        }

        // Also populate the most recent path in each input (optional - can be removed if not desired)
        // This auto-fills the last used path on startup
        const leftPath = history['diff-left']?.[0];
        const rightPath = history['diff-right']?.[0];
        const pathsPath = history['paths']?.[0];
        const logsPath = history['logs']?.[0];
        const cmpLeftPath = history['compare-left']?.[0];
        const cmpRightPath = history['compare-right']?.[0];

        if (leftPath) leftFilePathInput.value = leftPath;
        if (rightPath) rightFilePathInput.value = rightPath;
        if (pathsPath) pathsFilePathInput.value = pathsPath;
        if (logsPath) logFilePathInput.value = logsPath;
        if (cmpLeftPath) compareLeftPath.value = cmpLeftPath;
        if (cmpRightPath) compareRightPath.value = cmpRightPath;
    } catch (err) {
        console.error('Error loading file path history:', err);
    }
}

// ============================================================
// Diff Tab - DOM Elements
// ============================================================
const leftTextarea = document.getElementById('left-json');
const rightTextarea = document.getElementById('right-json');
const leftFilePathInput = document.getElementById('left-file-path');
const rightFilePathInput = document.getElementById('right-file-path');
const leftError = document.getElementById('left-error');
const rightError = document.getElementById('right-error');
const compareBtn = document.getElementById('compare-btn');
const formatLeftBtn = document.getElementById('format-left');
const formatRightBtn = document.getElementById('format-right');
const loadLeftBtn = document.getElementById('load-left');
const loadRightBtn = document.getElementById('load-right');
const reloadLeftBtn = document.getElementById('reload-left');
const reloadRightBtn = document.getElementById('reload-right');
const resultsDiv = document.getElementById('results');
const statsDiv = document.getElementById('stats');

// Normalization option checkboxes
const optSortKeys = document.getElementById('opt-sort-keys');
const optNormalizeNumbers = document.getElementById('opt-normalize-numbers');
const optTrimStrings = document.getElementById('opt-trim-strings');
const optNullEqualsAbsent = document.getElementById('opt-null-equals-absent');

// View mode toggle
const viewModeBtns = document.querySelectorAll('#diff-view-toggle .mode-btn');
let currentViewMode = 'structured';
let lastDiffResult = null; // Store the last diff result for view switching

// ============================================================
// Path Explorer Tab - DOM Elements
// ============================================================
const pathsTextarea = document.getElementById('paths-json');
const pathsFilePathInput = document.getElementById('paths-file-path');
const pathsError = document.getElementById('paths-error');
const extractBtn = document.getElementById('extract-btn');
const formatPathsBtn = document.getElementById('format-paths');
const loadPathsFileBtn = document.getElementById('load-paths-file');
const pathsResultsDiv = document.getElementById('paths-results');
const pathsStatsDiv = document.getElementById('paths-stats');
const optIncludeContainers = document.getElementById('opt-include-containers');

// ============================================================
// Diff Tab - Event Listeners
// ============================================================
compareBtn.addEventListener('click', handleCompare);
formatLeftBtn.addEventListener('click', () => handleFormat('left'));
formatRightBtn.addEventListener('click', () => handleFormat('right'));
loadLeftBtn.addEventListener('click', () => handleLoadFile('left'));
loadRightBtn.addEventListener('click', () => handleLoadFile('right'));
reloadLeftBtn.addEventListener('click', () => handleReloadFile('left'));
reloadRightBtn.addEventListener('click', () => handleReloadFile('right'));

// File path inputs - load on Enter
leftFilePathInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
        handleLoadFromPath('left');
    }
});
rightFilePathInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
        handleLoadFromPath('right');
    }
});

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

// View mode toggle
viewModeBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        const viewMode = btn.dataset.view;

        // Update button states
        viewModeBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        // Update current view mode
        currentViewMode = viewMode;

        // Re-render results if we have a diff
        if (lastDiffResult) {
            displayDiffInCurrentMode(lastDiffResult);
        }
    });
});

// ============================================================
// Path Explorer Tab - Event Listeners
// ============================================================
extractBtn.addEventListener('click', handleExtractPaths);
formatPathsBtn.addEventListener('click', () => handleFormatPaths());
loadPathsFileBtn.addEventListener('click', () => handleLoadPathsFile());

// File path input - load on Enter
pathsFilePathInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
        handleLoadPathsFromPath();
    }
});

let pathsTimeout;
pathsTextarea.addEventListener('input', () => {
    clearTimeout(pathsTimeout);
    pathsTimeout = setTimeout(() => validatePathsInput(), 300);
});

// ============================================================
// Log Analyzer Tab - DOM Elements
// ============================================================
const analyzeFileBtn = document.getElementById('analyze-file-btn');
const logFilePathInput = document.getElementById('log-file-path');
const logResultsDiv = document.getElementById('log-results');
const logStatsDiv = document.getElementById('log-stats');

// ============================================================
// Log Analyzer Tab - Event Listeners
// ============================================================
analyzeFileBtn.addEventListener('click', handleAnalyzeLogFile);
logFilePathInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
        handleAnalyzeFromPath();
    }
});

// ============================================================
// Shared Functions
// ============================================================

/**
 * Check if both diff panels have valid JSON and auto-compare if so
 */
async function tryAutoCompare() {
    const leftValue = leftTextarea.value.trim();
    const rightValue = rightTextarea.value.trim();

    // Both panels must have content
    if (!leftValue || !rightValue) return;

    // Both must be valid JSON
    const leftValid = await validateInput('left');
    const rightValid = await validateInput('right');

    if (leftValid && rightValid) {
        // Both sides are valid - trigger comparison
        await handleCompare();
    }
}

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
 * Load a JSON file into the specified textarea (diff tab).
 * If a path is provided in the input, try to load that file.
 * If no path or invalid path, open the file picker.
 */
async function handleLoadFile(side) {
    const textarea = side === 'left' ? leftTextarea : rightTextarea;
    const pathInput = side === 'left' ? leftFilePathInput : rightFilePathInput;
    const errorDiv = side === 'left' ? leftError : rightError;
    const historyKey = side === 'left' ? 'diff-left' : 'diff-right';

    const existingPath = pathInput.value.trim();

    // If there's a path, try to load it first
    if (existingPath) {
        try {
            const content = await ReadFilePath(existingPath);
            textarea.value = content;
            errorDiv.textContent = '';
            validateInput(side);

            // Save to history
            await saveToHistory(historyKey, existingPath);

            // Auto-compare if both sides have valid JSON
            await tryAutoCompare();
            return;
        } catch {
            // Path is invalid, fall through to file picker
        }
    }

    // No path or invalid path - open file picker
    try {
        const result = await OpenJSONFileWithPath();
        if (!result) return;

        textarea.value = result.content;
        pathInput.value = result.path;
        errorDiv.textContent = '';
        validateInput(side);

        // Save to history
        await saveToHistory(historyKey, result.path);

        // Auto-compare if both sides have valid JSON
        await tryAutoCompare();
    } catch (err) {
        errorDiv.textContent = err.message || 'Error loading file';
    }
}

/**
 * Load a JSON file from a pasted path
 */
async function handleLoadFromPath(side) {
    const textarea = side === 'left' ? leftTextarea : rightTextarea;
    const pathInput = side === 'left' ? leftFilePathInput : rightFilePathInput;
    const errorDiv = side === 'left' ? leftError : rightError;
    const historyKey = side === 'left' ? 'diff-left' : 'diff-right';

    const path = pathInput.value.trim();
    if (!path) {
        errorDiv.textContent = 'Please enter a file path';
        return;
    }

    try {
        const content = await ReadFilePath(path);
        textarea.value = content;
        errorDiv.textContent = '';
        validateInput(side);

        // Save to history
        await saveToHistory(historyKey, path);

        // Auto-compare if both sides have valid JSON
        await tryAutoCompare();
    } catch (err) {
        errorDiv.textContent = err.message || 'Error loading file';
    }
}

/**
 * Reload the current file from disk
 */
async function handleReloadFile(side) {
    const textarea = side === 'left' ? leftTextarea : rightTextarea;
    const pathInput = side === 'left' ? leftFilePathInput : rightFilePathInput;
    const errorDiv = side === 'left' ? leftError : rightError;

    const path = pathInput.value.trim();
    if (!path) {
        errorDiv.textContent = 'No file loaded to reload';
        return;
    }

    try {
        const content = await ReadFilePath(path);
        textarea.value = content;
        errorDiv.textContent = '';
        validateInput(side);

        // Show success feedback
        const fileName = path.split('/').pop();
        showCopyFeedback(`✓ Reloaded ${fileName}`);

        // Auto-compare after reload
        await tryAutoCompare();
    } catch (err) {
        errorDiv.textContent = err.message || 'Error reloading file';
    }
}

/**
 * Load a JSON file into paths textarea.
 * If a path is provided in the input, try to load that file.
 * If no path or invalid path, open the file picker.
 */
async function handleLoadPathsFile() {
    const existingPath = pathsFilePathInput.value.trim();

    // If there's a path, try to load it first
    if (existingPath) {
        try {
            const content = await ReadFilePath(existingPath);
            pathsTextarea.value = content;
            pathsError.textContent = '';
            validatePathsInput();

            // Save to history
            await saveToHistory('paths', existingPath);
            return;
        } catch {
            // Path is invalid, fall through to file picker
        }
    }

    // No path or invalid path - open file picker
    try {
        const result = await OpenJSONFileWithPath();
        if (!result) return;

        pathsTextarea.value = result.content;
        pathsFilePathInput.value = result.path;
        pathsError.textContent = '';
        validatePathsInput();

        // Save to history
        await saveToHistory('paths', result.path);
    } catch (err) {
        pathsError.textContent = err.message || 'Error loading file';
    }
}

/**
 * Load a JSON file from a pasted path into paths textarea
 */
async function handleLoadPathsFromPath() {
    const path = pathsFilePathInput.value.trim();
    if (!path) {
        pathsError.textContent = 'Please enter a file path';
        return;
    }

    try {
        const content = await ReadFilePath(path);
        pathsTextarea.value = content;
        pathsError.textContent = '';
        validatePathsInput();

        // Save to history
        await saveToHistory('paths', path);
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

        // Store the result and raw values for view switching
        lastDiffResult = {
            result: result,
            leftValue: leftValue,
            rightValue: rightValue
        };

        displayStats(result.stats);
        displayDiffInCurrentMode(lastDiffResult);
    } catch (err) {
        resultsDiv.innerHTML = `<p class="error">${escapeHtml(err.message || 'Comparison failed')}</p>`;
    }
}

/**
 * Display diff in the current view mode
 */
function displayDiffInCurrentMode(diffData) {
    resultsDiv.innerHTML = '';

    if (currentViewMode === 'structured') {
        displayDiff(diffData.result.root);
    } else {
        displaySideBySideDiff(diffData.leftValue, diffData.rightValue);
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
    if (node.path === '' && node.children && node.children.length > 0) {
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

/**
 * Display side-by-side diff view
 */
/**
 * Recursively sort object keys for consistent display
 */
function sortObjectKeys(obj) {
    if (obj === null || typeof obj !== 'object') {
        return obj;
    }

    if (Array.isArray(obj)) {
        return obj.map(sortObjectKeys);
    }

    // Sort keys alphabetically and recursively process values
    const sorted = {};
    const keys = Object.keys(obj).sort();
    for (const key of keys) {
        sorted[key] = sortObjectKeys(obj[key]);
    }
    return sorted;
}

function displaySideBySideDiff(leftValue, rightValue) {
    // Get current normalization options
    const shouldSortKeys = optSortKeys.checked;

    // Parse and format both JSON values
    let leftLines, rightLines;
    try {
        let leftObj = JSON.parse(leftValue);
        if (shouldSortKeys) {
            leftObj = sortObjectKeys(leftObj);
        }
        leftLines = JSON.stringify(leftObj, null, 2).split('\n');
    } catch {
        leftLines = leftValue.split('\n');
    }

    try {
        let rightObj = JSON.parse(rightValue);
        if (shouldSortKeys) {
            rightObj = sortObjectKeys(rightObj);
        }
        rightLines = JSON.stringify(rightObj, null, 2).split('\n');
    } catch {
        rightLines = rightValue.split('\n');
    }

    // Create container
    const container = document.createElement('div');
    container.className = 'sidebyside-container';

    // Create left panel
    const leftPanel = document.createElement('div');
    leftPanel.className = 'sidebyside-panel';
    leftPanel.innerHTML = `
        <div class="sidebyside-header">Left (Original)</div>
        <div class="sidebyside-content" id="sidebyside-left"></div>
    `;

    // Create right panel
    const rightPanel = document.createElement('div');
    rightPanel.className = 'sidebyside-panel';
    rightPanel.innerHTML = `
        <div class="sidebyside-header">Right (Modified)</div>
        <div class="sidebyside-content" id="sidebyside-right"></div>
    `;

    container.appendChild(leftPanel);
    container.appendChild(rightPanel);
    resultsDiv.appendChild(container);

    // Get content divs
    const leftContent = document.getElementById('sidebyside-left');
    const rightContent = document.getElementById('sidebyside-right');

    // Simple line-by-line comparison
    const maxLines = Math.max(leftLines.length, rightLines.length);

    for (let i = 0; i < maxLines; i++) {
        const leftLine = i < leftLines.length ? leftLines[i] : null;
        const rightLine = i < rightLines.length ? rightLines[i] : null;

        // Determine line status
        let leftClass = '';
        let rightClass = '';

        if (leftLine !== null && rightLine !== null) {
            if (leftLine !== rightLine) {
                leftClass = 'line-changed';
                rightClass = 'line-changed';
            }
        } else if (leftLine !== null && rightLine === null) {
            leftClass = 'line-removed';
            rightClass = 'line-empty';
        } else if (leftLine === null && rightLine !== null) {
            leftClass = 'line-empty';
            rightClass = 'line-added';
        }

        // Create left line
        const leftLineDiv = document.createElement('div');
        leftLineDiv.className = `sidebyside-line ${leftClass}`;
        leftLineDiv.innerHTML = `
            <div class="line-number">${leftLine !== null ? i + 1 : ''}</div>
            <div class="line-content">${leftLine !== null ? escapeHtml(leftLine) : ''}</div>
        `;
        leftContent.appendChild(leftLineDiv);

        // Create right line
        const rightLineDiv = document.createElement('div');
        rightLineDiv.className = `sidebyside-line ${rightClass}`;
        rightLineDiv.innerHTML = `
            <div class="line-number">${rightLine !== null ? i + 1 : ''}</div>
            <div class="line-content">${rightLine !== null ? escapeHtml(rightLine) : ''}</div>
        `;
        rightContent.appendChild(rightLineDiv);
    }

    // Sync scrolling between both panels
    leftContent.addEventListener('scroll', () => {
        rightContent.scrollTop = leftContent.scrollTop;
    });
    rightContent.addEventListener('scroll', () => {
        leftContent.scrollTop = rightContent.scrollTop;
    });
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
        const includeContainers = optIncludeContainers.checked;
        const result = await GetJSONPathsWithContainers(value, includeContainers);

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
 * Analyze a log file for JSON paths.
 * If a path is provided in the input, try to analyze that file.
 * If no path or invalid path, open the file picker.
 */
async function handleAnalyzeLogFile() {
    const existingPath = logFilePathInput.value.trim();

    // If there's a path, try to analyze it first
    if (existingPath) {
        logResultsDiv.innerHTML = '<p class="placeholder">Analyzing file...</p>';
        logStatsDiv.textContent = '';

        try {
            const result = await AnalyzeLogFilePath(existingPath);

            displayLogStats(result);
            displayLogPaths(result.paths);

            // Save to history
            await saveToHistory('logs', existingPath);
            return;
        } catch {
            // Path is invalid, fall through to file picker
        }
    }

    // No path or invalid path - open file picker
    logResultsDiv.innerHTML = '<p class="placeholder">Analyzing file...</p>';
    logStatsDiv.textContent = '';

    try {
        const response = await SelectAndAnalyzeLogFile();

        // User cancelled file dialog
        if (!response) {
            logResultsDiv.innerHTML = '<p class="placeholder">Click "Load File" to analyze JSON paths in a log file</p>';
            return;
        }

        // Show the selected file path
        logFilePathInput.value = response.path;

        displayLogStats(response.result);
        displayLogPaths(response.result.paths);

        // Save to history
        await saveToHistory('logs', response.path);
    } catch (err) {
        logResultsDiv.innerHTML = `<p class="error">${escapeHtml(err.message || 'Analysis failed')}</p>`;
    }
}

/**
 * Analyze a log file from a pasted path
 */
async function handleAnalyzeFromPath() {
    const path = logFilePathInput.value.trim();

    if (!path) {
        logResultsDiv.innerHTML = '<p class="error">Please enter a file path</p>';
        return;
    }

    logResultsDiv.innerHTML = '<p class="placeholder">Analyzing file...</p>';
    logStatsDiv.textContent = '';

    try {
        const result = await AnalyzeLogFilePath(path);

        displayLogStats(result);
        displayLogPaths(result.paths);

        // Save to history
        await saveToHistory('logs', path);
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

    // Clear any existing content (like "Analyzing file...")
    logResultsDiv.innerHTML = '';

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
            <td class="path-cell clickable-path" title="Click to copy jq command">${escapeHtml(item.path)}</td>
            <td class="${countClass}">${item.count.toLocaleString()}</td>
            <td class="objects-cell">${item.objectHits.toLocaleString()}</td>
            <td class="${distinctClass}" title="${distinctTitle}">${item.distinctCount.toLocaleString()}${keyIcon}</td>
        `;

        // Add click handler for path cell - copy to clipboard
        const pathCell = tr.querySelector('.clickable-path');
        pathCell.addEventListener('click', (e) => {
            e.stopPropagation();
            copyJqCommand(item.path);
        });

        // Add click handler for distinct count
        const distinctCell = tr.querySelector('.count-cell.clickable');
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

/**
 * Copy a jq command to the clipboard.
 * If a file path is available, copies a full bash command.
 * Otherwise, just copies the jq path pattern.
 */
async function copyJqCommand(path) {
    const filePath = logFilePathInput.value.trim();
    let textToCopy;

    if (filePath) {
        // Full jq command with file path
        textToCopy = `cat '${filePath}' | grep -E '^\\{' | jq '${path}'`;
    } else {
        // Just the jq path
        textToCopy = path;
    }

    try {
        await navigator.clipboard.writeText(textToCopy);
        showCopyFeedback('Copied!');
    } catch (err) {
        console.error('Failed to copy:', err);
    }
}

/**
 * Show brief feedback when something is copied
 */
function showCopyFeedback(message) {
    // Create or reuse feedback element
    let feedback = document.getElementById('copy-feedback');
    if (!feedback) {
        feedback = document.createElement('div');
        feedback.id = 'copy-feedback';
        document.body.appendChild(feedback);
    }

    feedback.textContent = message;
    feedback.classList.add('visible');

    setTimeout(() => {
        feedback.classList.remove('visible');
    }, 1500);
}

// ============================================================
// Log Analyzer Compare Mode - DOM Elements
// ============================================================
const logModeBtns = document.querySelectorAll('.log-mode-toggle .mode-btn');
const singleMode = document.getElementById('log-single-mode');
const compareMode = document.getElementById('log-compare-mode');
const compareLeftLoadBtn = document.getElementById('compare-left-load');
const compareRightLoadBtn = document.getElementById('compare-right-load');
const compareLeftPath = document.getElementById('compare-left-path');
const compareRightPath = document.getElementById('compare-right-path');
const compareLeftInfo = document.getElementById('compare-left-info');
const compareRightInfo = document.getElementById('compare-right-info');
const compareAnalysesBtn = document.getElementById('compare-analyses-btn');
const comparisonResultsContainer = document.getElementById('comparison-results-container');
const comparisonStats = document.getElementById('comparison-stats');
const comparisonTbody = document.getElementById('comparison-tbody');
const optShowOnlyChanges = document.getElementById('opt-show-only-changes');
const optShowTopValues = document.getElementById('opt-show-top-values');

// View mode toggle elements
const compareViewToggleBtns = document.querySelectorAll('#compare-view-toggle .mode-btn');
const comparisonStructuredView = document.getElementById('comparison-structured-view');
const comparisonSideBySideView = document.getElementById('comparison-sidebyside-view');
const sidebysideLeftTbody = document.getElementById('sidebyside-left-tbody');
const sidebysideRightTbody = document.getElementById('sidebyside-right-tbody');

// Store analysis results for comparison
let leftAnalysisResult = null;
let rightAnalysisResult = null;
let currentCompareViewMode = 'structured';

// ============================================================
// Log Analyzer Compare Mode - Event Listeners
// ============================================================

// Log analyzer mode toggle (Single Analysis / Compare Files)
logModeBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        const mode = btn.dataset.mode;

        // Update button states (only within this toggle group)
        logModeBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        // Show appropriate mode
        if (mode === 'single') {
            singleMode.style.display = 'flex';
            compareMode.style.display = 'none';
        } else {
            singleMode.style.display = 'none';
            compareMode.style.display = 'flex';
        }
    });
});

// File loading
compareLeftLoadBtn.addEventListener('click', () => handleCompareLoadFile('left'));
compareRightLoadBtn.addEventListener('click', () => handleCompareLoadFile('right'));

// File path inputs - load on Enter
compareLeftPath.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
        handleCompareLoadFromPath('left');
    }
});
compareRightPath.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
        handleCompareLoadFromPath('right');
    }
});

// Compare button
compareAnalysesBtn.addEventListener('click', handleCompareAnalyses);

// Filter options - re-render when changed
optShowOnlyChanges.addEventListener('change', renderComparison);
optShowTopValues.addEventListener('change', renderComparison);

// View mode toggle for comparison
compareViewToggleBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        const viewMode = btn.dataset.view;

        // Update button states
        compareViewToggleBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        // Update current view mode
        currentCompareViewMode = viewMode;

        // Toggle view containers
        if (viewMode === 'structured') {
            comparisonStructuredView.style.display = 'block';
            comparisonSideBySideView.style.display = 'none';
        } else {
            comparisonStructuredView.style.display = 'none';
            comparisonSideBySideView.style.display = 'block';
        }

        // Re-render if we have comparison data
        if (currentComparison) {
            renderComparison();
        }
    });
});

// ============================================================
// Log Analyzer Compare Mode - Core Functions
// ============================================================

/**
 * Load and analyze a file for comparison (left or right).
 * If a path is provided in the input, try to analyze that file.
 * If no path or invalid path, open the file picker.
 */
async function handleCompareLoadFile(side) {
    const pathInput = side === 'left' ? compareLeftPath : compareRightPath;
    const infoDiv = side === 'left' ? compareLeftInfo : compareRightInfo;
    const historyKey = side === 'left' ? 'compare-left' : 'compare-right';

    const existingPath = pathInput.value.trim();

    // If there's a path, try to analyze it first
    if (existingPath) {
        infoDiv.innerHTML = 'Analyzing file...';

        try {
            const result = await AnalyzeLogFilePath(existingPath);

            // Store the result
            if (side === 'left') {
                leftAnalysisResult = result;
            } else {
                rightAnalysisResult = result;
            }

            // Update UI
            displayCompareFileInfo(result, infoDiv);

            // Enable compare button if both files are loaded
            if (leftAnalysisResult && rightAnalysisResult) {
                compareAnalysesBtn.disabled = false;
            }

            // Save to history
            await saveToHistory(historyKey, existingPath);
            return;
        } catch {
            // Path is invalid, fall through to file picker
        }
    }

    // No path or invalid path - open file picker
    infoDiv.innerHTML = 'Loading file...';

    try {
        const response = await SelectAndAnalyzeLogFile();

        // User cancelled
        if (!response) {
            infoDiv.innerHTML = '';
            return;
        }

        // Store the result
        if (side === 'left') {
            leftAnalysisResult = response.result;
        } else {
            rightAnalysisResult = response.result;
        }

        // Update UI
        pathInput.value = response.path;
        displayCompareFileInfo(response.result, infoDiv);

        // Enable compare button if both files are loaded
        if (leftAnalysisResult && rightAnalysisResult) {
            compareAnalysesBtn.disabled = false;
        }

        // Save to history
        await saveToHistory(historyKey, response.path);
    } catch (err) {
        infoDiv.innerHTML = `<span style="color: var(--error-color)">Error: ${escapeHtml(err.message)}</span>`;
    }
}

/**
 * Load and analyze a file from a pasted path for comparison
 */
async function handleCompareLoadFromPath(side) {
    const pathInput = side === 'left' ? compareLeftPath : compareRightPath;
    const infoDiv = side === 'left' ? compareLeftInfo : compareRightInfo;
    const historyKey = side === 'left' ? 'compare-left' : 'compare-right';

    const path = pathInput.value.trim();
    if (!path) {
        infoDiv.innerHTML = '<span style="color: var(--error-color)">Please enter a file path</span>';
        return;
    }

    infoDiv.innerHTML = 'Analyzing file...';

    try {
        const result = await AnalyzeLogFilePath(path);

        // Store the result
        if (side === 'left') {
            leftAnalysisResult = result;
        } else {
            rightAnalysisResult = result;
        }

        // Update UI
        displayCompareFileInfo(result, infoDiv);

        // Enable compare button if both files are loaded
        if (leftAnalysisResult && rightAnalysisResult) {
            compareAnalysesBtn.disabled = false;
        }

        // Save to history
        await saveToHistory(historyKey, path);
    } catch (err) {
        infoDiv.innerHTML = `<span style="color: var(--error-color)">Error: ${escapeHtml(err.message)}</span>`;
    }
}

/**
 * Display file info summary in the compare panel
 */
function displayCompareFileInfo(result, infoDiv) {
    infoDiv.innerHTML = `
        <span class="success-icon">✓</span>
        <span class="stat">${result.jsonLines.toLocaleString()} lines</span>
        <span>|</span>
        <span class="stat">${result.totalPaths.toLocaleString()} paths</span>
    `;
}

// Store current comparison result
let currentComparison = null;

/**
 * Compare the two analysis results
 */
async function handleCompareAnalyses() {
    if (!leftAnalysisResult || !rightAnalysisResult) {
        return;
    }

    try {
        // Await the async Wails binding call
        const comparison = await CompareLogAnalyses(
            leftAnalysisResult,
            rightAnalysisResult,
            compareLeftPath.value,
            compareRightPath.value
        );

        currentComparison = comparison;

        // Show results container
        comparisonResultsContainer.style.display = 'flex';

        // Display stats and table
        displayComparisonStats(comparison.stats);
        renderComparison();
    } catch (err) {
        comparisonResultsContainer.style.display = 'flex';
        comparisonTbody.innerHTML = `<tr><td colspan="8" style="text-align: center; color: var(--error-color)">Error: ${escapeHtml(err.message || 'Comparison failed')}</td></tr>`;
    }
}

/**
 * Display comparison statistics
 */
function displayComparisonStats(stats) {
    const parts = [];
    if (stats.addedPaths > 0) parts.push(`<span class="stat-added">${stats.addedPaths} added</span>`);
    if (stats.removedPaths > 0) parts.push(`<span class="stat-removed">${stats.removedPaths} removed</span>`);
    if (stats.changedPaths > 0) parts.push(`<span class="stat-changed">${stats.changedPaths} changed</span>`);
    if (stats.equalPaths > 0) parts.push(`<span class="stat-equal">${stats.equalPaths} equal</span>`);

    comparisonStats.innerHTML = parts.join(' | ');
}

/**
 * Render the comparison based on current view mode and filters
 */
function renderComparison() {
    if (!currentComparison) return;

    const showOnlyChanges = optShowOnlyChanges.checked;
    const showTopValues = optShowTopValues.checked;

    // Filter comparisons
    let comparisons = currentComparison.comparisons;
    if (showOnlyChanges) {
        comparisons = comparisons.filter(c => c.status !== 'equal');
    }

    // Render both views (only one is visible at a time)
    renderStructuredComparison(comparisons, showTopValues);
    renderSideBySideComparison(comparisons, showTopValues);
}

/**
 * Render the structured (merged table) comparison view
 */
function renderStructuredComparison(comparisons, showTopValues) {
    // Clear table
    comparisonTbody.innerHTML = '';

    // Render each comparison
    for (const comp of comparisons) {
        const tr = document.createElement('tr');

        // Add row class based on status
        if (comp.status === 'added') {
            tr.classList.add('row-added');
        } else if (comp.status === 'removed') {
            tr.classList.add('row-removed');
        } else if (comp.status === 'changed') {
            tr.classList.add('row-changed');
        }

        // Path cell
        const pathCell = document.createElement('td');
        pathCell.className = 'path-cell';
        pathCell.textContent = comp.path;
        pathCell.title = 'Click to copy jq command';
        pathCell.addEventListener('click', (e) => {
            e.stopPropagation();
            copyJqCommandForComparison(comp.path, 'left');
        });
        tr.appendChild(pathCell);

        // Left count
        const leftCount = comp.left ? comp.left.count : null;
        const leftCountCell = document.createElement('td');
        leftCountCell.className = leftCount !== null ? 'count-cell' : 'count-cell empty';
        leftCountCell.textContent = leftCount !== null ? leftCount.toLocaleString() : '—';
        tr.appendChild(leftCountCell);

        // Right count
        const rightCount = comp.right ? comp.right.count : null;
        const rightCountCell = document.createElement('td');
        rightCountCell.className = rightCount !== null ? 'count-cell' : 'count-cell empty';
        rightCountCell.textContent = rightCount !== null ? rightCount.toLocaleString() : '—';
        tr.appendChild(rightCountCell);

        // Delta count
        const deltaCountCell = document.createElement('td');
        deltaCountCell.className = 'count-cell';
        if (comp.countDelta > 0) {
            deltaCountCell.classList.add('delta-positive');
            deltaCountCell.textContent = '+' + comp.countDelta.toLocaleString();
        } else if (comp.countDelta < 0) {
            deltaCountCell.classList.add('delta-negative');
            deltaCountCell.textContent = comp.countDelta.toLocaleString();
        } else {
            deltaCountCell.classList.add('delta-zero');
            deltaCountCell.textContent = '0';
        }
        tr.appendChild(deltaCountCell);

        // Left objects
        const leftObjects = comp.left ? comp.left.objectHits : null;
        const leftObjectsCell = document.createElement('td');
        leftObjectsCell.className = leftObjects !== null ? 'count-cell' : 'count-cell empty';
        leftObjectsCell.textContent = leftObjects !== null ? leftObjects.toLocaleString() : '—';
        tr.appendChild(leftObjectsCell);

        // Right objects
        const rightObjects = comp.right ? comp.right.objectHits : null;
        const rightObjectsCell = document.createElement('td');
        rightObjectsCell.className = rightObjects !== null ? 'count-cell' : 'count-cell empty';
        rightObjectsCell.textContent = rightObjects !== null ? rightObjects.toLocaleString() : '—';
        tr.appendChild(rightObjectsCell);

        // Delta objects
        const deltaObjectsCell = document.createElement('td');
        deltaObjectsCell.className = 'count-cell';
        if (comp.objectsDelta > 0) {
            deltaObjectsCell.classList.add('delta-positive');
            deltaObjectsCell.textContent = '+' + comp.objectsDelta.toLocaleString();
        } else if (comp.objectsDelta < 0) {
            deltaObjectsCell.classList.add('delta-negative');
            deltaObjectsCell.textContent = comp.objectsDelta.toLocaleString();
        } else {
            deltaObjectsCell.classList.add('delta-zero');
            deltaObjectsCell.textContent = '0';
        }
        tr.appendChild(deltaObjectsCell);

        // Status badge
        const statusCell = document.createElement('td');
        const badge = document.createElement('span');
        badge.className = `status-badge status-${comp.status}`;
        badge.textContent = comp.status.toUpperCase();
        statusCell.appendChild(badge);
        tr.appendChild(statusCell);

        comparisonTbody.appendChild(tr);

        // Add top values row if enabled
        if (showTopValues && (comp.left || comp.right)) {
            const detailRow = createComparisonDetailRow(comp);
            comparisonTbody.appendChild(detailRow);
        }
    }

    // Show message if no results
    if (comparisons.length === 0) {
        const tr = document.createElement('tr');
        const showOnlyChanges = optShowOnlyChanges.checked;
        tr.innerHTML = `<td colspan="8" style="text-align: center; color: var(--text-secondary); padding: 24px;">
            ${showOnlyChanges ? 'No changes found (all paths are equal)' : 'No comparisons to display'}
        </td>`;
        comparisonTbody.appendChild(tr);
    }
}

/**
 * Render the side-by-side comparison view
 */
function renderSideBySideComparison(comparisons, showTopValues) {
    // Clear both tables
    sidebysideLeftTbody.innerHTML = '';
    sidebysideRightTbody.innerHTML = '';

    // Build left table (paths from left analysis only)
    const leftPaths = comparisons.filter(c => c.left);
    for (const comp of leftPaths) {
        const tr = document.createElement('tr');

        // Add row class based on status
        if (comp.status === 'removed') {
            tr.classList.add('row-removed');
        } else if (comp.status === 'changed') {
            tr.classList.add('row-changed');
        }

        // Path
        const pathCell = document.createElement('td');
        pathCell.className = 'path-cell clickable-path';
        pathCell.textContent = comp.path;
        pathCell.title = 'Click to copy jq command';
        pathCell.addEventListener('click', (e) => {
            e.stopPropagation();
            copyJqCommandForComparison(comp.path, 'left');
        });
        tr.appendChild(pathCell);

        // Count
        const countCell = document.createElement('td');
        countCell.className = 'count-cell';
        countCell.textContent = comp.left.count.toLocaleString();
        tr.appendChild(countCell);

        // Objects
        const objectsCell = document.createElement('td');
        objectsCell.className = 'count-cell';
        objectsCell.textContent = comp.left.objectHits.toLocaleString();
        tr.appendChild(objectsCell);

        // Distinct
        const distinctCell = document.createElement('td');
        distinctCell.className = 'count-cell';
        distinctCell.textContent = comp.left.distinctCount.toLocaleString();
        tr.appendChild(distinctCell);

        sidebysideLeftTbody.appendChild(tr);
    }

    // Build right table (paths from right analysis only)
    const rightPaths = comparisons.filter(c => c.right);
    for (const comp of rightPaths) {
        const tr = document.createElement('tr');

        // Add row class based on status
        if (comp.status === 'added') {
            tr.classList.add('row-added');
        } else if (comp.status === 'changed') {
            tr.classList.add('row-changed');
        }

        // Path
        const pathCell = document.createElement('td');
        pathCell.className = 'path-cell clickable-path';
        pathCell.textContent = comp.path;
        pathCell.title = 'Click to copy jq command';
        pathCell.addEventListener('click', (e) => {
            e.stopPropagation();
            copyJqCommandForComparison(comp.path, 'right');
        });
        tr.appendChild(pathCell);

        // Count
        const countCell = document.createElement('td');
        countCell.className = 'count-cell';
        countCell.textContent = comp.right.count.toLocaleString();
        tr.appendChild(countCell);

        // Objects
        const objectsCell = document.createElement('td');
        objectsCell.className = 'count-cell';
        objectsCell.textContent = comp.right.objectHits.toLocaleString();
        tr.appendChild(objectsCell);

        // Distinct
        const distinctCell = document.createElement('td');
        distinctCell.className = 'count-cell';
        distinctCell.textContent = comp.right.distinctCount.toLocaleString();
        tr.appendChild(distinctCell);

        sidebysideRightTbody.appendChild(tr);
    }

    // Show message if no results in left
    if (leftPaths.length === 0) {
        const tr = document.createElement('tr');
        tr.innerHTML = `<td colspan="4" style="text-align: center; color: var(--text-secondary); padding: 24px;">No paths in left file</td>`;
        sidebysideLeftTbody.appendChild(tr);
    }

    // Show message if no results in right
    if (rightPaths.length === 0) {
        const tr = document.createElement('tr');
        tr.innerHTML = `<td colspan="4" style="text-align: center; color: var(--text-secondary); padding: 24px;">No paths in right file</td>`;
        sidebysideRightTbody.appendChild(tr);
    }
}

/**
 * Create a detail row showing top values for both sides
 */
function createComparisonDetailRow(comp) {
    const tr = document.createElement('tr');
    tr.className = 'value-detail-row';

    const td = document.createElement('td');
    td.colSpan = 8;

    let html = '<div class="value-detail" style="display: flex; gap: 24px;">';

    // Left values
    if (comp.left && comp.left.topValues && comp.left.topValues.length > 0) {
        html += '<div style="flex: 1;"><strong>Left (Top Values):</strong><ul>';
        for (const v of comp.left.topValues.slice(0, 5)) {
            const displayValue = v.value.length > 50 ? v.value.substring(0, 50) + '...' : v.value;
            html += `<li><span class="value-text">${escapeHtml(displayValue)}</span> <span class="value-count">(${v.count.toLocaleString()})</span></li>`;
        }
        if (comp.left.distinctCount > 5) {
            html += `<li class="more-values">... and ${(comp.left.distinctCount - 5).toLocaleString()} more</li>`;
        }
        html += '</ul></div>';
    } else {
        html += '<div style="flex: 1;"><strong>Left:</strong> <em style="color: var(--text-secondary);">No values</em></div>';
    }

    // Right values
    if (comp.right && comp.right.topValues && comp.right.topValues.length > 0) {
        html += '<div style="flex: 1;"><strong>Right (Top Values):</strong><ul>';
        for (const v of comp.right.topValues.slice(0, 5)) {
            const displayValue = v.value.length > 50 ? v.value.substring(0, 50) + '...' : v.value;
            html += `<li><span class="value-text">${escapeHtml(displayValue)}</span> <span class="value-count">(${v.count.toLocaleString()})</span></li>`;
        }
        if (comp.right.distinctCount > 5) {
            html += `<li class="more-values">... and ${(comp.right.distinctCount - 5).toLocaleString()} more</li>`;
        }
        html += '</ul></div>';
    } else {
        html += '<div style="flex: 1;"><strong>Right:</strong> <em style="color: var(--text-secondary);">No values</em></div>';
    }

    html += '</div>';

    td.innerHTML = html;
    tr.appendChild(td);

    return tr;
}

/**
 * Copy a jq command for comparison mode (uses left file by default)
 */
async function copyJqCommandForComparison(path, side) {
    const filePath = side === 'left' ? compareLeftPath.value : compareRightPath.value;
    let textToCopy;

    if (filePath) {
        textToCopy = `cat '${filePath}' | grep -E '^\\{' | jq '${path}'`;
    } else {
        textToCopy = path;
    }

    try {
        await navigator.clipboard.writeText(textToCopy);
        showCopyFeedback('Copied!');
    } catch (err) {
        console.error('Failed to copy:', err);
    }
}

// ============================================================
// Initialization
// ============================================================

// Load file path history when the app starts
// This file is loaded as a module, so DOM is already ready
loadFilePathHistory();
