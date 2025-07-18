package ruleshub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RuleParser defines the interface for parsing rule content
type RuleParser interface {
	// ParseRule parses a rule from the specified file path
	ParseRule(ctx context.Context, filePath string) (AgentRule, error)
}

// YamlRuleParser implements RuleParser for YAML files
type YamlRuleParser struct{}

// NewYamlRuleParser creates a new YamlRuleParser
func NewYamlRuleParser() *YamlRuleParser {
	return &YamlRuleParser{}
}

// ParseRule parses a rule from a YAML file
func (p *YamlRuleParser) ParseRule(ctx context.Context, filePath string) (AgentRule, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return AgentRule{}, ctx.Err()
	default:
		// Continue
	}

	// Validate file path
	if filePath == "" {
		return AgentRule{}, fmt.Errorf("file path is required")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return AgentRule{}, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return AgentRule{}, fmt.Errorf("reading file: %w", err)
	}

	// Parse YAML
	var yamlContent YamlRuleContent
	if err := yaml.Unmarshal(data, &yamlContent); err != nil {
		return AgentRule{}, fmt.Errorf("parsing YAML: %w", err)
	}

	// Validate required fields
	if yamlContent.Id == "" {
		return AgentRule{}, fmt.Errorf("rule ID is required in file: %s", filePath)
	}

	if yamlContent.Description == "" {
		return AgentRule{}, fmt.Errorf("rule description is required in file: %s", filePath)
	}

	if yamlContent.Rule == "" {
		return AgentRule{}, fmt.Errorf("rule content is required in file: %s", filePath)
	}

	// Create tags slice if nil
	tags := yamlContent.Tags
	if tags == nil {
		tags = []string{}
	}

	// Create rule
	rule := AgentRule{
		RuleId:      yamlContent.Id,
		Description: yamlContent.Description,
		Language:    yamlContent.Language,
		Tags:        tags,
		Source:      &FileSource{FilePath: filePath},
	}

	return rule, nil
}

// ParseRuleFiles parses all YAML rule files in a directory
func (p *YamlRuleParser) ParseRuleFiles(ctx context.Context, dirPath string) ([]AgentRule, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Continue
	}

	// Validate directory path
	if dirPath == "" {
		return nil, fmt.Errorf("directory path is required")
	}

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dirPath)
	}

	// Get YAML files
	files, err := filepath.Glob(filepath.Join(dirPath, "*.yaml"))
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

		rule, err := p.ParseRule(ctx, file)
		if err != nil {
			// Log error and continue
			fmt.Printf("Error parsing rule file %s: %v\n", file, err)
			continue
		}

		rules = append(rules, rule)
	}

	return rules, nil
}
