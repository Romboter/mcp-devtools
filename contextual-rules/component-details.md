# AgentRulesHub Component Details

This document provides detailed information about each component in the AgentRulesHub system, including interfaces, implementations, and responsibilities.

## Table of Contents

1. [Rule Loaders](#rule-loaders)
2. [Rule Parsers](#rule-parsers)
3. [Rule Repository](#rule-repository)
4. [Rule Loader Orchestrator](#rule-loader-orchestrator)
5. [Initialization Service](#initialization-service)
6. [MCP Tools](#mcp-tools)

## Rule Loaders

Rule loaders are responsible for loading rules from different sources. The system is designed to support multiple loader types through a common interface.

### IRuleLoader Interface

```csharp
public interface IRuleLoader
{
    string LoaderType { get; } // Used by the orchestrator to select the correct loader
    bool CanHandle(string loaderType);
    Task<IEnumerable<AgentRule>> LoadRulesAsync(RuleSourceOptions options, CancellationToken cancellationToken = default);
}
```

### YamlRuleLoader Implementation

```csharp
public class YamlRuleLoader : IRuleLoader
{
    private readonly IRuleParser _ruleParser;
    public string LoaderType => "YamlFile";

    public YamlRuleLoader(IRuleParser ruleParser)
    {
        _ruleParser = ruleParser ?? throw new ArgumentNullException(nameof(ruleParser));
    }

    public bool CanHandle(string loaderType)
    {
        return !string.IsNullOrWhiteSpace(loaderType) && loaderType.Equals(LoaderType, StringComparison.OrdinalIgnoreCase);
    }

    public async Task<IEnumerable<AgentRule>> LoadRulesAsync(RuleSourceOptions options, CancellationToken cancellationToken = default)
    {
        if (options == null)
        {
            throw new ArgumentNullException(nameof(options));
        }

        if (!options.Settings.TryGetValue("Path", out var pathObject) || pathObject is not string folderPath || string.IsNullOrWhiteSpace(folderPath))
        {
            throw new ArgumentException("Path setting is missing, not a string, or empty in RuleSourceOptions for YamlFile loader.", nameof(options));
        }

        if (!Directory.Exists(folderPath))
        {
            throw new DirectoryNotFoundException($"Directory not found: {folderPath}");
        }

        var yamlFiles = Directory.GetFiles(folderPath, "*.yaml", SearchOption.AllDirectories);
        var rules = new List<AgentRule>();

        foreach (var file in yamlFiles)
        {
            if (cancellationToken.IsCancellationRequested)
            {
                break;
            }

            try
            {
                var rule = await _ruleParser.ParseRuleAsync(file, cancellationToken);
                rules.Add(rule);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error parsing rule file {file}: {ex.Message}");
                // Continue with next file instead of failing the entire operation
                continue;
            }
        }

        return rules;
    }
}
```

### Key Responsibilities

- Determine if it can handle a specific loader type
- Load rules from a specific source type (e.g., YAML files)
- Parse rule files using the rule parser
- Handle errors gracefully
- Support cancellation

## Rule Parsers

Rule parsers convert rule content from specific formats into AgentRule objects.

### IRuleParser Interface

```csharp
public interface IRuleParser
{
    Task<AgentRule> ParseRuleAsync(string filePath, CancellationToken cancellationToken = default);
}
```

### YamlRuleParser Implementation

```csharp
public class YamlRuleParser : IRuleParser
{
    public async Task<AgentRule> ParseRuleAsync(string filePath, CancellationToken cancellationToken = default)
    {
        var deserializer = new DeserializerBuilder()
            .WithNamingConvention(CamelCaseNamingConvention.Instance)
            .Build();

        using var reader = new StreamReader(filePath);
        var content = await reader.ReadToEndAsync(cancellationToken);
        var yamlContent = deserializer.Deserialize<YamlRuleContent>(content);

        var rule = new AgentRule
        {
            RuleId = yamlContent.Id,
            Description = yamlContent.Description,
            Language = yamlContent.Language,
            Tags = yamlContent.Tags ?? new List<string>(),
            Source = new FileSource { FilePath = filePath }
        };

        return rule;
    }
}
```

### Key Responsibilities

- Parse rule content from a specific format (e.g., YAML)
- Convert parsed content to AgentRule objects
- Set up the rule source for content retrieval
- Handle file I/O operations

## Rule Repository

The rule repository stores and provides access to rule metadata.

### IRuleMetadataIndexRepository Interface

```csharp
public interface IRuleMetadataIndexRepository
{
    Task AddRuleMetadataAsync(AgentRule rule, CancellationToken cancellationToken = default);
    Task AddRulesMetadataAsync(IEnumerable<AgentRule> rules, CancellationToken cancellationToken = default);
    Task<AgentRule?> GetRuleMetadataByIdAsync(string ruleId, CancellationToken cancellationToken = default);
    Task<IEnumerable<AgentRule>> GetAllRulesMetadataAsync(CancellationToken cancellationToken = default);
}
```

### InMemoryRuleRepository Implementation

```csharp
public class InMemoryRuleRepository : IRuleMetadataIndexRepository
{
    private readonly ConcurrentDictionary<string, AgentRule> _rules = new();

    public Task AddRuleMetadataAsync(AgentRule rule, CancellationToken cancellationToken = default)
    {
        if (rule == null)
        {
            throw new ArgumentNullException(nameof(rule));
        }
        if (string.IsNullOrWhiteSpace(rule.RuleId))
        {
            throw new ArgumentException("RuleId cannot be null or whitespace.", nameof(rule));
        }

        _rules[rule.RuleId] = rule;
        return Task.CompletedTask;
    }

    public Task AddRulesMetadataAsync(IEnumerable<AgentRule> rules, CancellationToken cancellationToken = default)
    {
        if (rules == null)
        {
            throw new ArgumentNullException(nameof(rules));
        }

        foreach (var rule in rules)
        {
            if (cancellationToken.IsCancellationRequested)
            {
                break;
            }
            if (rule == null)
            {
                // Log or decide how to handle null rules in a collection
                Console.WriteLine("Encountered a null rule in the collection. Skipping.");
                continue;
            }
            if (string.IsNullOrWhiteSpace(rule.RuleId))
            {
                // Log or decide how to handle rules with invalid RuleId
                Console.WriteLine($"Encountered a rule with invalid RuleId. Description: {rule.Description}. Skipping.");
                continue;
            }
            _rules[rule.RuleId] = rule;
        }
        return Task.CompletedTask;
    }

    public Task<AgentRule?> GetRuleMetadataByIdAsync(string ruleId, CancellationToken cancellationToken = default)
    {
        if (string.IsNullOrWhiteSpace(ruleId))
        {
            // Consider throwing ArgumentException or returning null based on desired contract
            return Task.FromResult<AgentRule?>(null);
        }
        _rules.TryGetValue(ruleId, out var rule);
        return Task.FromResult(rule);
    }

    public Task<IEnumerable<AgentRule>> GetAllRulesMetadataAsync(CancellationToken cancellationToken = default)
    {
        // Returning a snapshot to prevent modification of the internal collection
        return Task.FromResult(_rules.Values.ToList().AsEnumerable());
    }
}
```

### Key Responsibilities

- Store rules in a thread-safe collection
- Provide methods to add, retrieve, and query rules
- Validate rule data
- Return snapshots to prevent modification of internal collection
- Support cancellation

## Rule Loader Orchestrator

The orchestrator coordinates loading rules from multiple sources.

### IRuleLoaderOrchestrator Interface

```csharp
public interface IRuleLoaderOrchestrator
{
    Task<IEnumerable<AgentRule>> LoadRulesAsync(CancellationToken cancellationToken = default);
}
```

### RuleLoaderOrchestrator Implementation

```csharp
public class RuleLoaderOrchestrator : IRuleLoaderOrchestrator
{
    private readonly IEnumerable<IRuleLoader> _loaders;
    private readonly RuleSourcesOptions _ruleSourcesOptions;

    public RuleLoaderOrchestrator(IEnumerable<IRuleLoader> loaders, IOptions<RuleSourcesOptions> ruleSourcesOptions)
    {
        _loaders = loaders ?? throw new ArgumentNullException(nameof(loaders));

        // Get rule source configurations from the bound object
        List<RuleSourceOptions>? ruleSourceOptionsList = ruleSourcesOptions.Value?.Sources ?? new List<RuleSourceOptions>();

        if (!ruleSourceOptionsList.Any())
        {
           throw new InvalidOperationException("No rule sources configured in appsettings.json. Exiting.");
        }
        _ruleSourcesOptions = ruleSourcesOptions.Value!;
    }

    public async Task<IEnumerable<AgentRule>> LoadRulesAsync(CancellationToken cancellationToken = default)
    {
        if (_ruleSourcesOptions == null)
        {
            throw new ArgumentNullException(nameof(_ruleSourcesOptions));
        }

        var allRules = new List<AgentRule>();

        foreach (var options in _ruleSourcesOptions.Sources)
        {
            if (cancellationToken.IsCancellationRequested)
            {
                break;
            }

            var loader = _loaders.FirstOrDefault(l => l.CanHandle(options.LoaderType));

            if (loader == null)
            {
                Console.WriteLine($"No loader found for type: {options.LoaderType}. Skipping.");
                // Or throw an exception, depending on desired behavior
                // throw new InvalidOperationException($"No loader found for type: {options.LoaderType}");
                continue;
            }

            try
            {
                var rulesFromSource = await loader.LoadRulesAsync(options, cancellationToken);
                allRules.AddRange(rulesFromSource);
            }
            catch (Exception ex)
            {
                // Log the exception and continue with other sources
                Console.WriteLine($"Error loading rules from source with type {options.LoaderType}: {ex.Message}");
                // Depending on requirements, might re-throw or handle more gracefully
                continue;
            }
        }

        return allRules;
    }
}
```

### Key Responsibilities

- Iterate through configured rule sources
- Find appropriate loader for each source
- Aggregate rules from all sources
- Handle errors gracefully
- Support cancellation

## Initialization Service

The initialization service loads rules at application startup.

### RuleInitializationService Implementation

```csharp
public class RuleInitializationService : IHostedService
{
    private readonly IRuleLoaderOrchestrator _ruleLoaderOrchestrator;
    private readonly IRuleMetadataIndexRepository _ruleRepository;
    private readonly ILogger<RuleInitializationService> _logger;

    public RuleInitializationService(
        IRuleLoaderOrchestrator ruleLoaderOrchestrator,
        IRuleMetadataIndexRepository ruleRepository,
        ILogger<RuleInitializationService> logger)
    {
        _ruleLoaderOrchestrator = ruleLoaderOrchestrator ?? throw new ArgumentNullException(nameof(ruleLoaderOrchestrator));
        _ruleRepository = ruleRepository ?? throw new ArgumentNullException(nameof(ruleRepository));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
    }

    public async Task StartAsync(CancellationToken cancellationToken)
    {
        _logger.LogInformation("Rule Initialization Service starting.");

        try
        {
            var loadedRules = await _ruleLoaderOrchestrator.LoadRulesAsync();
            if (loadedRules != null && loadedRules.Any())
            {
                await _ruleRepository.AddRulesMetadataAsync(loadedRules);
                _logger.LogInformation($"Successfully loaded {loadedRules.Count()} rules into the repository via background service.");
            }
            else
            {
                _logger.LogInformation("No rules were loaded by the orchestrator.");
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error occurred during rule initialization.");
            // Depending on the application's requirements, you might want to stop the application
            // or handle this more gracefully. For now, we just log.
        }

        _logger.LogInformation("Rule Initialization Service has completed its startup task.");
    }

    public Task StopAsync(CancellationToken cancellationToken)
    {
        _logger.LogInformation("Rule Initialization Service stopping.");
        return Task.CompletedTask;
    }
}
```

### Key Responsibilities

- Implement IHostedService for integration with .NET hosting
- Load rules via the orchestrator during startup
- Add loaded rules to the repository
- Handle and log errors
- Support cancellation

## MCP Tools

MCP tools expose functionality to MCP clients.

### RuleProviderTools Implementation

```csharp
[McpServerToolType]
public static class RuleProviderTools
{
    [McpServerTool, Description("Gets the content of a specific rule by its ID.")]
    public static async Task<string?> GetRuleContentByIdAsync(
        string ruleId,
        IRuleMetadataIndexRepository ruleRepository)
    {
        var rule = await ruleRepository.GetRuleMetadataByIdAsync(ruleId);
        if (rule?.Source != null)
        {
            return await rule.Source.GetRuleContentAsync(default);
        }
        return null;
    }

    [McpServerTool, Description("Gets the metadata for all rules in the index.")]
    public static async Task<IEnumerable<AgentRule>> GetAllRulesMetadataAsync(
        IRuleMetadataIndexRepository ruleRepository)
    {
        return await ruleRepository.GetAllRulesMetadataAsync();
    }
}
```

### Key Responsibilities

- Provide static methods decorated with MCP attributes
- Expose rule retrieval functionality
- Integrate with the rule repository
- Return appropriate data types for MCP clients

## Application Startup

The application startup configures and initializes all components.

### Program.cs Implementation

```csharp
public class Program
{
    public static async Task Main(string[] args)
    {
        var builder = Host.CreateApplicationBuilder(args); // Pass args

        // Configure MCP logging (as per example)
        builder.Logging.ClearProviders(); // Optional: clear existing providers if any
        builder.Logging.AddConsole(consoleLogOptions =>
        {
            consoleLogOptions.LogToStandardErrorThreshold = LogLevel.Trace;
        });

        // Existing AgentRulesHub Configuration and Services
        builder.Configuration.SetBasePath(Directory.GetCurrentDirectory());
        // Ensure appsettings.json is added if not already implicitly by CreateApplicationBuilder
        // builder.Configuration.AddJsonFile("appsettings.json", optional: false, reloadOnChange: true);

        builder.Services.AddOptions<RuleSourcesOptions>()
          .Bind(builder.Configuration.GetSection(RuleSourcesOptions.SectionName))
          .ValidateDataAnnotations()
          .ValidateOnStart();

        builder.Services.AddSingleton<IRuleParser, YamlRuleParser>();
        builder.Services.AddSingleton<IRuleLoader, YamlRuleLoader>();
        builder.Services.AddSingleton<IRuleLoaderOrchestrator, RuleLoaderOrchestrator>();
        builder.Services.AddSingleton<IRuleMetadataIndexRepository, InMemoryRuleRepository>();
        builder.Services.AddHostedService<RuleInitializationService>();

        // Add MCP Server services
        builder.Services
            .AddMcpServer()
            .WithStdioServerTransport()
            .WithToolsFromAssembly(); // This will discover RuleProviderTools

        // Build and run the host
        await builder.Build().RunAsync();
    }
}
```

### Key Responsibilities

- Configure logging
- Configure application settings
- Register services in the dependency injection container
- Configure MCP server
- Build and run the application
