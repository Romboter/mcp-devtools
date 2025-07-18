package ruleshub

import (
	"context"
	"errors"
	"sync"
)

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

// InMemoryRepository implements RuleRepository using an in-memory map
type InMemoryRepository struct {
	rules map[string]AgentRule
	mu    sync.RWMutex
}

// NewInMemoryRepository creates a new InMemoryRepository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		rules: make(map[string]AgentRule),
	}
}

// AddRuleMetadata adds a single rule to the repository
func (r *InMemoryRepository) AddRuleMetadata(ctx context.Context, rule AgentRule) error {
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
func (r *InMemoryRepository) AddRulesMetadata(ctx context.Context, rules []AgentRule) error {
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
func (r *InMemoryRepository) GetRuleMetadataById(ctx context.Context, ruleId string) (*AgentRule, error) {
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
func (r *InMemoryRepository) GetAllRulesMetadata(ctx context.Context) ([]AgentRule, error) {
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

	rules := make([]AgentRule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}

	return rules, nil
}
