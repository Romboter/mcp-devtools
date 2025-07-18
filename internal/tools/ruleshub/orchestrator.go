package ruleshub

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// RuleLoaderOrchestrator defines the interface for coordinating rule loading
type RuleLoaderOrchestrator interface {
	// LoadRules loads rules from all configured sources
	LoadRules(ctx context.Context) ([]AgentRule, error)
}

// DefaultRuleLoaderOrchestrator implements RuleLoaderOrchestrator
type DefaultRuleLoaderOrchestrator struct {
	loaders       []RuleLoader
	sourceOptions RuleSourcesOptions
}

// NewRuleLoaderOrchestrator creates a new DefaultRuleLoaderOrchestrator
func NewRuleLoaderOrchestrator(loaders []RuleLoader) *DefaultRuleLoaderOrchestrator {
	return &DefaultRuleLoaderOrchestrator{
		loaders:       loaders,
		sourceOptions: RuleSourcesOptions{Sources: []RuleSourceOptions{}},
	}
}

// LoadRules loads rules from all configured sources
func (o *DefaultRuleLoaderOrchestrator) LoadRules(ctx context.Context) ([]AgentRule, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Continue
	}

	// Load configuration from environment variables
	if err := o.loadConfigFromEnv(); err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	// Validate configuration
	if len(o.sourceOptions.Sources) == 0 {
		return nil, fmt.Errorf("no rule sources configured")
	}

	// Load rules from all sources
	var allRules []AgentRule
	for _, options := range o.sourceOptions.Sources {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue
		}

		// Find loader for this source
		var loader RuleLoader
		for _, l := range o.loaders {
			if l.CanHandle(options.LoaderType) {
				loader = l
				break
			}
		}

		if loader == nil {
			fmt.Printf("No loader found for type: %s. Skipping.\n", options.LoaderType)
			continue
		}

		// Load rules from this source
		rules, err := loader.LoadRules(ctx, options)
		if err != nil {
			fmt.Printf("Error loading rules from source with type %s: %v\n", options.LoaderType, err)
			continue
		}

		// Add rules to the result
		allRules = append(allRules, rules...)
	}

	return allRules, nil
}

// loadConfigFromEnv loads configuration from environment variables
func (o *DefaultRuleLoaderOrchestrator) loadConfigFromEnv() error {
	// Clear existing sources
	o.sourceOptions.Sources = []RuleSourceOptions{}

	// Find all source indices
	sourceIndices := make(map[int]bool)
	prefix := "RULESHUB_SOURCES_"
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := parts[0]
			key = strings.TrimPrefix(key, prefix)
			indexStr := strings.SplitN(key, "_", 2)[0]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				continue
			}

			sourceIndices[index] = true
		}
	}

	// Load each source
	for index := range sourceIndices {
		loaderTypeKey := fmt.Sprintf("%s%d_LOADERTYPE", prefix, index)
		loaderType := os.Getenv(loaderTypeKey)
		if loaderType == "" {
			return fmt.Errorf("loader type not specified for source %d", index)
		}

		// Create source options
		options := RuleSourceOptions{
			LoaderType: loaderType,
			Settings:   make(map[string]interface{}),
		}

		// Load settings
		settingsPrefix := fmt.Sprintf("%s%d_SETTINGS_", prefix, index)
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, settingsPrefix) {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) != 2 {
					continue
				}

				key := parts[0]
				value := parts[1]
				key = strings.TrimPrefix(key, settingsPrefix)
				options.Settings[key] = value
			}
		}

		// Add source to options
		o.sourceOptions.Sources = append(o.sourceOptions.Sources, options)
	}

	return nil
}
