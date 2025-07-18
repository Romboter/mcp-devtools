package ruleshub

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AgentRule represents a rule with its metadata
type AgentRule struct {
	RuleId      string     `json:"ruleId"`
	Description string     `json:"description"`
	Language    string     `json:"language,omitempty"`
	Tags        []string   `json:"tags"`
	Source      RuleSource `json:"-"`
}

// RuleSource defines the interface for retrieving rule content
type RuleSource interface {
	// SourceType returns the type of the source (e.g., "File")
	SourceType() string

	// GetRuleContent retrieves the content of the rule
	GetRuleContent(ctx context.Context) (string, error)
}

// FileSource implements RuleSource for file-based rules
type FileSource struct {
	FilePath string
}

// SourceType returns the type of the source
func (fs *FileSource) SourceType() string {
	return "File"
}

// GetRuleContent retrieves the content of the rule from a file
func (fs *FileSource) GetRuleContent(ctx context.Context) (string, error) {
	if fs.FilePath == "" {
		return "", errors.New("file path is not set")
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Continue
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

// YamlRuleContent represents the structure of a YAML rule file
type YamlRuleContent struct {
	Id          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Language    string   `yaml:"language,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Rule        string   `yaml:"rule"`
}

// RuleSourceOptions represents configuration for a rule source
type RuleSourceOptions struct {
	LoaderType string                 `json:"loaderType"`
	Settings   map[string]interface{} `json:"settings"`
}

// RuleSourcesOptions represents configuration for all rule sources
type RuleSourcesOptions struct {
	Sources []RuleSourceOptions `json:"sources"`
}
