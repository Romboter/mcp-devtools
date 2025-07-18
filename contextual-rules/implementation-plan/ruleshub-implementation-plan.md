# AgentRulesHub Implementation Plan

This document outlines the implementation plan for integrating the AgentRulesHub functionality into the MCP DevTools server. The AgentRulesHub is a tool designed to manage and provide contextual rules for AI agents, allowing them to dynamically retrieve rules based on various criteria such as programming language or specific rule identifiers.

## Table of Contents

- [AgentRulesHub Implementation Plan](#agentruleshub-implementation-plan)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Implementation Phases](#implementation-phases)
    - [Phase 1: Project Setup and Core Models](#phase-1-project-setup-and-core-models)
    - [Phase 2: Rule Loading and Parsing](#phase-2-rule-loading-and-parsing)
    - [Phase 3: MCP Tool Integration](#phase-3-mcp-tool-integration)
    - [Phase 4: Testing and Documentation](#phase-4-testing-and-documentation)
  - [Directory Structure](#directory-structure)
  - [Implementation Details](#implementation-details)
    - [Core Interfaces](#core-interfaces)
      - [RuleLoader Interface](#ruleloader-interface)
      - [RuleParser Interface](#ruleparser-interface)
      - [RuleRepository Interface](#rulerepository-interface)
      - [RuleLoaderOrchestrator Interface](#ruleloaderorchestrator-interface)
      - [RuleSource Interface](#rulesource-interface)
    - [Data Models](#data-models)
      - [AgentRule Struct](#agentrule-struct)
      - [FileSource Struct](#filesource-struct)
      - [YamlRuleContent Struct](#yamlrulecontent-struct)
      - [Configuration Structs](#configuration-structs)
    - [MCP Tool Definition](#mcp-tool-definition)
    - [Configuration](#configuration)
    - [Error Handling](#error-handling)
  - [Testing Strategy](#testing-strategy)
  - [Documentation Requirements](#documentation-requirements)

## Overview

The AgentRulesHub tool will be integrated into the existing MCP DevTools server, following the project's structure and guidelines. The tool will provide functionality for:

- Loading rules from YAML files
- Storing rule metadata in memory
- Retrieving rules based on criteria
- Exposing rule functionality through MCP tools

## Implementation Phases

### Phase 1: Project Setup and Core Models

1. **Create Directory Structure**
   - [ ] Create `internal/tools/ruleshub/` directory
   - [ ] Create `internal/tools/ruleshub/README.md` with tool documentation

2. **Implement Core Data Models**
   - [ ] Create `internal/tools/ruleshub/types.go` with:
     - [ ] `AgentRule` struct (rule metadata)
     - [ ] `RuleSource` interface
     - [ ] `FileSource` struct (implementation for file-based rules)
     - [ ] `YamlRuleContent` struct (for parsing YAML files)
     - [ ] Configuration structs for rule sources

3. **Implement Rule Repository**
   - [ ] Create `internal/tools/ruleshub/repository.go` with:
     - [ ] `RuleRepository` interface
     - [ ] `InMemoryRepository` implementation
     - [ ] Methods for adding and retrieving rules

### Phase 2: Rule Loading and Parsing

4. **Implement Rule Parser**
   - [ ] Create `internal/tools/ruleshub/parser.go` with:
     - [ ] `RuleParser` interface
     - [ ] `YamlRuleParser` implementation
     - [ ] Methods for parsing rule files

5. **Implement Rule Loader**
   - [ ] Create `internal/tools/ruleshub/loader.go` with:
     - [ ] `RuleLoader` interface
     - [ ] `YamlRuleLoader` implementation
     - [ ] Methods for loading rules from sources

6. **Implement Rule Loader Orchestrator**
   - [ ] Create `internal/tools/ruleshub/orchestrator.go` with:
     - [ ] `RuleLoaderOrchestrator` interface
     - [ ] Implementation for coordinating rule loading from multiple sources

### Phase 3: MCP Tool Integration

7. **Implement MCP Tools**
   - [ ] Create `internal/tools/ruleshub/ruleshub.go` with:
     - [ ] Tool struct implementing the `tools.Tool` interface
     - [ ] `init()` function to register the tool
     - [ ] `Definition()` method defining tool parameters
     - [ ] `Execute()` method implementing tool logic for:
       - [ ] `GetRuleContentById` functionality
       - [ ] `GetAllRulesMetadata` functionality

8. **Configuration Management**
   - [ ] Implement configuration loading from environment variables
   - [ ] Add validation for configuration

### Phase 4: Testing and Documentation

9. **Unit Testing**
   - [ ] Create `tests/tools/ruleshub_test.go` with:
     - [ ] Tests for rule parsing
     - [ ] Tests for rule loading
     - [ ] Tests for MCP tool functionality

10. **Documentation**
    - [ ] Update main `README.md` to mention the new tool
    - [ ] Create `docs/tools/ruleshub.md` with detailed documentation
    - [ ] Document configuration options

11. **Integration Testing**
    - [ ] Test the tool with sample rule files
    - [ ] Test integration with MCP client

## Directory Structure

```
internal/tools/ruleshub/
├── README.md
├── types.go
├── repository.go
├── parser.go
├── loader.go
├── orchestrator.go
└── ruleshub.go

docs/tools/
└── ruleshub.md

tests/tools/
└── ruleshub_test.go
```

## Implementation Details

### Core Interfaces

#### RuleLoader Interface

```go
// RuleLoader defines the interface for loading rules from a specific source type
type RuleLoader interface {
    // LoaderType returns the type of loader (e.g., "YamlFile")
    LoaderType() string
    
    // CanHandle checks if this loader can handle the specified loader type
    CanHandle(loaderType string) bool
    
    // LoadRules loads rules from the specified source options
    LoadRules(ctx context.Context, options RuleSourceOptions) ([]AgentRule, error)
}
```

#### RuleParser Interface

```go
// RuleParser defines the interface for parsing rule content
type RuleParser interface {
    // ParseRule parses a rule from the specified file path
    ParseRule(ctx context.Context, filePath string) (AgentRule, error)
}
```

#### RuleRepository Interface

```go
// RuleRepository defines the interface for storing and retrieving rule metadata
type RuleRepository interface {
    // AddRuleMetadata adds a single rule to the repository
    AddRuleMetadata(ctx context.Context, rule AgentRule) error
    
    // AddRulesMetadata adds multiple rules to the repository
    AddRulesMetadata(ctx context.Context, rules []AgentRule) error
    
    // GetRuleMetadataById retrieves a rule by its ID
    GetRuleMetadataById(ctx context.Context, ruleId string) (*AgentRule, error)
    
    // GetAllRulesMetadata retrieves all rules in the repository
    GetAllRulesMetadata(ctx context.Context) ([]AgentRule, error)
}
```

#### RuleLoaderOrchestrator Interface

```go
// RuleLoaderOrchestrator defines the interface for coordinating rule loading
type RuleLoaderOrchestrator interface {
    // LoadRules loads rules from all configured sources
    LoadRules(ctx context.Context) ([]AgentRule, error)
}
```

#### RuleSource Interface

```go
// RuleSource defines the interface for retrieving rule content
type RuleSource interface {
    // SourceType returns the type of the source (e.g., "File")
    SourceType() string
    
    // GetRuleContent retrieves the content of the rule
    GetRuleContent(ctx context.Context) (string, error)
}
```

### Data Models

#### AgentRule Struct

```go
// AgentRule represents a rule with its metadata
type AgentRule struct {
    RuleId      string     `json:"ruleId"`
    Description string     `json:"description"`
    Language    string     `json:"language,omitempty"`
    Tags        []string   `json:"tags"`
    Source      RuleSource `json:"-"`
}
```

#### FileSource Struct

```go
// FileSource implements RuleSource for file-based rules
type FileSource struct {
    FilePath string
}

func (fs *FileSource) SourceType() string {
    return "File"
}

func (fs *FileSource) GetRuleContent(ctx context.Context) (string, error) {
    if fs.FilePath == "" {
        return "", errors.New("file path is not set")
    }
    
    data, err := os.ReadFile(fs.FilePath)
    if err != nil {
        return "", fmt.Errorf("error reading file: %w", err)
    }
    
    var yamlContent map[string]interface{}
    if err := yaml.Unmarshal(data, &yamlContent); err != nil {
        return "", fmt.Errorf("error parsing YAML: %w", err)
    }
    
    rule, ok := yamlContent["rule"].(string)
    if !ok {
        return "", errors.New("rule content not found or not a string")
    }
    
    return rule, nil
}
```

#### YamlRuleContent Struct

```go
// YamlRuleContent represents the structure of a YAML rule file
type YamlRuleContent struct {
    Id          string   `yaml:"id"`
    Description string   `yaml:"description"`
    Language    string   `yaml:"language,omitempty"`
    Tags        []string `yaml:"tags,omitempty"`
    Rule        string   `yaml:"rule"`
}
```

#### Configuration Structs

```go
// RuleSourceOptions represents configuration for a rule source
type RuleSourceOptions struct {
    LoaderType string                 `json:"loaderType"`
    Settings   map[string]interface{} `json:"settings"`
}

// RuleSourcesOptions represents configuration for all rule sources
type RuleSourcesOptions struct {
    Sources []RuleSourceOptions `json:"sources"`
}
```

### MCP Tool Definition

The tool will implement the `tools.Tool` interface as defined in `internal/tools/tools.go`:

```go
type Tool interface {
    // Definition returns the tool's definition for MCP registration
    Definition() mcp.Tool

    // Execute executes the tool's logic
    Execute(ctx context.Context, logger *logrus.Logger, cache *sync.Map, args map[string]interface{}) (*mcp.CallToolResult, error)
}
```

The tool will expose two main functions:
1. `GetRuleContentById` - Retrieves the content of a specific rule by its ID
2. `GetAllRulesMetadata` - Retrieves metadata for all available rules

Example Definition:

```go
func (t *RuleHubTool) Definition() mcp.Tool {
    return mcp.NewTool(
        "ruleshub",
        mcp.WithDescription("A tool for managing and providing contextual rules for AI agents"),
        mcp.WithString("action",
            mcp.Required(),
            mcp.Description("The action to perform: 'GetRuleContentById' or 'GetAllRulesMetadata'"),
            mcp.Enum("GetRuleContentById", "GetAllRulesMetadata"),
        ),
        mcp.WithString("ruleId",
            mcp.Description("The ID of the rule to retrieve (required for GetRuleContentById)"),
        ),
    )
}
```

### Configuration

The tool will be configurable through environment variables, following the pattern used in other tools in the project. Configuration options will include:

- Rule sources (type and settings)
- Path to rule files
- Other loader-specific settings

### Error Handling

The implementation will include robust error handling, including:

- Validation of input parameters
- Graceful handling of missing files or invalid YAML
- Proper error propagation
- Detailed error messages
- Logging of errors

## Testing Strategy

The testing strategy will include:

1. **Unit Tests**:
   - Test each component in isolation
   - Mock dependencies for focused testing
   - Test error handling and edge cases

2. **Integration Tests**:
   - Test the complete flow from rule loading to retrieval
   - Test with sample rule files
   - Test MCP tool functionality

3. **Manual Testing**:
   - Test integration with MCP client
   - Verify rule content retrieval
   - Test with various rule formats and structures

## Documentation Requirements

Documentation will include:

1. **README.md**:
   - Overview of the tool
   - Installation and configuration instructions
   - Usage examples

2. **Tool Documentation**:
   - Detailed description of the tool's functionality
   - Parameter descriptions
   - Example requests and responses

3. **Rule Format Documentation**:
   - Description of the YAML rule format
   - Example rule files
   - Required and optional fields
