package ruleshub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sammcj/mcp-devtools/internal/registry"
	"github.com/sirupsen/logrus"
)

// RuleHubTool implements the tools.Tool interface and provides methods for managing
// and retrieving contextual rules for AI agents. It uses a repository to store
// rules metadata and an orchestrator to load rules from various sources.
//
// Fields:
// - repository: Interface for accessing and managing rules metadata.
// - orchestrator: Interface for loading rules from different sources.
// - initialized: Indicates whether the tool has been initialized.
// - initMutex: Mutex to ensure thread-safe initialization.
type RuleHubTool struct {
	repository   RuleRepository
	orchestrator RuleLoaderOrchestrator
	initialized  bool
	initMutex    sync.Mutex
}

// init registers the tool with the registry
func init() {
	registry.Register(&RuleHubTool{
		repository: NewInMemoryRepository(),
		orchestrator: NewRuleLoaderOrchestrator([]RuleLoader{
			NewYamlRuleLoader(NewYamlRuleParser()),
		}),
		initialized: false,
	})
}

// Definition returns the tool's definition for MCP registration. It specifies
// the tool's name, description, and the parameters it accepts.
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

// Execute executes the tool's logic based on the provided action. It supports
// the following actions:
// - GetRuleContentById: Retrieves the content of a rule by its ID.
// - GetAllRulesMetadata: Retrieves metadata for all rules.
//
// Parameters:
// - ctx: Context for managing request-scoped values and deadlines.
// - logger: Logger for recording execution details.
// - cache: Cache for storing temporary data.
// - args: Map of arguments for the action.
func (t *RuleHubTool) Execute(ctx context.Context, logger *logrus.Logger, cache *sync.Map, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// Initialize the repository if not already initialized
	if err := t.ensureInitialized(ctx, logger); err != nil {
		return nil, fmt.Errorf("initializing rule repository: %w", err)
	}

	// Parse action parameter
	action, ok := args["action"].(string)
	if !ok {
		return nil, errors.New("action parameter is required")
	}

	// Execute the appropriate action
	switch action {
	case "GetRuleContentById":
		return t.getRuleContentById(ctx, logger, args)
	case "GetAllRulesMetadata":
		return t.getAllRulesMetadata(ctx, logger)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// ensureInitialized ensures the repository is initialized by loading rules
// from all sources and adding them to the repository. It uses a mutex to
// prevent concurrent initialization.
func (t *RuleHubTool) ensureInitialized(ctx context.Context, logger *logrus.Logger) error {
	t.initMutex.Lock()
	defer t.initMutex.Unlock()

	if t.initialized {
		return nil
	}

	logger.Info("Initializing rule repository")

	// Load rules from all sources
	rules, err := t.orchestrator.LoadRules(ctx)
	if err != nil {
		return fmt.Errorf("loading rules: %w", err)
	}

	// Add rules to repository
	if err := t.repository.AddRulesMetadata(ctx, rules); err != nil {
		return fmt.Errorf("adding rules to repository: %w", err)
	}

	logger.Infof("Loaded %d rules into repository", len(rules))
	t.initialized = true
	return nil
}

// getRuleContentById retrieves the content of a rule by its ID. It validates
// the ruleId parameter, fetches the rule metadata from the repository, and
// retrieves the rule content from its source.
func (t *RuleHubTool) getRuleContentById(ctx context.Context, logger *logrus.Logger, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// Parse ruleId parameter
	ruleId, ok := args["ruleId"].(string)
	if !ok || ruleId == "" {
		return nil, errors.New("ruleId parameter is required for GetRuleContentById")
	}

	// Get rule from repository
	rule, err := t.repository.GetRuleMetadataById(ctx, ruleId)
	if err != nil {
		return nil, fmt.Errorf("getting rule metadata: %w", err)
	}

	if rule == nil {
		return nil, fmt.Errorf("rule not found: %s", ruleId)
	}

	// Get rule content
	content, err := rule.Source.GetRuleContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting rule content: %w", err)
	}

	// Return result
	result := map[string]interface{}{
		"ruleId":      rule.RuleId,
		"description": rule.Description,
		"language":    rule.Language,
		"tags":        rule.Tags,
		"content":     content,
	}

	// Convert result to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// getAllRulesMetadata retrieves metadata for all rules in the repository. It
// converts the metadata into a format suitable for returning as a result.
func (t *RuleHubTool) getAllRulesMetadata(ctx context.Context, logger *logrus.Logger) (*mcp.CallToolResult, error) {
	// Get all rules from repository
	rules, err := t.repository.GetAllRulesMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all rules metadata: %w", err)
	}

	// Convert rules to result format
	var result []map[string]interface{}
	for _, rule := range rules {
		result = append(result, map[string]interface{}{
			"ruleId":      rule.RuleId,
			"description": rule.Description,
			"language":    rule.Language,
			"tags":        rule.Tags,
		})
	}

	// Convert result to JSON
	resultMap := map[string]interface{}{
		"rules": result,
	}
	jsonBytes, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
