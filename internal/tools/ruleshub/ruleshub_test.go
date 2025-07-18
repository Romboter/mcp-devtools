package ruleshub

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleHubTool_Definition(t *testing.T) {
	tool := &RuleHubTool{
		rules: make(map[string]*Rule),
	}
	def := tool.Definition()

	assert.Equal(t, "ruleshub", def.Name)
	assert.Contains(t, def.Description, "contextual rules")
}

func TestRule_ParseRuleFile(t *testing.T) {
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
	require.NoError(t, err)

	// Create tool and parse the rule
	tool := &RuleHubTool{
		rules: make(map[string]*Rule),
	}
	rule, err := tool.parseRuleFile(testFilePath)

	// Verify the results
	require.NoError(t, err)
	assert.Equal(t, "test-rule", rule.ID)
	assert.Equal(t, "A test rule", rule.Description)
	assert.Equal(t, "go", rule.Language)
	assert.Equal(t, []string{"test", "example"}, rule.Tags)
	assert.Contains(t, rule.Content, "Test Rule Content")
	assert.Equal(t, testFilePath, rule.FilePath)
}

func TestRule_ParseRuleFile_InvalidYAML(t *testing.T) {
	// Create a temporary test file with invalid YAML
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "invalid-rule.yaml")

	testContent := `
id: test-rule
description: A test rule
rule: |
  This is invalid YAML
  [unclosed bracket
`
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	require.NoError(t, err)

	tool := &RuleHubTool{
		rules: make(map[string]*Rule),
	}
	_, err = tool.parseRuleFile(testFilePath)
	assert.NoError(t, err) // YAML parsing should still work, content is just text
}

func TestRule_ParseRuleFile_MissingRequiredFields(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name: "missing ID",
			content: `
description: A test rule
rule: Some content
`,
			errMsg: "rule ID is required",
		},
		{
			name: "missing description",
			content: `
id: test-rule
rule: Some content
`,
			errMsg: "rule description is required",
		},
		{
			name: "missing content",
			content: `
id: test-rule
description: A test rule
`,
			errMsg: "rule content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFilePath := filepath.Join(tempDir, tt.name+".yaml")
			err := os.WriteFile(testFilePath, []byte(tt.content), 0644)
			require.NoError(t, err)

			tool := &RuleHubTool{
				rules: make(map[string]*Rule),
			}
			_, err = tool.parseRuleFile(testFilePath)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestRuleHubTool_LoadRulesFromDirectory(t *testing.T) {
	// Create temporary directory with test rules
	tempDir := t.TempDir()

	// Create test rule files
	rule1Content := `
id: rule-1
description: First rule
language: go
tags: [test]
rule: Content of rule 1
`
	rule2Content := `
id: rule-2
description: Second rule
language: python
tags: [test, example]
rule: Content of rule 2
`

	err := os.WriteFile(filepath.Join(tempDir, "rule1.yaml"), []byte(rule1Content), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "rule2.yml"), []byte(rule2Content), 0644)
	require.NoError(t, err)

	// Create tool and load rules
	tool := &RuleHubTool{
		rules:    make(map[string]*Rule),
		rulesDir: tempDir,
	}

	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	err = tool.loadRulesFromDirectory(context.Background(), logger)
	require.NoError(t, err)

	// Verify rules were loaded
	assert.Len(t, tool.rules, 2)
	assert.Contains(t, tool.rules, "rule-1")
	assert.Contains(t, tool.rules, "rule-2")

	rule1 := tool.rules["rule-1"]
	assert.Equal(t, "First rule", rule1.Description)
	assert.Equal(t, "go", rule1.Language)
	assert.Equal(t, []string{"test"}, rule1.Tags)

	rule2 := tool.rules["rule-2"]
	assert.Equal(t, "Second rule", rule2.Description)
	assert.Equal(t, "python", rule2.Language)
	assert.Equal(t, []string{"test", "example"}, rule2.Tags)
}

