# AgentRulesHub Tool

The AgentRulesHub tool is designed to manage and provide contextual rules for AI agents. It allows agents to dynamically retrieve rules based on various criteria such as programming language or specific rule identifiers.

## Features

- Load rules from YAML files
- Store rule metadata in memory
- Retrieve rules based on criteria
- Expose rule functionality through MCP tools

## Usage

The tool exposes two main functions through the MCP protocol:

1. `GetRuleContentById` - Retrieves the content of a specific rule by its ID
2. `GetAllRulesMetadata` - Retrieves metadata for all available rules

### Example MCP Tool Call

```json
{
  "name": "ruleshub",
  "arguments": {
    "action": "GetRuleContentById",
    "ruleId": "js-formatting"
  }
}
```

## Configuration

The tool is configurable through environment variables:

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

## Implementation

The tool is implemented with the following components:

- `types.go`: Data models for rules and rule sources
- `repository.go`: In-memory repository for rule metadata
- `parser.go`: Parser for rule content
- `loader.go`: Loader for rules from different sources
- `orchestrator.go`: Orchestrator for coordinating rule loading
- `ruleshub.go`: MCP tool implementation

## Development

See the [implementation plan](../../../contextual-rules/implementation-plan/ruleshub-implementation-plan.md) for detailed information about the implementation approach and tasks.
