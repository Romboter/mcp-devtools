# RulesHub Tool

The RulesHub tool provides a simple and efficient way to manage and retrieve contextual rules for AI agents. It loads rules from YAML files and makes them available through the MCP protocol.

## Features

- **Simple Configuration**: Uses environment variables or defaults to user's home directory
- **YAML Rule Files**: Easy-to-write rule definitions in YAML format
- **Fast Loading**: Loads all rules into memory for quick access
- **Cross-Platform**: Works on both macOS and Linux
- **MCP Integration**: Seamlessly integrates with the Model Context Protocol

## Usage

The tool exposes two main actions through the MCP protocol:

### 1. GetAllRulesMetadata
Retrieves metadata for all available rules (without content).

```json
{
  "name": "ruleshub",
  "arguments": {
    "action": "GetAllRulesMetadata"
  }
}
```

**Response:**
```json
{
  "rules": [
    {
      "ruleId": "go-formatting",
      "description": "Go code formatting guidelines",
      "language": "go",
      "tags": ["formatting", "style"]
    }
  ],
  "count": 1
}
```

### 2. GetRuleContentById
Retrieves the complete rule including its content.

```json
{
  "name": "ruleshub",
  "arguments": {
    "action": "GetRuleContentById",
    "ruleId": "go-formatting"
  }
}
```

**Response:**
```json
{
  "ruleId": "go-formatting",
  "description": "Go code formatting guidelines",
  "language": "go",
  "tags": ["formatting", "style"],
  "content": "# Go Formatting Rules\n\n- Use gofmt for formatting\n- Keep lines under 100 characters\n..."
}
```

## Configuration

### Rules Directory

The tool looks for rules in the following order:

1. **Environment Variable**: `RULE_DIRECTORY` - Custom path to rules directory
2. **Default Location**: `~/.mcp-devtools/rules` - Default rules directory in user's home

**Examples:**
```bash
# Use custom directory
export RULE_DIRECTORY="/path/to/my/rules"

# Use relative path (will be converted to absolute)
export RULE_DIRECTORY="./project-rules"
```

## Rule File Format

Rules are defined in YAML files (`.yaml` or `.yml`) with the following structure:

```yaml
id: unique-rule-id
description: Human-readable description of the rule
language: programming-language  # Optional (e.g., "go", "python", "javascript")
tags:  # Optional list of tags
  - formatting
  - best-practices
  - style
rule: |
  # Rule Content
  
  This is the actual content of the rule.
  It can include:
  - Markdown formatting
  - Code examples
  - Guidelines and best practices
  - Any text-based content
```

### Required Fields
- `id`: Unique identifier for the rule (will be normalized to lowercase with spaces replaced by hyphens)
- `description`: Brief description of what the rule covers
- `rule`: The actual rule content

### Optional Fields
- `language`: Programming language this rule applies to
- `tags`: List of tags for categorization

### Example Rule File

**File: `~/.mcp-devtools/rules/go-error-handling.yaml`**
```yaml
id: go-error-handling
description: Best practices for error handling in Go
language: go
tags:
  - error-handling
  - best-practices
  - go
rule: |
  # Go Error Handling Best Practices
  
  ## Always Check Errors
  ```go
  result, err := someFunction()
  if err != nil {
      return fmt.Errorf("operation failed: %w", err)
  }
  ```
  
  ## Use Meaningful Error Messages
  - Include context about what operation failed
  - Use error wrapping with %w verb
  - Don't ignore errors unless absolutely necessary
  
  ## Custom Error Types
  ```go
  type ValidationError struct {
      Field string
      Value interface{}
  }
  
  func (e ValidationError) Error() string {
      return fmt.Sprintf("validation failed for field %s: %v", e.Field, e.Value)
  }
  ```
```

## Implementation

The tool is implemented as a single, simple Go file with the following key components:

- **Rule Struct**: Represents a rule with metadata and content
- **RuleHubTool**: Main tool implementation with MCP integration
- **File Loading**: Scans directory for YAML files and parses them
- **In-Memory Storage**: Keeps all rules in memory for fast access
- **Thread Safety**: Uses read-write mutexes for concurrent access

### Key Benefits of Simplified Design

1. **Easy to Understand**: Single file implementation, no complex abstractions
2. **Fast Performance**: All rules loaded in memory, no file I/O during requests
3. **Simple Testing**: Straightforward unit tests without complex mocking
4. **Easy Maintenance**: Clear code flow, minimal dependencies
5. **Reliable**: Fewer moving parts means fewer potential failure points

## Development

### Running Tests
```bash
go test ./internal/tools/ruleshub -v
```

### Building
```bash
make clean && make build
```

### Testing the Tool
```bash
# Create a test rule
mkdir -p ~/.mcp-devtools/rules
cat > ~/.mcp-devtools/rules/test-rule.yaml << EOF
id: test-rule
description: A simple test rule
tags: [test]
rule: |
  # Test Rule
  This is a test rule for demonstration.
EOF

# Test the tool
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "ruleshub", "arguments": {"action": "GetAllRulesMetadata"}}}' | ./bin/mcp-devtools stdio
```

## Error Handling

The tool handles various error conditions gracefully:

- **Missing Rules Directory**: Creates the directory automatically
- **No Rule Files**: Returns empty results (not an error)
- **Invalid YAML**: Logs warning and skips the file
- **Missing Required Fields**: Logs warning and skips the file
- **File Read Errors**: Logs warning and continues with other files

This ensures the tool remains functional even with problematic rule files.
