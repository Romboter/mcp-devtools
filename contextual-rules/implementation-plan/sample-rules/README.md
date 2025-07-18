# Sample Rules for AgentRulesHub

This directory contains sample rule files for testing and reference during the implementation of the AgentRulesHub tool.

## Rule File Format

Each rule file is in YAML format and contains the following fields:

- `id` (required): A unique identifier for the rule
- `description` (required): A description of the rule's purpose
- `language` (optional): The programming language the rule applies to
- `tags` (optional): A list of tags for categorizing the rule
- `rule` (required): The actual content of the rule

## Sample Files

1. **basic-rule.yaml**: A minimal rule with only the required fields
2. **complex-rule.yaml**: A more complex rule with all available fields, including language and tags
3. **javascript-formatting.yaml**: A rule for JavaScript formatting guidelines

## Usage

These sample files can be used for:

1. **Testing Rule Loading**: Test the rule loader's ability to load rules from YAML files
2. **Testing Rule Parsing**: Test the rule parser's ability to parse rule content
3. **Testing Rule Retrieval**: Test the rule repository's ability to store and retrieve rules
4. **Integration Testing**: Test the complete flow from rule loading to retrieval via MCP tools

## Adding New Sample Rules

To add a new sample rule:

1. Create a new YAML file in this directory
2. Include at least the required fields: `id`, `description`, and `rule`
3. Add any optional fields as needed
4. Update this README.md to include the new sample file

## Example Usage

```go
// Load rules from the sample directory
yamlLoader := NewYamlRuleLoader(NewYamlRuleParser())
rules, err := yamlLoader.LoadRules(ctx, RuleSourceOptions{
    LoaderType: "YamlFile",
    Settings: map[string]interface{}{
        "Path": "contextual-rules/implementation-plan/sample-rules",
    },
})
if err != nil {
    log.Fatalf("Error loading rules: %v", err)
}

// Print loaded rules
for _, rule := range rules {
    fmt.Printf("Rule ID: %s\n", rule.RuleId)
    fmt.Printf("Description: %s\n", rule.Description)
    if rule.Language != "" {
        fmt.Printf("Language: %s\n", rule.Language)
    }
    fmt.Printf("Tags: %v\n", rule.Tags)
    fmt.Println("---")
}
