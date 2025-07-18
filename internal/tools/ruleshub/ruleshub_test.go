package ruleshub

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRuleHubTool_Definition(t *testing.T) {
	tool := &RuleHubTool{}
	def := tool.Definition()

	assert.Equal(t, "ruleshub", def.Name)
	assert.Contains(t, def.Description, "contextual rules")
}

func TestYamlRuleParser_ParseRule(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "test-rule.yaml")

	// Write test content to the file
	testContent := `
id: test-rule
description: A test rule
language: go
tags:
  - test
  - example
rule: |
  # Test Rule Content
  
  This is a test rule content.
`
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	assert.NoError(t, err)

	// Create parser and parse the rule
	parser := NewYamlRuleParser()
	rule, err := parser.ParseRule(context.Background(), testFilePath)

	// Verify the results
	assert.NoError(t, err)
	assert.Equal(t, "test-rule", rule.RuleId)
	assert.Equal(t, "A test rule", rule.Description)
	assert.Equal(t, "go", rule.Language)
	assert.Equal(t, []string{"test", "example"}, rule.Tags)
	assert.NotNil(t, rule.Source)

	// Verify the rule content
	content, err := rule.Source.GetRuleContent(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, content, "Test Rule Content")
}

func TestInMemoryRepository(t *testing.T) {
	repo := NewInMemoryRepository()
	ctx := context.Background()

	// Create test rules
	rule1 := AgentRule{
		RuleId:      "rule1",
		Description: "Rule 1",
		Language:    "go",
		Tags:        []string{"test"},
		Source:      &FileSource{FilePath: "rule1.yaml"},
	}

	rule2 := AgentRule{
		RuleId:      "rule2",
		Description: "Rule 2",
		Language:    "python",
		Tags:        []string{"test", "example"},
		Source:      &FileSource{FilePath: "rule2.yaml"},
	}

	// Add rules to repository
	err := repo.AddRuleMetadata(ctx, rule1)
	assert.NoError(t, err)

	err = repo.AddRulesMetadata(ctx, []AgentRule{rule2})
	assert.NoError(t, err)

	// Get rule by ID
	retrievedRule, err := repo.GetRuleMetadataById(ctx, "rule1")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedRule)
	assert.Equal(t, "rule1", retrievedRule.RuleId)
	assert.Equal(t, "Rule 1", retrievedRule.Description)

	// Get all rules
	allRules, err := repo.GetAllRulesMetadata(ctx)
	assert.NoError(t, err)
	assert.Len(t, allRules, 2)
}

func TestRuleHubTool_Execute(t *testing.T) {
	// This is a more complex test that would require setting up
	// a complete environment with test rules. For a placeholder,
	// we'll just test the basic initialization.

	tool := &RuleHubTool{
		repository: NewInMemoryRepository(),
		orchestrator: NewRuleLoaderOrchestrator([]RuleLoader{
			NewYamlRuleLoader(NewYamlRuleParser()),
		}),
		initialized: false,
	}

	// Create a logger for testing
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	// Test initialization
	err := tool.ensureInitialized(context.Background(), logger)

	// Since we don't have any rule sources configured, we expect an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no rule sources configured")
}
