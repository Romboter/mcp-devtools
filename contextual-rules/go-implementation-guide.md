# AgentRulesHub Go Implementation Guide

This document provides guidance for implementing the AgentRulesHub system in Go, including interfaces, data models, and implementation patterns.

## Table of Contents

1. [Core Interfaces](#core-interfaces)
2. [Data Models](#data-models)
3. [Implementation Patterns](#implementation-patterns)
4. [Package Structure](#package-structure)
5. [Error Handling](#error-handling)
6. [Configuration Management](#configuration-management)
7. [Concurrency Patterns](#concurrency-patterns)
8. [MCP Integration](#mcp-integration)
9. [Sample Implementation](#sample-implementation)

## Core Interfaces

### RuleLoader Interface

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

### RuleParser Interface

```go
// RuleParser defines the interface for parsing rule content
type RuleParser interface {
    // ParseRule parses a rule from the specified file path
    ParseRule(ctx context.Context, filePath string) (AgentRule, error)
}
```

### RuleRepository Interface

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

### RuleLoaderOrchestrator Interface

```go
// RuleLoaderOrchestrator defines the interface for coordinating rule loading
type RuleLoaderOrchestrator interface {
    // LoadRules loads rules from all configured sources
    LoadRules(ctx context.Context) ([]AgentRule, error)
}
```

### RuleSource Interface

```go
// RuleSource defines the interface for retrieving rule content
type RuleSource interface {
    // SourceType returns the type of the source (e.g., "File")
    SourceType() string
    
    // GetRuleContent retrieves the content of the rule
    GetRuleContent(ctx context.Context) (string, error)
}
```

## Data Models

### AgentRule Struct

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

### FileSource Struct

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

### YamlRuleContent Struct

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

### Configuration Structs

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

## Implementation Patterns

### Dependency Injection

Go doesn't have built-in dependency injection like .NET, but you can implement it using constructor functions:

```go
// NewYamlRuleLoader creates a new YamlRuleLoader with dependencies
func NewYamlRuleLoader(parser RuleParser) *YamlRuleLoader {
    return &YamlRuleLoader{
        parser: parser,
    }
}

// YamlRuleLoader implements RuleLoader for YAML files
type YamlRuleLoader struct {
    parser RuleParser
}

// Implement RuleLoader interface methods...
```

### Error Handling

Go uses explicit error handling rather than exceptions:

```go
func (l *YamlRuleLoader) LoadRules(ctx context.Context, options RuleSourceOptions) ([]AgentRule, error) {
    // Validate options
    if options.LoaderType == "" {
        return nil, errors.New("loader type is required")
    }
    
    // Extract path from settings
    pathInterface, ok := options.Settings["Path"]
    if !ok {
        return nil, errors.New("path setting is required")
    }
    
    path, ok := pathInterface.(string)
    if !ok || path == "" {
        return nil, errors.New("path must be a non-empty string")
    }
    
    // Check if directory exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf("directory not found: %s", path)
    }
    
    // Load rules
    // ...
}
```

### Concurrency

Go uses goroutines and channels for concurrency:

```go
func (l *YamlRuleLoader) LoadRules(ctx context.Context, options RuleSourceOptions) ([]AgentRule, error) {
    // ... validation code ...
    
    // Get YAML files
    files, err := filepath.Glob(filepath.Join(path, "*.yaml"))
    if err != nil {
        return nil, fmt.Errorf("error finding YAML files: %w", err)
    }
    
    // Use a wait group to wait for all goroutines to finish
    var wg sync.WaitGroup
    
    // Create a channel for results
    resultChan := make(chan struct {
        rule AgentRule
        err  error
    }, len(files))
    
    // Process each file in a goroutine
    for _, file := range files {
        wg.Add(1)
        go func(filePath string) {
            defer wg.Done()
            
            // Check for context cancellation
            select {
            case <-ctx.Done():
                resultChan <- struct {
                    rule AgentRule
                    err  error
                }{AgentRule{}, ctx.Err()}
                return
            default:
                // Continue processing
            }
            
            // Parse rule
            rule, err := l.parser.ParseRule(ctx, filePath)
            resultChan <- struct {
                rule AgentRule
                err  error
            }{rule, err}
        }(file)
    }
    
    // Close the channel when all goroutines are done
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // Collect results
    var rules []AgentRule
    var errors []error
    
    for result := range resultChan {
        if result.err != nil {
            errors = append(errors, result.err)
            continue
        }
        rules = append(rules, result.rule)
    }
    
    // Return rules and possibly an error
    if len(errors) > 0 {
        // Log errors but return partial results
        for _, err := range errors {
            log.Printf("Error parsing rule: %v", err)
        }
    }
    
    return rules, nil
}
```

## Package Structure

A typical Go project structure for AgentRulesHub might look like this:

```
agent-rules-hub/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   ├── agent_rule.go
│   │   ├── rule_source.go
│   │   └── yaml_rule.go
│   ├── loaders/
│   │   ├── loader.go
│   │   └── yaml_loader.go
│   ├── parsers/
│   │   ├── parser.go
│   │   └── yaml_parser.go
│   ├── repository/
│   │   └── memory_repository.go
│   ├── orchestrator/
│   │   └── loader_orchestrator.go
│   └── mcp/
│       └── tools.go
├── pkg/
│   └── mcp/
│       └── server.go
├── go.mod
└── go.sum
```

### Package Descriptions

- **cmd/server**: Contains the main application entry point
- **internal/config**: Configuration management
- **internal/models**: Data models
- **internal/loaders**: Rule loaders
- **internal/parsers**: Rule parsers
- **internal/repository**: Rule repository
- **internal/orchestrator**: Rule loader orchestrator
- **internal/mcp**: MCP tools
- **pkg/mcp**: MCP server implementation (can be used by other projects)

## Error Handling

Go uses explicit error handling rather than exceptions. Here are some best practices:

1. **Return errors explicitly**:
   ```go
   func DoSomething() (Result, error) {
       // ...
       if err != nil {
           return Result{}, fmt.Errorf("doing something: %w", err)
       }
       // ...
   }
   ```

2. **Use error wrapping**:
   ```go
   if err != nil {
       return fmt.Errorf("context: %w", err)
   }
   ```

3. **Check for specific errors**:
   ```go
   if errors.Is(err, os.ErrNotExist) {
       // Handle file not found
   }
   ```

4. **Create custom error types**:
   ```go
   type NotFoundError struct {
       ID string
   }

   func (e *NotFoundError) Error() string {
       return fmt.Sprintf("rule with ID %s not found", e.ID)
   }
   ```

## Configuration Management

Go has several options for configuration management:

### Using Viper

[Viper](https://github.com/spf13/viper) is a popular configuration library for Go:

```go
package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    RuleSources RuleSourcesOptions
}

func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    viper.SetConfigType("json") // or "yaml"
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("unmarshaling config: %w", err)
    }
    
    return &config, nil
}
```

### Environment Variables

Viper also supports environment variables:

```go
func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    viper.SetConfigType("json")
    
    // Set environment variable prefix
    viper.SetEnvPrefix("RULES_HUB")
    
    // Replace dots with underscores in env vars
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // Enable environment variables
    viper.AutomaticEnv()
    
    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("unmarshaling config: %w", err)
    }
    
    return &config, nil
}
```

## Concurrency Patterns

Go provides powerful concurrency primitives:

### Worker Pool Pattern

```go
func ProcessRules(ctx context.Context, rules []AgentRule, numWorkers int) []ProcessedRule {
    // Create input and output channels
    input := make(chan AgentRule, len(rules))
    output := make(chan ProcessedRule, len(rules))
    
    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for rule := range input {
                // Process rule
                result := ProcessRule(ctx, rule)
                
                // Send result to output channel
                select {
                case output <- result:
                    // Result sent
                case <-ctx.Done():
                    // Context cancelled
                    return
                }
            }
        }()
    }
    
    // Send rules to input channel
    go func() {
        for _, rule := range rules {
            select {
            case input <- rule:
                // Rule sent
            case <-ctx.Done():
                // Context cancelled
                break
            }
        }
        close(input)
    }()
    
    // Close output channel when all workers are done
    go func() {
        wg.Wait()
        close(output)
    }()
    
    // Collect results
    var results []ProcessedRule
    for result := range output {
        results = append(results, result)
    }
    
    return results
}
```

### Context for Cancellation

```go
func (o *RuleLoaderOrchestrator) LoadRules(ctx context.Context) ([]AgentRule, error) {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // ... load rules ...
    
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue
    }
    
    // ... continue loading rules ...
}
```

## MCP Integration

Implementing MCP server functionality in Go requires:

1. **Implementing the MCP protocol**
2. **Exposing tools via the protocol**
3. **Handling tool requests**

### MCP Server Implementation

```go
package mcp

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "os"
)

// Server represents an MCP server
type Server struct {
    tools map[string]ToolFunc
}

// ToolFunc is a function that handles a tool request
type ToolFunc func(ctx context.Context, args json.RawMessage) (interface{}, error)

// NewServer creates a new MCP server
func NewServer() *Server {
    return &Server{
        tools: make(map[string]ToolFunc),
    }
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(name string, fn ToolFunc) {
    s.tools[name] = fn
}

// Start starts the server using stdio transport
func (s *Server) Start(ctx context.Context) error {
    // Read from stdin, write to stdout
    go s.handleRequests(ctx, os.Stdin, os.Stdout)
    
    // Wait for context cancellation
    <-ctx.Done()
    return ctx.Err()
}

// handleRequests handles MCP requests
func (s *Server) handleRequests(ctx context.Context, r io.Reader, w io.Writer) {
    decoder := json.NewDecoder(r)
    encoder := json.NewEncoder(w)
    
    for {
        // Check for context cancellation
        select {
        case <-ctx.Done():
            return
        default:
            // Continue
        }
        
        // Read request
        var req Request
        if err := decoder.Decode(&req); err != nil {
            if err == io.EOF {
                return
            }
            log.Printf("Error decoding request: %v", err)
            continue
        }
        
        // Handle request
        resp := s.handleRequest(ctx, req)
        
        // Write response
        if err := encoder.Encode(resp); err != nil {
            log.Printf("Error encoding response: %v", err)
            continue
        }
    }
}

// Request represents an MCP request
type Request struct {
    ID        string          `json:"id"`
    ToolName  string          `json:"tool_name"`
    Arguments json.RawMessage `json:"arguments"`
}

// Response represents an MCP response
type Response struct {
    ID      string          `json:"id"`
    Result  interface{}     `json:"result,omitempty"`
    Error   string          `json:"error,omitempty"`
}

// handleRequest handles a single MCP request
func (s *Server) handleRequest(ctx context.Context, req Request) Response {
    // Find tool
    tool, ok := s.tools[req.ToolName]
    if !ok {
        return Response{
            ID:    req.ID,
            Error: fmt.Sprintf("tool not found: %s", req.ToolName),
        }
    }
    
    // Call tool
    result, err := tool(ctx, req.Arguments)
    if err != nil {
        return Response{
            ID:    req.ID,
            Error: err.Error(),
        }
    }
    
    // Return result
    return Response{
        ID:     req.ID,
        Result: result,
    }
}
```

### Registering MCP Tools

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/example/agent-rules-hub/internal/config"
    "github.com/example/agent-rules-hub/internal/mcp"
    "github.com/example/agent-rules-hub/internal/repository"
    "github.com/example/agent-rules-hub/pkg/mcp"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("config.json")
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }
    
    // Create repository
    repo := repository.NewInMemoryRepository()
    
    // Create MCP server
    server := mcp.NewServer()
    
    // Register tools
    server.RegisterTool("GetRuleContentById", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
        var params struct {
            RuleId string `json:"ruleId"`
        }
        
        if err := json.Unmarshal(args, &params); err != nil {
            return nil, fmt.Errorf("invalid arguments: %w", err)
        }
        
        rule, err := repo.GetRuleMetadataById(ctx, params.RuleId)
        if err != nil {
            return nil, err
        }
        
        if rule == nil {
            return nil, fmt.Errorf("rule not found: %s", params.RuleId)
        }
        
        content, err := rule.Source.GetRuleContent(ctx)
        if err != nil {
            return nil, err
        }
        
        return content, nil
    })
    
    server.RegisterTool("GetAllRulesMetadata", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
        return repo.GetAllRulesMetadata(ctx)
    })
    
    // Create context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        cancel()
    }()
    
    // Start server
    if err := server.Start(ctx); err != nil && err != context.Canceled {
        log.Fatalf("Error running server: %v", err)
    }
}
```

## Sample Implementation

Here's a sample implementation of the YamlRuleParser:

```go
package parsers

import (
    "context"
    "fmt"
    "os"
    
    "gopkg.in/yaml.v3"
    
    "github.com/example/agent-rules-hub/internal/models"
)

// YamlRuleParser implements RuleParser for YAML files
type YamlRuleParser struct{}

// NewYamlRuleParser creates a new YamlRuleParser
func NewYamlRuleParser() *YamlRuleParser {
    return &YamlRuleParser{}
}

// ParseRule parses a rule from a YAML file
func (p *YamlRuleParser) ParseRule(ctx context.Context, filePath string) (models.AgentRule, error) {
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return models.AgentRule{}, ctx.Err()
    default:
        // Continue
    }
    
    // Read file
    data, err := os.ReadFile(filePath)
    if err != nil {
        return models.AgentRule{}, fmt.Errorf("reading file: %w", err)
    }
    
    // Parse YAML
    var yamlContent models.YamlRuleContent
    if err := yaml.Unmarshal(data, &yamlContent); err != nil {
        return models.AgentRule{}, fmt.Errorf("parsing YAML: %w", err)
    }
    
    // Validate required fields
    if yamlContent.Id == "" {
        return models.AgentRule{}, fmt.Errorf("rule ID is required")
    }
    
    if yamlContent.Description == "" {
        return models.AgentRule{}, fmt.Errorf("rule description is required")
    }
    
    // Create tags slice if nil
    tags := yamlContent.Tags
    if tags == nil {
        tags = []string{}
    }
    
    // Create rule
    rule := models.AgentRule{
        RuleId:      yamlContent.Id,
        Description: yamlContent.Description,
        Language:    yamlContent.Language,
        Tags:        tags,
        Source:      &models.FileSource{FilePath: filePath},
    }
    
    return rule, nil
}
```

And here's a sample implementation of the InMemoryRepository:

```go
package repository

import (
    "context"
    "errors"
    "sync"
    
    "github.com/example/agent-rules-hub/internal/models"
)

// InMemoryRepository implements RuleRepository using an in-memory map
type InMemoryRepository struct {
    rules map[string]models.AgentRule
    mu    sync.RWMutex
}

// NewInMemoryRepository creates a new InMemoryRepository
func NewInMemoryRepository() *InMemoryRepository {
    return &InMemoryRepository{
        rules: make(map[string]models.AgentRule),
    }
}

// AddRuleMetadata adds a single rule to the repository
func (r *InMemoryRepository) AddRuleMetadata(ctx context.Context, rule models.AgentRule) error {
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue
    }
    
    // Validate rule
    if rule.RuleId == "" {
        return errors.New("rule ID is required")
    }
    
    // Add rule to map
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.rules[rule.RuleId] = rule
    return nil
}

// AddRulesMetadata adds multiple rules to the repository
func (r *InMemoryRepository) AddRulesMetadata(ctx context.Context, rules []models.AgentRule) error {
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue
    }
    
    // Add rules to map
    r.mu.Lock()
    defer r.mu.Unlock()
    
    for _, rule := range rules {
        // Validate rule
        if rule.RuleId == "" {
            // Log error and continue
            continue
        }
        
        r.rules[rule.RuleId] = rule
    }
    
    return nil
}

// GetRuleMetadataById retrieves a rule by its ID
func (r *InMemoryRepository) GetRuleMetadataById(ctx context.Context, ruleId string) (*models.AgentRule, error) {
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue
    }
    
    // Validate rule ID
    if ruleId == "" {
        return nil, errors.New("rule ID is required")
    }
    
    // Get rule from map
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    rule, ok := r.rules[ruleId]
    if !ok {
        return nil, nil
    }
    
    return &rule, nil
}

// GetAllRulesMetadata retrieves all rules in the repository
func (r *InMemoryRepository) GetAllRulesMetadata(ctx context.Context) ([]models.AgentRule, error) {
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue
    }
    
    // Get all rules from map
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    rules := make([]models.AgentRule, 0, len(r.rules))
    for _, rule := range r.rules {
        rules = append(rules, rule)
    }
    
    return rules, nil
}
