package ruleshub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sammcj/mcp-devtools/internal/registry"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Rule represents a contextual rule with its metadata and content
type Rule struct {
	ID          string   `yaml:"id" json:"ruleId"`
	Description string   `yaml:"description" json:"description"`
	Language    string   `yaml:"language,omitempty" json:"language,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags"`
	Content     string   `yaml:"rule" json:"content"`
	FilePath    string   `json:"-"` // For reference only
}

// RuleHubTool provides methods for managing and retrieving contextual rules
type RuleHubTool struct {
	rulesDir    string
	rules       map[string]*Rule
	mu          sync.RWMutex
	initialized bool
}

// init registers the tool with the registry
func init() {
	registry.Register(&RuleHubTool{
		rules: make(map[string]*Rule),
	})
}

// Definition returns the tool's definition for MCP registration
func (t *RuleHubTool) Definition() mcp.Tool {
	return mcp.NewTool(
		"ruleshub",
		mcp.WithDescription("A tool for managing and providing contextual rules for AI agents"),
		mcp.WithString("action",
			mcp.Description("The action to perform: 'GetRuleContentById' or 'GetAllRulesMetadata'"),
			mcp.Enum("GetRuleContentById", "GetAllRulesMetadata"),
		),
		mcp.WithString("ruleId",
			mcp.Description("The ID of the rule to retrieve (required for GetRuleContentById)"),
		),
	)
}

// Execute executes the tool's logic based on the provided action
func (t *RuleHubTool) Execute(ctx context.Context, logger *logrus.Logger, cache *sync.Map, args map[string]interface{}) (*mcp.CallToolResult, error) {
	if err := t.ensureInitialized(ctx, logger); err != nil {
		return nil, fmt.Errorf("initializing rule hub: %w", err)
	}

	action, ok := args["action"].(string)
	if !ok {
		return nil, errors.New("action parameter is required")
	}

	switch action {
	case "GetRuleContentById":
		return t.getRuleContentById(ctx, args)
	case "GetAllRulesMetadata":
		return t.getAllRulesMetadata(ctx)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// ensureInitialized ensures the tool is initialized by loading rules if needed
func (t *RuleHubTool) ensureInitialized(ctx context.Context, logger *logrus.Logger) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.initialized {
		return nil
	}

	// Determine rules directory
	rulesDir, err := t.getRulesDirectory()
	if err != nil {
		return fmt.Errorf("determining rules directory: %w", err)
	}

	t.rulesDir = rulesDir
	logger.Infof("Using rules directory: %s", t.rulesDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(t.rulesDir, 0755); err != nil {
		return fmt.Errorf("creating rules directory: %w", err)
	}

	// Load rules from directory
	if err := t.loadRulesFromDirectory(ctx, logger); err != nil {
		return fmt.Errorf("loading rules: %w", err)
	}

	t.initialized = true
	logger.Infof("Loaded %d rules from directory", len(t.rules))
	return nil
}

// getRulesDirectory determines the rules directory path from environment or default
func (t *RuleHubTool) getRulesDirectory() (string, error) {
	// Check environment variable first
	if envPath := os.Getenv("RULE_DIRECTORY"); envPath != "" {
		if filepath.IsAbs(envPath) {
			return envPath, nil
		}
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			return "", fmt.Errorf("resolving relative path: %w", err)
		}
		return absPath, nil
	}

	// Use default path in user's home directory
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}

	return filepath.Join(usr.HomeDir, ".mcp-devtools", "rules"), nil
}

// loadRulesFromDirectory loads all YAML rule files from the rules directory
func (t *RuleHubTool) loadRulesFromDirectory(ctx context.Context, logger *logrus.Logger) error {
	// Check if directory exists
	if _, err := os.Stat(t.rulesDir); os.IsNotExist(err) {
		logger.Warnf("Rules directory does not exist: %s", t.rulesDir)
		return nil // Not an error, just no rules to load
	}

	// Find all YAML files
	pattern := filepath.Join(t.rulesDir, "*.yaml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("finding YAML files: %w", err)
	}

	// Also check for .yml files
	ymlPattern := filepath.Join(t.rulesDir, "*.yml")
	ymlFiles, err := filepath.Glob(ymlPattern)
	if err != nil {
		return fmt.Errorf("finding YML files: %w", err)
	}
	files = append(files, ymlFiles...)

	if len(files) == 0 {
		logger.Infof("No YAML rule files found in directory: %s", t.rulesDir)
		return nil
	}

	// Parse each file
	for _, filePath := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rule, err := t.parseRuleFile(filePath)
		if err != nil {
			logger.Warnf("Error parsing rule file %s: %v", filePath, err)
			continue
		}

		t.rules[rule.ID] = rule
		logger.Debugf("Loaded rule: %s from %s", rule.ID, filePath)
	}

	return nil
}

// parseRuleFile parses a single YAML rule file
func (t *RuleHubTool) parseRuleFile(filePath string) (*Rule, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var rule Rule
	if err := yaml.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Validate required fields
	if rule.ID == "" {
		return nil, errors.New("rule ID is required")
	}
	if rule.Description == "" {
		return nil, errors.New("rule description is required")
	}
	if rule.Content == "" {
		return nil, errors.New("rule content is required")
	}

	// Normalize rule ID (remove spaces, convert to lowercase)
	rule.ID = strings.ToLower(strings.ReplaceAll(rule.ID, " ", "-"))

	// Set file path for reference
	rule.FilePath = filePath

	// Initialize tags if nil
	if rule.Tags == nil {
		rule.Tags = []string{}
	}

	return &rule, nil
}

// getRuleContentById retrieves the content of a rule by its ID
func (t *RuleHubTool) getRuleContentById(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	ruleId, ok := args["ruleId"].(string)
	if !ok || ruleId == "" {
		return nil, errors.New("ruleId parameter is required for GetRuleContentById")
	}

	t.mu.RLock()
	rule, exists := t.rules[ruleId]
	t.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleId)
	}

	// Return the complete rule with content
	jsonBytes, err := json.Marshal(rule)
	if err != nil {
		return nil, fmt.Errorf("marshaling rule to JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// getAllRulesMetadata retrieves metadata for all rules in the repository
func (t *RuleHubTool) getAllRulesMetadata(ctx context.Context) (*mcp.CallToolResult, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Create metadata-only version of rules (without content)
	var rulesMetadata []map[string]interface{}
	for _, rule := range t.rules {
		metadata := map[string]interface{}{
			"ruleId":      rule.ID,
			"description": rule.Description,
			"language":    rule.Language,
			"tags":        rule.Tags,
		}
		rulesMetadata = append(rulesMetadata, metadata)
	}

	result := map[string]interface{}{
		"rules": rulesMetadata,
		"count": len(rulesMetadata),
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshaling rules metadata to JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
