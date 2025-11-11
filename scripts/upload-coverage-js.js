#!/usr/bin/env node

/**
 * Upload JavaScript/TypeScript coverage to Librecov
 * 
 * Converts Istanbul/V8 coverage format to Coveralls format and uploads.
 * 
 * Usage: 
 *   PROJECT_TOKEN=your-token node upload-coverage-js.js
 *   PROJECT_TOKEN=your-token npm run upload:coverage
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const https = require('https');
const http = require('http');

// Configuration
const LIBRECOV_URL = process.env.LIBRECOV_URL || 'http://localhost:4000';
const PROJECT_TOKEN = process.env.PROJECT_TOKEN;
const COVERAGE_DIR = process.env.COVERAGE_DIR || 'coverage';

if (!PROJECT_TOKEN) {
    console.error('Error: PROJECT_TOKEN environment variable is required');
    console.error('Usage: PROJECT_TOKEN=your-token node upload-coverage-js.js');
    process.exit(1);
}

// Change to frontend directory
const frontendDir = path.join(__dirname, '..', 'frontend');
process.chdir(frontendDir);

console.log('==> Running tests with coverage...');
try {
    execSync('npm run test:coverage', { stdio: 'inherit' });
} catch (err) {
    console.error('Error running tests:', err.message);
    process.exit(1);
}

// Check if coverage file exists
const coverageFilePath = path.join(COVERAGE_DIR, 'coverage-final.json');
if (!fs.existsSync(coverageFilePath)) {
    console.error(`Error: Coverage file not found at ${coverageFilePath}`);
    process.exit(1);
}

console.log('\n==> Reading coverage data...');
const coverageData = JSON.parse(fs.readFileSync(coverageFilePath, 'utf8'));

// Get git information
function execGit(args) {
    try {
        return execSync(`git ${args}`, { encoding: 'utf8' }).trim();
    } catch {
        return '';
    }
}

const gitBranch = execGit('rev-parse --abbrev-ref HEAD') || 'main';
const gitCommit = execGit('rev-parse HEAD') || 'unknown';
const gitMessage = execGit('log -1 --pretty=%B') || 'No commit message';

console.log(`   Branch: ${gitBranch}`);
console.log(`   Commit: ${gitCommit.substring(0, 8)}`);

console.log('\n==> Converting coverage to Coveralls format...');

// Convert to Coveralls format
const sourceFiles = [];
let totalFiles = 0;
let processedFiles = 0;

for (const [filePath, fileData] of Object.entries(coverageData)) {
    totalFiles++;
    
    // Read source file
    let source = '';
    try {
        source = fs.readFileSync(filePath, 'utf8');
    } catch (err) {
        console.warn(`   Warning: Could not read source file: ${filePath}`);
        continue;
    }

    // Build coverage array (null for non-executable, 0 for not covered, >0 for covered)
    const lines = source.split('\n');
    const coverage = new Array(lines.length).fill(null);

    // Map statement coverage to line coverage
    if (fileData.statementMap && fileData.s) {
        for (const [key, stmt] of Object.entries(fileData.statementMap)) {
            const line = stmt.start.line - 1;
            if (line >= 0 && line < coverage.length) {
                const count = fileData.s[key];
                // Only update if not already set or this is a hit
                if (coverage[line] === null || count > 0) {
                    coverage[line] = count;
                }
            }
        }
    }

    // Also map function coverage
    if (fileData.fnMap && fileData.f) {
        for (const [key, fn] of Object.entries(fileData.fnMap)) {
            const line = fn.loc.start.line - 1;
            if (line >= 0 && line < coverage.length) {
                const count = fileData.f[key];
                if (coverage[line] === null || count > 0) {
                    coverage[line] = count;
                }
            }
        }
    }

    // Convert absolute path to relative
    const relativePath = path.relative(path.join(frontendDir, '..'), filePath);

    sourceFiles.push({
        name: relativePath,
        source: source,
        coverage: coverage
    });

    processedFiles++;
}

console.log(`   Processed ${processedFiles}/${totalFiles} files`);

// Calculate overall coverage
let totalLines = 0;
let coveredLines = 0;
sourceFiles.forEach(file => {
    file.coverage.forEach(count => {
        if (count !== null) {
            totalLines++;
            if (count > 0) coveredLines++;
        }
    });
});

const overallCoverage = totalLines > 0 ? (coveredLines / totalLines * 100).toFixed(2) : 0;
console.log(`   Overall coverage: ${overallCoverage}%`);

// Build Coveralls JSON
const coverallsJson = {
    repo_token: PROJECT_TOKEN,
    service_name: 'manual',
    git: {
        branch: gitBranch,
        head: {
            id: gitCommit,
            message: gitMessage
        }
    },
    source_files: sourceFiles
};

console.log(`\n==> Uploading coverage to Librecov at ${LIBRECOV_URL}...`);

// Upload to Librecov
const jsonData = JSON.stringify(coverallsJson);
const url = new URL(`${LIBRECOV_URL}/api/upload`);
const client = url.protocol === 'https:' ? https : http;

const options = {
    hostname: url.hostname,
    port: url.port || (url.protocol === 'https:' ? 443 : 80),
    path: url.pathname,
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(jsonData)
    }
};

const req = client.request(options, (res) => {
    let data = '';

    res.on('data', (chunk) => {
        data += chunk;
    });

    res.on('end', () => {
        if (res.statusCode === 200) {
            console.log('✓ Coverage uploaded successfully!');
            try {
                const response = JSON.parse(data);
                console.log(JSON.stringify(response, null, 2));
            } catch {
                console.log(data);
            }
            console.log('\nDone!');
        } else {
            console.error(`✗ Upload failed with status code ${res.statusCode}`);
            console.error(data);
            process.exit(1);
        }
    });
});

req.on('error', (err) => {
    console.error('✗ Upload failed:', err.message);
    process.exit(1);
});

req.write(jsonData);
req.end();
