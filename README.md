# License Audit

A comprehensive license auditing tool for various package managers written in Go. Scan your project dependencies and generate detailed license reports with audit capabilities.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

- **Multi-Language Support**: Scans Node.js, Go, Python, Ruby, Java, and Docker projects
- **License Detection**: Automatically extracts license information from packages
- **Audit Engine**: Identifies dangerous, unclear, or tainted licenses
- **Multiple Output Formats**: Generate reports in JSON or Markdown format
- **Configurable**: TOML configuration files with home directory and project-level support
- **Ignore Patterns**: `.licignore` file support similar to `.gitignore`
- **License Overrides**: Manual license specification for packages
- **CI/CD Integration**: Exit codes for pipeline integration

## Installation

### From Source

```bash
git clone https://github.com/your-username/license-audit.git
cd license-audit
go build -o license-audit
```

### Using Go Install

```bash
go install github.com/your-username/license-audit@latest
```

## Quick Start

```bash
# Scan current directory and output JSON to stdout
license-audit

# Scan specific path and output to file
license-audit --path ./my-project --output license-report.json

# Generate Markdown report
license-audit --format markdown --output report.md

# Disable auditing (only generate dependency list)
license-audit --audit=false
```

## Supported Package Managers

| Language/Tool | Files Detected | License Sources |
|---------------|----------------|-----------------|
| **Node.js** | `package.json`, `package-lock.json` | package.json license field, node_modules LICENSE files |
| **Go** | `go.mod`, `go.sum` | Module cache, vendor directory, LICENSE files |
| **Python** | `requirements.txt` | PyPI metadata (planned), LICENSE files |
| **Ruby** | `Gemfile`, `Gemfile.lock` | Gem metadata (planned), LICENSE files |
| **Java** | `pom.xml`, `build.gradle` | Maven/Gradle metadata (planned) |
| **Docker** | `Dockerfile` | Base images, package manager commands |

## Configuration

License Audit uses TOML configuration files. It searches for configuration in this order:

1. `~/.license-audit.toml` (global config)
2. `./.license-audit.toml` (project config)
3. File specified with `--config` flag

### Configuration Example

```toml
# Output settings
output_format = "json"  # or "markdown"
output_file = "license-report.json"
enable_audit = true

# Paths to scan
scan_paths = [".", "./subproject"]

# Ignore file (similar to .gitignore)
ignore_file = ".licignore"

# License overrides for packages without clear license info
[license_overrides]
"some-package" = "MIT"
"another-package" = "Apache-2.0"

# Define what constitutes dangerous licenses
dangerous_licenses = [
  "GPL-2.0", "GPL-3.0", "AGPL-3.0",
  "LGPL-2.1", "LGPL-3.0",
  "CDDL-1.0", "CDDL-1.1",
  "EPL-1.0", "EPL-2.0"
]

# Define unclear license indicators
unclear_licenses = [
  "UNKNOWN", "UNLICENSED", "PROPRIETARY", "COMMERCIAL", ""
]

# Scanner configuration
[scanners]
nodejs = true
go = true
docker = true
python = true
ruby = true
java = true
```

### Default Configuration

To generate a default configuration file:

```bash
license-audit --generate-config > .license-audit.toml
```

## Ignore Patterns (.licignore)

Create a `.licignore` file to exclude files and directories from scanning:

```gitignore
# Ignore common directories
node_modules/
vendor/
.git/
build/
dist/

# Ignore temporary files
*.tmp
*.temp

# Ignore specific packages
some-internal-package
test-*

# Don't ignore important test files
!test-important-package
```

## Output Formats

### JSON Output

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "scan_path": "./",
  "dependencies": [
    {
      "name": "express",
      "version": "4.18.0",
      "license_type": "MIT",
      "license_text": "MIT License...",
      "repository": "https://github.com/expressjs/express",
      "homepage": "http://expressjs.com/",
      "package_type": "npm",
      "file_path": "./package.json"
    }
  ],
  "issues": [
    {
      "severity": "error",
      "type": "dangerous_license",
      "message": "GPL-3.0 is a copyleft license...",
      "dependency": {...},
      "suggestion": "Consider using MIT or Apache-2.0 alternatives"
    }
  ],
  "summary": {
    "total_dependencies": 150,
    "license_breakdown": {
      "MIT": 89,
      "Apache-2.0": 23,
      "BSD-3-Clause": 15,
      "UNKNOWN": 23
    },
    "package_breakdown": {
      "npm": 120,
      "go": 30
    }
  }
}
```

### Markdown Output

```markdown
# License Audit Report

