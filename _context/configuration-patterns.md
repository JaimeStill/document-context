# Configuration Patterns Reference

This document provides the canonical reference for configuration patterns used throughout the document-context project. These patterns implement Layer 1 (Package) transformation boundaries from [Layered Composition Architecture](./_context/lca/layered-composition-architecture.md).

**Core Principle**: Configuration structures are ephemeral data containers that transform into domain objects at package boundaries through finalization, validation, and initialization functions. Configuration is data; domain objects are behavior.

---

## Table of Contents

1. [Configuration Transformation Pattern](#configuration-transformation-pattern)
   - [Type 1: Initialization-Only](#type-1-initialization-only-configuration)
   - [Type 2: Immutable Runtime Settings](#type-2-immutable-runtime-settings)
   - [Type 3: Mutable Runtime Settings](#type-3-mutable-runtime-settings)
2. [Configuration Composition Pattern](#configuration-composition-pattern)
   - [Base Pattern](#base-pattern)
   - [Enhanced Pattern: Embedded Base Configuration](#enhanced-pattern-embedded-base-configuration)
3. [Decision Framework](#decision-framework)
4. [Codebase Examples](#codebase-examples)

---

## Configuration Transformation Pattern

### Core Concept

Configuration serves three distinct purposes, each with different lifecycle and persistence characteristics:

**Type 1: Initialization-Only Configuration**
- Configuration transforms into domain objects and is discarded after initialization
- Domain object stores resolved objects/interfaces, not the configuration
- Configuration exists only during initialization; domain objects persist throughout runtime
- **Example**: Observer name string → resolved observer interface

**Type 2: Immutable Runtime Settings**
- Configuration represents runtime settings that remain constant after initialization
- Domain object stores the configuration directly as settings (no field extraction/duplication)
- Settings persist for the lifetime of the domain object
- Access via `Settings()` method returning the stored settings
- **Example**: ImageConfig stored in ImageMagickRenderer as `settings` field

**Type 3: Mutable Runtime Settings**
- Configuration represents runtime settings that can be adjusted during execution
- Domain object stores configuration privately with controlled mutation
- Setters validate changes and maintain thread safety (mutex protection)
- Access via getter/setter methods with validation
- **Example**: Connection pool max connections, cache size limits

---

## Type 1: Initialization-Only Configuration

### When to Use

- Configuration resolves to objects or interfaces that provide behavior
- The configuration data itself has no runtime value after initialization
- Example: String identifiers that resolve to interface implementations

### Pattern Structure

**Configuration Structure** (Data):
```go
type DomainConfig struct {
    Field1 string `json:"field1,omitempty"`
    Field2 int    `json:"field2,omitempty"`
}

func DefaultDomainConfig() DomainConfig {
    return DomainConfig{
        Field1: "default-value",
        Field2: 100,
    }
}

func (c *DomainConfig) Merge(source *DomainConfig) {
    if source.Field1 != "" {
        c.Field1 = source.Field1
    }
    if source.Field2 > 0 {
        c.Field2 = source.Field2
    }
}

func (c *DomainConfig) Finalize() {
    defaults := DefaultDomainConfig()
    defaults.Merge(c)
    *c = defaults
}
```

**Transformation Function**:
```go
func NewDomainObject(cfg DomainConfig) (Interface, error) {
    cfg.Finalize()  // Apply defaults

    // Validate domain constraints
    if cfg.Field2 < 10 || cfg.Field2 > 1000 {
        return nil, fmt.Errorf("field2 must be 10-1000, got %d", cfg.Field2)
    }

    // Transform to domain object (extract values, discard config)
    return &domainObjectImpl{
        field1: cfg.Field1,
        field2: cfg.Field2,
        // Configuration discarded after extraction
    }, nil
}
```

### Lifecycle

```
1. Load Configuration (JSON, code, etc.)
   ↓
2. Finalize (merge defaults)
   ↓
3. Transform via New*() (validate + create domain object)
   ↓
4. Use Domain Object (config discarded, only extracted values remain)
```

### Codebase Example

**pkg/logger/slogger.go**: LoggerConfig → Logger interface

---

## Type 2: Immutable Runtime Settings

### When to Use

- Configuration represents immutable settings needed throughout object lifetime
- Settings must be accessible for operations like cache key generation
- No runtime mutation is required
- Example: Image rendering settings (format, DPI, quality)

### Pattern Structure

**Configuration Storage** (store entire config directly):
```go
type imagemagickRenderer struct {
    settings config.ImageConfig  // Store entire config as settings
}

func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()

    // Validate configuration values
    if cfg.Quality < 1 || cfg.Quality > 100 {
        return nil, fmt.Errorf("quality must be 1-100, got %d", cfg.Quality)
    }

    // Store configuration as immutable runtime settings
    return &imagemagickRenderer{
        settings: cfg,  // Settings persist for object lifetime
    }, nil
}

// Access settings throughout lifetime
func (r *imagemagickRenderer) Settings() config.ImageConfig {
    return r.settings
}

// Use stored settings in operations
func (r *imagemagickRenderer) Render(input []byte) ([]byte, error) {
    // Access r.settings.Format, r.settings.DPI, etc. as needed
    return r.executeRender(input, r.settings)
}
```

### Why This Matters

Cache key generation requires access to complete rendering parameters (DPI, quality, filters). Without the `Settings()` method, this would be impossible.

### Codebase Example

**pkg/image/imagemagick.go**: ImageConfig stored as `settings` field

---

## Type 3: Mutable Runtime Settings

### When to Use

- Settings must be adjustable during runtime (resize limits, tune parameters)
- Changes require validation to maintain invariants
- Thread safety is required for concurrent access
- Example: Resource pool limits, performance tuning parameters

### Pattern Structure

```go
type ConnectionPool struct {
    mutex    sync.RWMutex
    settings config.PoolConfig  // Private, mutable through validated setters
}

func NewConnectionPool(cfg config.PoolConfig) (*ConnectionPool, error) {
    cfg.Finalize()

    if cfg.MaxConnections < 1 {
        return nil, fmt.Errorf("max connections must be >= 1")
    }

    return &ConnectionPool{
        settings: cfg,
    }, nil
}

// Read access with read lock
func (p *ConnectionPool) MaxConnections() int {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    return p.settings.MaxConnections
}

// Write access with validation and write lock
func (p *ConnectionPool) SetMaxConnections(max int) error {
    if max < 1 {
        return fmt.Errorf("max connections must be >= 1, got %d", max)
    }

    p.mutex.Lock()
    defer p.mutex.Unlock()

    p.settings.MaxConnections = max
    return p.resize()  // Trigger any necessary adjustments
}
```

### Codebase Example

Not yet implemented in document-context. Placeholder for future enhancement.

---

## Configuration Composition Pattern

### When to Use

**Problem**: An interface has multiple implementations with divergent configuration requirements. Some settings are universal (shared across all implementations), while others are implementation-specific.

**Solution**: Base configuration with universal fields + Options map for implementation-specific settings.

### Base Pattern

**Structure**:
```go
// Base configuration - universal settings + Options map
type CacheConfig struct {
    Logger  LoggerConfig   `json:"logger"`   // Common dependency config
    Options map[string]any `json:"options"`  // Implementation-specific
}

// Implementation-specific typed config
type FilesystemCacheConfig struct {
    Directory string
}

// Parse Options map → Typed config
func parseFilesystemConfig(options map[string]any) (*FilesystemCacheConfig, error) {
    dir, ok := options["directory"]
    if !ok {
        return nil, fmt.Errorf("directory option is required")
    }

    directory, ok := dir.(string)
    if !ok {
        return nil, fmt.Errorf("directory option must be a string")
    }

    if directory == "" {
        return nil, fmt.Errorf("directory option cannot be empty")
    }

    return &FilesystemCacheConfig{Directory: directory}, nil
}
```

**Transformation Flow**:
```
BaseConfig (data, JSON-serializable)
    ↓
Parse Options map[string]any → ImplementationConfig (typed struct)
    ↓
Initialize dependencies from embedded configurations
    ↓
Create implementation with dependencies → Concrete type implementing Interface
```

**Factory Pattern**:
```go
func NewFilesystem(c *config.CacheConfig) (Cache, error) {
    // 1. Parse implementation-specific options
    fsConfig, err := parseFilesystemConfig(c.Options)
    if err != nil {
        return nil, err
    }

    // 2. Initialize dependencies from embedded configurations
    log, err := logger.NewSlogger(c.Logger)
    if err != nil {
        return nil, err
    }

    // 3. Validate implementation logic
    absPath, err := filepath.Abs(fsConfig.Directory)
    if err != nil {
        return nil, err
    }

    // 4. Create implementation (satisfies Cache interface)
    return &FilesystemCache{
        directory: absPath,
        logger:    log,
    }, nil
}
```

### Codebase Example

**pkg/cache/filesystem.go**: CacheConfig → FilesystemCache

---

## Enhanced Pattern: Embedded Base Configuration

### When to Use Enhanced Pattern

- Implementation needs access to both base and specific config during operations
- Configuration is Type 2 (Immutable Runtime Settings) - stored for object lifetime
- Interface requires both Settings() and Parameters() methods
- Cleaner encapsulation outweighs slightly increased parsing complexity

### Structure

```go
// Base configuration (universal settings)
type ImageConfig struct {
    Format  string         `json:"format,omitempty"`
    DPI     int            `json:"dpi,omitempty"`
    Quality int            `json:"quality,omitempty"`
    Options map[string]any `json:"options,omitempty"`  // Implementation-specific
}

// Implementation-specific config EMBEDS base config
type ImageMagickConfig struct {
    Config     ImageConfig  // Embedded base configuration
    Background string       // Implementation-specific field
    Brightness *int         // Implementation-specific field
    Contrast   *int         // Implementation-specific field
}

// Renderer stores single complete configuration
type imagemagickRenderer struct {
    settings ImageMagickConfig  // Contains both base + specific
}
```

### Transformation Pattern

```go
func parseImageMagickConfig(cfg config.ImageConfig) (*ImageMagickConfig, error) {
    imCfg := &ImageMagickConfig{
        Config:     cfg,      // Embed base config
        Background: "white",  // Default implementation-specific value
    }

    // Parse Options map into typed fields
    if bg, ok := cfg.Options["background"]; ok {
        bgStr, ok := bg.(string)
        if !ok {
            return nil, fmt.Errorf("background must be a string")
        }
        imCfg.Background = bgStr
    }

    if b, ok := cfg.Options["brightness"]; ok {
        brightness, ok := b.(int)
        if !ok {
            return nil, fmt.Errorf("brightness must be an integer")
        }
        imCfg.Brightness = &brightness
    }

    return imCfg, nil
}

func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()

    // Validate universal settings
    if cfg.Format != "png" && cfg.Format != "jpg" {
        return nil, fmt.Errorf("unsupported format: %s", cfg.Format)
    }

    // Parse Options into ImageMagickConfig (embeds base + adds specific)
    imCfg, err := parseImageMagickConfig(cfg)
    if err != nil {
        return nil, err
    }

    return &imagemagickRenderer{
        settings: *imCfg,  // Single field contains everything
    }, nil
}
```

### Interface Method Implementation

```go
// Settings() returns embedded base config
func (r *imagemagickRenderer) Settings() config.ImageConfig {
    return r.settings.Config  // Access embedded base
}

// Parameters() returns implementation-specific params for cache keys
func (r *imagemagickRenderer) Parameters() []string {
    params := []string{
        fmt.Sprintf("background=%s", r.settings.Background),
    }
    if r.settings.Brightness != nil {
        params = append(params, fmt.Sprintf("brightness=%d", *r.settings.Brightness))
    }
    return params
}

// Render() accesses both base and specific settings
func (r *imagemagickRenderer) Render(input string, page int, output string) error {
    // Access base config: r.settings.Config.DPI
    // Access specific config: r.settings.Background, r.settings.Brightness
    args := r.buildArgs()
    // ...
}
```

### Benefits

**Enhanced Pattern**:
- **Single storage field**: One `settings` field instead of separate `settings` and `options` fields
- **Natural encapsulation**: ImageMagickConfig is the "complete" configuration for ImageMagick
- **Cleaner access**: `r.settings.Config.DPI` for universal, `r.settings.Background` for specific
- **Simple interface methods**: Settings() returns `r.settings.Config`, Parameters() uses `r.settings`

**General Pattern**:
- JSON-serializable configuration supports file-based and API-based config
- Type safety after transformation boundary (map → struct → validation → behavior)
- New implementations don't require base config changes
- Clear error messages during parsing phase
- Implementation configs evolve independently

### Codebase Example

**pkg/image/imagemagick.go**: ImageConfig with embedded ImageMagickConfig

---

## Decision Framework

Use this decision tree to select the appropriate pattern:

### Configuration Type Decision

```
Does configuration resolve to interface/object?
├─ YES → Type 1 (Initialization-Only)
│        Examples: Logger name → Logger interface
│                 Observer name → Observer interface
│
└─ NO → Are settings needed throughout object lifetime?
       ├─ YES → Will settings change after initialization?
       │        ├─ YES → Type 3 (Mutable Runtime Settings)
       │        │        Examples: Pool limits, cache sizes
       │        │
       │        └─ NO → Type 2 (Immutable Runtime Settings)
       │                 Examples: Renderer settings for cache keys
       │
       └─ NO → Type 1 (Initialization-Only)
                Extract values, discard config
```

### Composition Pattern Decision

```
Does interface have multiple implementations?
├─ NO → Use simple configuration struct
│
└─ YES → Do implementations have different config needs?
         ├─ NO → Use simple configuration struct
         │
         └─ YES → Use Configuration Composition Pattern
                  ↓
                  Does implementation need both base + specific config
                  during operations (e.g., for cache keys)?
                  ├─ YES → Use Enhanced Pattern (Embedded Base)
                  │        Examples: Renderer with Settings() + Parameters()
                  │
                  └─ NO → Use Base Pattern
                           Examples: Cache with common logger dependency
```

---

## Codebase Examples

### Type 1: Initialization-Only
- **pkg/logger/slogger.go**: `LoggerConfig` → `Logger` interface
  - Config contains log level, format, output destination
  - Transforms into Logger interface with slog implementation
  - Config discarded after creating logger

### Type 2: Immutable Runtime Settings
- **pkg/image/imagemagick.go**: `ImageConfig` → `Renderer` interface
  - Config contains format, DPI, quality, filters
  - Stored as `settings` field in imagemagickRenderer
  - Accessed via `Settings()` method for cache key generation

### Configuration Composition (Base Pattern)
- **pkg/cache/filesystem.go**: `CacheConfig` → `Cache` interface
  - Base contains Logger config (common dependency)
  - Options map contains implementation-specific directory
  - Parses into FilesystemCacheConfig

### Configuration Composition (Enhanced Pattern)
- **pkg/image/imagemagick.go**: `ImageConfig` → `Renderer` interface
  - ImageMagickConfig embeds ImageConfig
  - Single `settings` field contains both base and specific
  - Settings() returns embedded base, Parameters() returns specific

---

## Configuration Responsibilities

**Configuration (Data)** - All Types:
- Structure definitions with JSON serialization
- Default value creation via `Default*()` functions
- Configuration merging via `Merge()` methods
- Finalization via `Finalize()` method (merges defaults)
- **NO validation** (structural only, like JSON tags)
- **NO business logic**
- **NO domain knowledge**

**Domain Object (Behavior)**:
- Created through transformation function (`New*()`)
- Validated during construction
- Always in valid state after creation
- Encapsulates business logic
- Interacts through interfaces

**Type 1 Specific**:
- Configuration discarded after creation
- Domain object stores extracted values or resolved interfaces

**Type 2 Specific**:
- Configuration stored as `settings` field (no extraction)
- Settings accessible via `Settings()` method
- Settings immutable after construction

**Type 3 Specific**:
- Configuration stored as private `settings` field
- Mutation through validated setter methods
- Thread-safe access with mutex protection

---

## Related Documentation

- [Layered Composition Architecture](./_context/lca/layered-composition-architecture.md) - Complete framework documentation
- [LCA Synopsis](./_context/lca/lca-synopsis.md) - Quick overview
- [CLAUDE.md](../CLAUDE.md) - Project-specific design principles
- [ARCHITECTURE.md](../ARCHITECTURE.md) - Current implementation details
