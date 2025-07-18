# RulesHub Tool

The RulesHub tool is designed to manage and provide contextual rules for AI agents. It allows agents to dynamically retrieve rules based on various criteria such as programming language or specific rule identifiers.

## Overview

The RulesHub tool provides a centralized way to manage and access rules across different agents or projects. It enables agents to retrieve only relevant rules, minimizing context length and improving efficiency.

## Features

- **Dynamic Rule Retrieval**: Access rules based on language, ID, or other metadata
- **YAML-based Rule Storage**: Rules are defined in YAML files
- **Configurable Rule Sources**: Specify locations for rule files via environment variables
- **MCP Integration**: Exposes functionality as tools consumable by MCP clients

## Usage

The tool exposes two main functions:

### GetRuleContentById

Retrieves the content of a specific rule by its ID.

**Input Parameters**:
- `action`: (required) Must be set to `"GetRuleContentById"`
- `ruleId`: (required) The ID of the rule to retrieve

**Example**:
```json
{
  "name": "ruleshub",
  "arguments": {
    "action": "GetRuleContentById",
    "ruleId": "js-formatting"
  }
}
```

**Response**:
```json
{
  "ruleId": "js-formatting",
  "description": "JavaScript formatting guidelines",
  "language": "javascript",
  "tags": ["formatting", "style", "best-practices"],
  "content": "# JavaScript Formatting Guidelines\n\n## Indentation\n\nUse 2 spaces for indentation..."
}
```

### GetAllRulesMetadata

Retrieves metadata for all available rules.

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
      "ruleId": "js-formatting",
      "description": "JavaScript formatting guidelines",
      "language": "javascript",
      "tags": ["formatting", "style", "best-practices"]
    },
    {
      "ruleId": "python-imports",
      "description": "Python import organization guidelines",
      "language": "python",
      "tags": ["imports", "organization", "best-practices"]
    }
  ]
}
```

## Configuration

The tool is configured through environment variables:

- `RULESHUB_SOURCES_0_LOADERTYPE`: Type of rule loader (e.g., "YamlFile")
- `RULESHUB_SOURCES_0_SETTINGS_PATH`: Path to rule files

Multiple sources can be configured by incrementing the index:

```
RULESHUB_SOURCES_0_LOADERTYPE=YamlFile
RULESHUB_SOURCES_0_SETTINGS_PATH=/path/to/rules1
RULESHUB_SOURCES_1_LOADERTYPE=YamlFile
RULESHUB_SOURCES_1_SETTINGS_PATH=/path/to/rules2
```

## Rule File Format

Rules are defined in YAML files with the following structure:

```yaml
id: rule-id
description: Rule description
language: programming-language  # Optional
tags:  # Optional
  - tag1
  - tag2
rule: |
  # Rule content
  
  This is the actual content of the rule.
  It can include markdown formatting, code examples, etc.
```

### Example Rule File

```yaml
id: js-formatting
description: JavaScript formatting guidelines
language: javascript
tags:
  - formatting
  - style
  - best-practices
rule: |
  # JavaScript Formatting Guidelines
  
  ## Indentation
  
  Use 2 spaces for indentation. Do not use tabs.
  
  ```js
  // Good
  function example() {
    const x = 1;
    if (x > 0) {
      console.log('Positive');
    }
  }
  
  // Bad
  function example() {
      const x = 1;
      if (x > 0) {
          console.log('Positive');
      }
  }
  ```
  
  ## Line Length
  
  Keep lines under 80 characters when possible.
```

## Workflow Example

Here's an example of how an AI agent might use the RulesHub tool:

1. **Retrieve Rule Index**:
   ```json
   {
     "name": "ruleshub",
     "arguments": {
       "action": "GetAllRulesMetadata"
     }
   }
   ```

2. **Analyze Task & Identify Relevant Rules**:
   The agent analyzes the task (e.g., "Format this JavaScript code") and identifies relevant rules based on the metadata (e.g., rules with language="javascript" and tags=["formatting"]).

3. **Retrieve Rule Content**:
   ```json
   {
     "name": "ruleshub",
     "arguments": {
       "action": "GetRuleContentById",
       "ruleId": "js-formatting"
     }
   }
   ```

4. **Apply Rules to Task**:
   The agent uses the retrieved rule content to guide its actions (e.g., formatting JavaScript code according to the guidelines).

## Best Practices

- **Organize Rules Logically**: Group related rules together in the same directory
- **Use Clear IDs**: Make rule IDs descriptive and easy to understand
- **Include Tags**: Add relevant tags to make rules easier to find
- **Keep Rules Focused**: Each rule should address a specific concern
- **Use Markdown**: Format rule content with markdown for better readability