func TestRuleHubTool_GetRuleContentById(t *testing.T) {
	// Setup tool with test rule
	tool := &RuleHubTool{
		rules:       make(map[string]*Rule),
		initialized: true,
	}

	testRule := &Rule{
		ID:          "test-rule",
		Description: "A test rule",
		Language:    "go",
		Tags:        []string{"test"},
		Content:     "Test rule content",
		FilePath:    "/path/to/test-rule.yaml",
	}
	tool.rules["test-rule"] = testRule

	// Test valid rule ID
	result, err := tool.getRuleContentById(context.Background(), map[string]interface{}{
		"ruleId": "test-rule",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Parse the JSON result
	var parsedRule Rule
	require.True(t, len(result.Content) > 0, "Expected content in result")

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "Expected TextContent")

	err = json.Unmarshal([]byte(textContent.Text), &parsedRule)
	require.NoError(t, err)
	assert.Equal(t, "test-rule", parsedRule.ID)
	assert.Equal(t, "A test rule", parsedRule.Description)
	assert.Equal(t, "Test rule content", parsedRule.Content)

	// Test invalid rule ID
	_, err = tool.getRuleContentById(context.Background(), map[string]interface{}{
		"ruleId": "invalid-rule",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule not found")

	// Test missing rule ID
	_, err = tool.getRuleContentById(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ruleId parameter is required")
}

func TestRuleHubTool_GetAllRulesMetadata(t *testing.T) {
	// Setup tool with test rules
	tool := &RuleHubTool{
		rules:       make(map[string]*Rule),
		initialized: true,
	}

	rule1 := &Rule{
		ID:          "rule-1",
		Description: "First rule",
		Language:    "go",
		Tags:        []string{"test"},
		Content:     "Content 1",
	}
	rule2 := &Rule{
		ID:          "rule-2",
		Description: "Second rule",
		Language:    "python",
		Tags:        []string{"example"},
		Content:     "Content 2",
	}

	tool.rules["rule-1"] = rule1
	tool.rules["rule-2"] = rule2

	// Test retrieving all rules metadata
	result, err := tool.getAllRulesMetadata(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Parse result as JSON
	var response map[string]interface{}
	require.True(t, len(result.Content) > 0, "Expected content in result")

	textContent, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "Expected TextContent")

	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)

	rules, ok := response["rules"].([]interface{})
	require.True(t, ok)
	assert.Len(t, rules, 2)

	count, ok := response["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(2), count)

	// Verify that content is not included in metadata
	for _, ruleInterface := range rules {
		rule, ok := ruleInterface.(map[string]interface{})
		require.True(t, ok)
		_, hasContent := rule["content"]
		assert.False(t, hasContent, "Metadata should not include content")
	}
}

func TestRuleHubTool_Execute(t *testing.T) {
	// Create temporary directory with a test rule
	tempDir := t.TempDir()
	ruleContent := `
id: test-rule
description: A test rule
rule: Test content
`
	err := os.WriteFile(filepath.Join(tempDir, "test.yaml"), []byte(ruleContent), 0644)
	require.NoError(t, err)

	// Set environment variable to use temp directory
	originalEnv := os.Getenv("RULE_DIRECTORY")
	defer func() {
		if originalEnv != "" {
			os.Setenv("RULE_DIRECTORY", originalEnv)
		} else {
			os.Unsetenv("RULE_DIRECTORY")
		}
	}()
	os.Setenv("RULE_DIRECTORY", tempDir)

	// Create tool
	tool := &RuleHubTool{
		rules: make(map[string]*Rule),
	}

	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	// Test GetAllRulesMetadata
	result, err := tool.Execute(context.Background(), logger, nil, map[string]interface{}{
		"action": "GetAllRulesMetadata",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Test GetRuleContentById
	result, err = tool.Execute(context.Background(), logger, nil, map[string]interface{}{
		"action": "GetRuleContentById",
		"ruleId": "test-rule",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Test invalid action
	_, err = tool.Execute(context.Background(), logger, nil, map[string]interface{}{
		"action": "InvalidAction",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action")

	// Test missing action
	_, err = tool.Execute(context.Background(), logger, nil, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action parameter is required")
}

func TestRuleHubTool_GetRulesDirectory(t *testing.T) {
	tool := &RuleHubTool{
		rules: make(map[string]*Rule),
	}

	// Test with environment variable
	originalEnv := os.Getenv("RULE_DIRECTORY")
	defer func() {
		if originalEnv != "" {
			os.Setenv("RULE_DIRECTORY", originalEnv)
		} else {
			os.Unsetenv("RULE_DIRECTORY")
		}
	}()

	testDir := "/tmp/test-rules"
	os.Setenv("RULE_DIRECTORY", testDir)

	dir, err := tool.getRulesDirectory()
	require.NoError(t, err)
	// On Windows, absolute paths get the drive letter prepended
	// Just check that the path ends with the expected directory structure
	assert.True(t, filepath.IsAbs(dir), "Expected absolute path")
	assert.Contains(t, dir, "tmp")
	assert.Contains(t, dir, "test-rules")

	// Test with relative path
	os.Setenv("RULE_DIRECTORY", "relative/path")
	dir, err = tool.getRulesDirectory()
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(dir))

	// Test without environment variable (default)
	os.Unsetenv("RULE_DIRECTORY")
	dir, err = tool.getRulesDirectory()
	require.NoError(t, err)
	// Check for the directory name in a cross-platform way
	assert.Contains(t, dir, ".mcp-devtools")
	assert.Contains(t, dir, "rules")
}
