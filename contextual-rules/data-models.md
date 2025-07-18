# AgentRulesHub Data Models

This document provides detailed information about the data models used in the AgentRulesHub system.

## Table of Contents

1. [AgentRule](#agentrule)
2. [RuleSource](#rulesource)
3. [FileSource](#filesource)
4. [YamlRuleContent](#yamlrulecontent)
5. [Configuration Models](#configuration-models)
6. [Rule File Format](#rule-file-format)

## AgentRule

The core model representing a rule with its metadata.

```csharp
public class AgentRule
{
    required public string RuleId { get; set; }
    required public string Description { get; set; }
    public string? Language { get; set; }
    public List<string> Tags { get; set; } = new();
    public RuleSource Source { get; set; } = null!;
}
```

### Properties

- **RuleId**: A unique identifier for the rule
- **Description**: A description of the rule's purpose
- **Language**: The programming language the rule applies to (optional)
- **Tags**: A list of tags for categorizing the rule
- **Source**: The source of the rule content (e.g., file, database)

## RuleSource

An abstract base class for different rule sources.

```csharp
public abstract class RuleSource
{
    public abstract string SourceType { get; }
    public abstract Task<string> GetRuleContentAsync(CancellationToken cancellationToken);
}
```

### Properties

- **SourceType**: The type of the rule source (e.g., "File", "Database")

### Methods

- **GetRuleContentAsync**: Retrieves the content of the rule asynchronously

## FileSource

A concrete implementation of RuleSource for file-based rules.

```csharp
public class FileSource : RuleSource
{
    public override string SourceType { get; } = "File"; 
    public string FilePath { get; set; } = string.Empty;
    
    public override Task<string> GetRuleContentAsync(CancellationToken cancellationToken)
    {
        if (string.IsNullOrEmpty(FilePath))
            throw new InvalidOperationException("FilePath is not set");

        using var fileStream = File.OpenText(FilePath);
        var deserializer = new DeserializerBuilder()
            .WithNamingConvention(CamelCaseNamingConvention.Instance)
            .Build();
        var yamlContent = deserializer.Deserialize<Dictionary<string,object>>(fileStream);
        return Task.FromResult((string)yamlContent["rule"]);
    }
}
```

### Properties

- **SourceType**: Always "File" for this implementation
- **FilePath**: The path to the file containing the rule content

### Methods

- **GetRuleContentAsync**: Reads the rule content from the file, deserializes the YAML, and returns the "rule" property

## YamlRuleContent

A model for deserializing YAML rule content.

```csharp
public class YamlRuleContent
{
    [YamlMember(Alias = "id")]
    public string Id { get; set; } = string.Empty;

    [YamlMember(Alias = "description")]
    public string Description { get; set; } = string.Empty;

    [YamlMember(Alias = "language")]
    public string? Language { get; set; }

    [YamlMember(Alias = "tags")]
    public List<string>? Tags { get; set; }

    [YamlMember(Alias = "rule")]
    public string Rule { get; set; } = string.Empty;
}
```

### Properties

- **Id**: The unique identifier for the rule
- **Description**: A description of the rule's purpose
- **Language**: The programming language the rule applies to (optional)
- **Tags**: A list of tags for categorizing the rule
- **Rule**: The actual content of the rule

## Configuration Models

Models for configuring rule sources.

### RuleSourceOptions

```csharp
public class RuleSourceOptions
{
    public string LoaderType { get; set; } = string.Empty; // e.g., "YamlFile", "Database"
    public Dictionary<string, object> Settings { get; set; } = new(); // Loader-specific settings, e.g., "Path" for FileLoader
}
```

### Properties

- **LoaderType**: The type of loader to use for this source (e.g., "YamlFile")
- **Settings**: A dictionary of settings specific to the loader type

### RuleSourcesOptions

```csharp
public class RuleSourcesOptions
{
    public const string SectionName = "RuleSources";
    required public List<RuleSourceOptions> Sources { get; set; }
}
```

### Properties

- **SectionName**: The name of the configuration section ("RuleSources")
- **Sources**: A list of rule source options

## Rule File Format

Rules are defined in YAML files with a specific structure.

### Example Rule File

```yaml
id: sample001
description: A sample rule for testing purposes.
language: csharp
tags:
  - example
  - test
rule: |
  // This is a sample C# rule content.
  public class SampleRuleClass
  {
      public void Execute()
      {
          Console.WriteLine("Sample rule executed!");
      }
  }
```

### Fields

- **id**: A unique identifier for the rule
- **description**: A description of the rule's purpose
- **language**: The programming language the rule applies to (optional)
- **tags**: A list of tags for categorizing the rule
- **rule**: The actual content of the rule

### More Complex Example

```yaml
id: csharp-standards-rule
description: This rule checks for adherence to C# coding standards, including naming conventions and formatting
language: csharp
tags:
    - coding-standards
    - best-practices
    - naming-conventions
    - formatting
rule: |
    # C# Style and Formatting Guide

    #-------------------------------------------------------------------------------
    # General Formatting
    #-------------------------------------------------------------------------------
    [General]
    # Indentation: Use 4 spaces for indentation. Do not use tabs.
    Indentation: 4 spaces

    # MaxLineLength: While not strictly enforced in the example, aim for readability.
    # Consider a soft limit of 120 characters and a hard limit of 160.
    MaxLineLength: 120

    # FileEncoding: Use UTF-8 for all source files.
    FileEncoding: UTF-8

    # Newlines: Use LF (Unix-style) line endings. (Common in modern cross-platform dev)
    # TrailingWhitespace: Remove trailing whitespace from all lines.

    #-------------------------------------------------------------------------------
    # Namespace and Using Directives
    #-------------------------------------------------------------------------------
    [Namespaces]
    # Style: Use file-scoped namespaces (C# 10+ feature).
    # Example: namespace MyCompany.MyProduct.MyModule;
    Style: File-scoped

    [UsingDirectives]
    # Placement: Place all 'using' directives *after* the file-scoped namespace declaration.
    Placement: AfterNamespaceDeclaration
```

## Data Flow

The following diagram illustrates how data flows through the system:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ YAML File   │────►│ YamlContent │────►│ AgentRule   │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Rule        │◄────┤ Rule        │◄────┤ Repository  │
│ Content     │     │ Source      │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
```

1. YAML files are parsed into YamlRuleContent objects
2. YamlRuleContent objects are converted to AgentRule objects
3. AgentRule objects are stored in the repository
4. Rule content is retrieved from the rule source when needed
