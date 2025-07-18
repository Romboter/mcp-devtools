# RulesHub Tool

The RulesHub tool provides a simple and efficient way to manage and retrieve contextual rules for AI agents. It loads rules from YAML files and makes them available through the MCP protocol.

## Overview

The RulesHub tool provides a centralized way to manage and access rules for AI agents. It enables agents to retrieve specific rules or browse all available rules, helping maintain consistent coding standards, best practices, and guidelines across projects.

## Features

- **Simple Configuration**: Uses environment variables or defaults to user's home directory
- **YAML Rule Files**: Easy-to-write rule definitions in YAML format
- **Fast Loading**: Loads all rules into memory for quick access
- **Cross-Platform**: Works on both macOS and Linux
- **MCP Integration**: Seamlessly integrates with the Model Context Protocol

## Usage

The tool exposes two main actions through the MCP protocol:

### GetAllRulesMetadata

Retrieves metadata for all available rules (without content).

**Input Parameters**:
- `action`: (required) Must be set to `"GetAllRulesMetadata"`

**Example**:
```json
{
  "name": "ruleshub",
  "arguments": {
    "action": "GetAllRulesMetadata"
  }
}
```

**Response**:
```json
{
  "rules": [
    {
      "ruleId": "go-formatting",
      "description": "Go code formatting guidelines",
      "language": "go",
      "tags": ["formatting", "style"]
    },
    {
      "ruleId": "python-imports",
      "description": "Python import organization guidelines",
      "language": "python",
      "tags": ["imports", "organization"]
    }
  ],
  "count": 2
}
```

### GetRuleContentById

Retrieves the complete rule including its content.

**Input Parameters**:
- `action`: (required) Must be set to `"GetRuleContentById"`
- `ruleId`: (required) The ID of the rule to retrieve

**Example**:
```json
{
  "name": "ruleshub",
  "arguments": {
    "action": "GetRuleContentById",
    "ruleId": "go-formatting"
  }
}
```

**Response**:
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

**Examples**:
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

## Workflow Example

Here's an example of how an AI agent might use the RulesHub tool:

1. **Browse Available Rules**:
   ```json
   {
     "name": "ruleshub",
     "arguments": {
       "action": "GetAllRulesMetadata"
     }
   }
   ```

2. **Analyze Task & Identify Relevant Rules**:
   The agent analyzes the task (e.g., "Review this Go code") and identifies relevant rules based on the metadata (e.g., rules with `language="go"` and `tags=["error-handling"]`).

3. **Retrieve Specific Rule Content**:
   ```json
   {
     "name": "ruleshub",
     "arguments": {
       "action": "GetRuleContentById",
       "ruleId": "go-error-handling"
     }
   }
   ```

4. **Apply Rules to Task**:
   The agent uses the retrieved rule content to guide its analysis and provide feedback according to the established guidelines.

## Testing the Tool

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

## Best Practices

- **Organize Rules Logically**: Use descriptive filenames and organize by topic or language
- **Use Clear IDs**: Make rule IDs descriptive and easy to understand
- **Include Tags**: Add relevant tags to make rules easier to categorize and find
- **Keep Rules Focused**: Each rule should address a specific concern or topic
- **Use Markdown**: Format rule content with markdown for better readability
- **Version Control**: Keep your rules directory in version control to track changes
- **Share Rules**: Consider sharing useful rules with your team or the community

## Error Handling

The tool handles various error conditions gracefully:

- **Missing Rules Directory**: Creates the directory automatically
- **No Rule Files**: Returns empty results (not an error)
- **Invalid YAML**: Logs warning and skips the file
- **Missing Required Fields**: Logs warning and skips the file
- **File Read Errors**: Logs warning and continues with other files

This ensures the tool remains functional even with problematic rule files.

## Implementation Details

The tool is implemented as a single, efficient Go file with the following key features:

- **In-Memory Storage**: All rules are loaded into memory for fast access
- **Thread Safety**: Uses read-write mutexes for concurrent access
- **Simple Design**: No complex abstractions, easy to understand and maintain
- **Cross-Platform**: Works consistently across different operating systems
- **Minimal Dependencies**: Only requires the Go standard library and YAML parsing