**Generated:** 2024-01-15 10:30:00
**Scan Path:** ./

## Summary

- **Total Dependencies:** 150
- **Issues Found:** 5

### License Distribution

| License | Count |
|---------|-------|
| MIT | 89 |
| Apache-2.0 | 23 |
| BSD-3-Clause | 15 |
| UNKNOWN | 23 |

## Issues

### 🚨 Errors

#### some-gpl-package

- **Type:** Dangerous License
- **Message:** GPL-3.0 is a copyleft license that may require releasing your source code
- **Package:** some-gpl-package@1.0.0 (npm)
- **Suggestion:** Consider using MIT or Apache-2.0 alternatives

## Dependencies

| Name | Version | License | Type | File Path |
|------|---------|---------|------|-----------|
| express | 4.18.0 | MIT | npm | ./package.json |
```

## License Auditing

The audit engine categorizes issues into three types:

### Dangerous Licenses (Error Level)
- **GPL family**: GPL-2.0, GPL-3.0, AGPL-3.0
- **LGPL family**: LGPL-2.1, LGPL-3.0  
- **Copyleft licenses**: CDDL, EPL, OSL, QPL
- **Impact**: May require open-sourcing your code

### Unclear Licenses (Warning Level)
- **UNKNOWN**: No license information found
- **PROPRIETARY**: Custom proprietary license
- **UNLICENSED**: Explicitly no license
- **Impact**: Legal uncertainty

### Tainted Licenses (Warning Level)
- **Dual licenses**: Multiple conflicting licenses
- **Modified licenses**: Custom modifications to standard licenses
- **Mixed terms**: Commercial + GPL combinations
- **Impact**: Complex legal requirements

## CI/CD Integration

License Audit is designed for CI/CD pipelines:

```bash
# In your CI script
license-audit --format json --output license-report.json

# The tool exits with code 1 if issues are found
if [ $? -eq 1 ]; then
  echo "License issues found. Check the report."
  exit 1
fi
```

### GitHub Actions Example

```yaml
name: License Audit
on: [push, pull_request]

jobs:
  license-audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.21'
          
      - name: Install License Audit
        run: go install github.com/your-username/license-audit@latest
        
      - name: Run License Audit
        run: license-audit --format markdown --output license-report.md
        
      - name: Upload Report
        uses: actions/upload-artifact@v3
        with:
          name: license-report
          path: license-report.md
```

## Advanced Usage

### Scanning Multiple Projects

```bash
license-audit \
  --path ./frontend \
  --path ./backend \
  --path ./mobile \
  --format markdown \
  --output combined-report.md
```

### Custom Configuration

```bash
# Use specific config file
license-audit --config ./custom-config.toml

# Override specific settings
license-audit \
  --format json \
  --output custom-report.json \
  --audit=true
```

### Integration with Other Tools

```bash
# Generate JSON and pipe to jq for processing
license-audit --format json | jq '.dependencies[] | select(.license_type == "UNKNOWN")'

# Generate report and send to Slack/Teams
license-audit --format markdown | curl -X POST -H 'Content-Type: application/json' \
  -d "{\"text\": \"$(cat)\"}" YOUR_WEBHOOK_URL
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o license-audit
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Troubleshooting

### Common Issues

**Q: License information is not found for Go modules**
A: Ensure the `go` command is available and the module cache is populated by running `go mod download`.

**Q: Node.js packages show UNKNOWN licenses**
A: Check if `node_modules` exists and contains the packages. Run `npm install` if needed.

**Q: Configuration file is not loaded**
A: Verify the file path and TOML syntax. Use `--config` to specify the exact path.

**Q: Ignore patterns are not working**
A: Check `.licignore` syntax and ensure patterns match the file paths being scanned.

### Getting Help

- 📚 [Documentation](https://github.com/your-username/license-audit/wiki)
- 🐛 [Issue Tracker](https://github.com/your-username/license-audit/issues)
- 💬 [Discussions](https://github.com/your-username/license-audit/discussions)

---

Made with ❤️ for open source compliance