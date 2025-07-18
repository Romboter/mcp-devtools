package ruleshub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RuleLoader defines the interface for loading rules from a specific source type
type RuleLoader interface {
	// LoaderType returns the type of loader (e.g., "YamlFile")
	LoaderType() string

	// CanHandle checks if this loader can handle the specified loader type
	CanHandle(loaderType string) bool

	// LoadRules loads rules from the specified source options
	LoadRules(ctx context.Context, options RuleSourceOptions) ([]AgentRule, error)
}

// YamlRuleLoader implements RuleLoader for YAML files
type YamlRuleLoader struct {
	parser RuleParser
}

// NewYamlRuleLoader creates a new YamlRuleLoader
func NewYamlRuleLoader(parser RuleParser) *YamlRuleLoader {
	return &YamlRuleLoader{
		parser: parser,
	}
}

// LoaderType returns the type of loader
func (l *YamlRuleLoader) LoaderType() string {
	return "YamlFile"
}

// CanHandle checks if this loader can handle the specified loader type
func (l *YamlRuleLoader) CanHandle(loaderType string) bool {
	return strings.EqualFold(loaderType, l.LoaderType())
}

// LoadRules loads rules from the specified source options
func (l *YamlRuleLoader) LoadRules(ctx context.Context, options RuleSourceOptions) ([]AgentRule, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Continue
	}

	// Validate options
	if options.LoaderType == "" {
		return nil, fmt.Errorf("loader type is required")
	}

	// Extract path from settings
	pathInterface, ok := options.Settings["Path"]
	if !ok {
		return nil, fmt.Errorf("path setting is required")
	}

	path, ok := pathInterface.(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("path must be a non-empty string")
	}

	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory not found: %s", path)
	}

	// Get YAML files
	files, err := filepath.Glob(filepath.Join(path, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("finding YAML files: %w", err)
	}

	// Parse each file
	var rules []AgentRule
	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue
		}

		rule, err := l.parser.ParseRule(ctx, file)
		if err != nil {
			// Log error and continue
			fmt.Printf("Error parsing rule file %s: %v\n", file, err)
			continue
		}

		rules = append(rules, rule)
	}

	return rules, nil
}
