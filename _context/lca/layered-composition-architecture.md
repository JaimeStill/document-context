# Layered Composition Architecture

## Overview

This document describes **Layered Composition Architecture** (LCA), a comprehensive architectural philosophy for building robust, maintainable software systems through consistent application of boundary-driven design patterns across all composition layers - from individual packages to cloud platforms.

**Core Principle**: At every composition layer, data structures (configurations, specifications, manifests) are ephemeral containers that transform into domain objects (behavior, runtime instances, active systems) at explicit boundaries through validation and initialization. All interactions happen through well-defined interfaces, ensuring clear contracts and hidden implementation details.

**The Composition Stack**: Modern software systems compose through six distinct layers, each with natural boundaries and transformation points:

1. **Package** - Individual code modules with clear public APIs
2. **Library (Module)** - Collections of packages forming cohesive units
3. **Server (Service)** - Network-accessible services with request/response boundaries
4. **Image (Service Host)** - Self-contained runtime environments
5. **Container (Image Runtime)** - Isolated execution contexts
6. **Platform (Container Host)** - Orchestration and management infrastructure

**The Unifying Pattern**: The same architectural pattern applies at each boundary:
```
Configuration (Data) → Transformation → Validation → Domain Object (Behavior)
```

This creates a consistent mental model from package-level object initialization through platform-level infrastructure deployment. By applying the same principles at every layer, we achieve:

- **Safety**: Objects/systems always initialized in valid states
- **Clarity**: Explicit boundaries between data and behavior at every level
- **Consistency**: Same pattern recognition across all architectural decisions
- **Scalability**: Natural extension from libraries to distributed systems

The following sections explore this pattern in detail, starting with foundational concepts and building upward through the complete composition stack.

---

## Table of Contents

### Part I: Foundational Concepts
1. [Foundation: Data vs Behavior](#foundation-data-vs-behavior)
2. [Configuration Transformation Pattern](#configuration-transformation-pattern)
3. [Interface-Based Public APIs](#interface-based-public-apis)
4. [Layered Dependency Hierarchy](#layered-dependency-hierarchy)

### Part II: Composition Layers
5. [Layer 1: Package](#layer-1-package)
6. [Layer 2: Library (Module)](#layer-2-library-module)
7. [Layer 3: Server (Service)](#layer-3-server-service)
8. [Layer 4: Image (Service Host)](#layer-4-image-service-host)
9. [Layer 5: Container (Image Runtime)](#layer-5-container-image-runtime)
10. [Layer 6: Platform (Container Host)](#layer-6-platform-container-host)

### Part III: Unified Patterns
11. [Cross-Layer Pattern Consistency](#cross-layer-pattern-consistency)
12. [Design Principles Summary](#design-principles-summary)

---

## Foundation: Data vs Behavior

### Core Concept

**Data**: Passive structures that hold values, serialize/deserialize, merge, and validate basic constraints.

**Behavior**: Active objects that encapsulate business logic, execute operations, and maintain runtime state.

### The Boundary

The transformation from data to behavior happens at explicit initialization boundaries:

```
Data (Config/Request) → Finalize → Validate → Transform → Behavior (Domain Object)
```

### Why Separate?

**Data Responsibilities**:
- JSON/YAML serialization
- Default value management
- Value merging (overrides, composition)
- Structural validation only

**Behavior Responsibilities**:
- Business logic execution
- Runtime state management
- Complex validation (domain rules)
- Interface-based interaction

**Benefits**:
- Clean separation of concerns
- Configuration doesn't leak into business logic
- Domain objects always constructed in valid state
- Easy testing through interface mocking
- Clear architectural boundaries

---

## Configuration Transformation Pattern

### Configuration Types

Configuration serves three distinct purposes at the package layer, each with different lifecycle and persistence characteristics:

**Type 1: Initialization-Only Configuration**
- Configuration transforms into domain objects and is discarded after initialization
- Domain object stores resolved objects/interfaces, not the configuration itself
- Configuration exists only during initialization; domain objects persist throughout runtime
- Common pattern: String identifiers resolve to interface implementations
- Example: Observer name → observer interface, renderer name → renderer interface

**Type 2: Immutable Runtime Settings**
- Configuration represents runtime settings that remain constant after initialization
- Domain object stores the configuration structure directly as settings (no field extraction/duplication)
- Settings persist for the lifetime of the domain object, immutable after construction
- Access via `Settings()` method returning the stored settings
- Common pattern: Settings needed for operations throughout object lifetime
- Example: ImageConfig stored as `settings` field in ImageMagickRenderer (format, DPI, quality for cache keys)

**Type 3: Mutable Runtime Settings**
- Configuration represents runtime settings that can be adjusted during execution
- Domain object stores configuration privately with controlled mutation through setters
- Setters validate changes and maintain thread safety (typically with mutex protection)
- Access via getter/setter methods with validation at each mutation point
- Common pattern: Operational parameters that need runtime tuning
- Example: Connection pool limits, cache size adjustments, performance tuning parameters

**Decision Framework**:

Use **Type 1** when:
- Configuration resolves to objects or interfaces that provide behavior
- The configuration data itself has no runtime value after initialization
- Example: String identifiers that resolve to interface implementations

Use **Type 2** when:
- Configuration represents immutable settings needed throughout object lifetime
- Settings must be accessible for operations (cache key generation, serialization)
- No runtime mutation is required
- Example: Image rendering settings, format conversion parameters

Use **Type 3** when:
- Settings must be adjustable during runtime (resize limits, tune parameters)
- Changes require validation to maintain system invariants
- Thread safety is required for concurrent access to settings
- Example: Resource pool limits, rate limiting thresholds, cache size bounds

### Pattern Structure (Type 1: Initialization-Only)

```go
// 1. Configuration Structure (Data)
type DomainConfig struct {
    Field1 string `json:"field1,omitempty"`
    Field2 int    `json:"field2,omitempty"`
    Field3 *int   `json:"field3,omitempty"`  // Optional
}

// 2. Default Configuration
func DefaultDomainConfig() DomainConfig {
    return DomainConfig{
        Field1: "default-value",
        Field2: 100,
        Field3: nil,  // Optional, not set by default
    }
}

// 3. Merge Method (Override Pattern)
func (c *DomainConfig) Merge(source *DomainConfig) {
    // Non-empty strings override
    if source.Field1 != "" {
        c.Field1 = source.Field1
    }

    // Non-zero values override
    if source.Field2 > 0 {
        c.Field2 = source.Field2
    }

    // Non-nil pointers override (even if pointing to zero)
    if source.Field3 != nil {
        c.Field3 = source.Field3
    }
}

// 4. Finalize Method (Apply Defaults)
func (c *DomainConfig) Finalize() {
    defaults := DefaultDomainConfig()
    defaults.Merge(c)
    *c = defaults
}

// 5. Transformation Function (Validate + Transform)
func NewDomainObject(cfg DomainConfig) (DomainInterface, error) {
    // Finalize first (apply defaults)
    cfg.Finalize()

    // Validate domain constraints
    if cfg.Field1 != "valid-value" && cfg.Field1 != "other-valid-value" {
        return nil, fmt.Errorf("field1 must be 'valid-value' or 'other-valid-value', got %q", cfg.Field1)
    }

    if cfg.Field2 < 10 || cfg.Field2 > 1000 {
        return nil, fmt.Errorf("field2 must be 10-1000, got %d", cfg.Field2)
    }

    if cfg.Field3 != nil && (*cfg.Field3 < 0 || *cfg.Field3 > 100) {
        return nil, fmt.Errorf("field3 must be 0-100, got %d", *cfg.Field3)
    }

    // Transform to domain object
    return &domainObjectImpl{
        field1: cfg.Field1,
        field2: cfg.Field2,
        field3: cfg.Field3,
    }, nil
}
```

### Type 1 Lifecycle Flow (Initialization-Only)

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Load Configuration                                           │
│    - From JSON file                                             │
│    - From code construction                                     │
│    - From environment variables                                 │
└────────────────────────┬────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│ 2. Finalize (Apply Defaults)                                    │
│    cfg := DomainConfig{Field1: "custom"}                        │
│    cfg.Finalize()  // Field2 gets default, Field3 stays nil     │
└────────────────────────┬────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│ 3. Transform (Validate + Create Domain Object)                  │
│    obj, err := NewDomainObject(cfg)                             │
│    - Calls cfg.Finalize() internally                            │
│    - Validates all fields                                       │
│    - Returns interface, not concrete type                       │
│    - Configuration discarded after extraction                   │
└────────────────────────┬────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│ 4. Use Domain Object (Config Discarded)                         │
│    result, err := obj.Execute()                                 │
│    - Configuration no longer exists                             │
│    - Only domain object with extracted values/interfaces        │
└─────────────────────────────────────────────────────────────────┘
```

### Key Characteristics

**Configuration (Data)** - All Types:
- Serializable structures
- Default value functions
- Merge logic for composition
- Finalize method for default application
- **No validation** (structural only, like JSON tags)
- **No business logic**
- **No domain knowledge**

**Domain Object (Behavior)** - Type 1 (Initialization-Only):
- Created through transformation function
- Validated during construction
- Always in valid state
- Encapsulates business logic
- Interacts through interfaces
- **Configuration discarded after creation**
- Stores extracted values or resolved interfaces

**Domain Object (Behavior)** - Type 2 (Immutable Runtime Settings):
- Created through transformation function with validation
- Stores entire configuration structure as `settings` field (no field extraction)
- Settings accessible via `Settings()` method
- Settings immutable after construction
- Settings used throughout object lifetime
- Example: `renderer.Settings()` returns ImageConfig for cache keys

**Domain Object (Behavior)** - Type 3 (Mutable Runtime Settings):
- Created through transformation function with initial validation
- Stores configuration as private `settings` field (not exposed directly)
- Mutation through validated setter methods
- Thread-safe access with mutex protection
- Maintains invariants at each mutation point
- Example: `pool.SetMaxConnections(n)` validates before updating `settings.MaxConnections`

---

## Configuration Composition Pattern

### When Multiple Implementations Need Different Configuration

**Problem**: An interface has multiple implementations with divergent configuration requirements. Some settings are universal (shared across all implementations), while others are implementation-specific.

**Solution**: Base configuration with universal fields + Options map for implementation-specific settings. Implementation-specific typed config parses and validates the Options map.

### Base Pattern Structure

```go
// Base configuration - universal settings + Options map
type ImageConfig struct {
    Format  string         `json:"format,omitempty"`   // Universal
    DPI     int            `json:"dpi,omitempty"`      // Universal
    Quality int            `json:"quality,omitempty"`  // Universal
    Options map[string]any `json:"options,omitempty"`  // Implementation-specific
}

// Implementation-specific typed configuration
type FilesystemCacheConfig struct {
    Directory string  // Specific to filesystem implementation
}

// Parsing function: Options map → Typed config
func parseFilesystemConfig(options map[string]any) (*FilesystemCacheConfig, error) {
    dir, ok := options["directory"].(string)
    if !ok {
        return nil, fmt.Errorf("directory must be a string")
    }

    return &FilesystemCacheConfig{
        Directory: dir,
    }, nil
}
```

### Enhanced Pattern: Embedded Base Configuration

**Refinement**: Implementation-specific config embeds the base config, creating a single "fully realized" configuration containing both universal and specific settings.

**Structure**:
```go
// Base configuration (universal settings)
type ImageConfig struct {
    Format  string         `json:"format,omitempty"`
    DPI     int            `json:"dpi,omitempty"`
    Quality int            `json:"quality,omitempty"`
    Options map[string]any `json:"options,omitempty"`
}

// Implementation-specific config EMBEDS base
type ImageMagickConfig struct {
    Config     ImageConfig  // Embedded base configuration
    Background string       // ImageMagick-specific
    Brightness *int         // ImageMagick-specific (nil = omit)
    Contrast   *int         // ImageMagick-specific (nil = omit)
}

// Domain object stores single complete configuration
type imagemagickRenderer struct {
    settings ImageMagickConfig  // Contains both base + specific
}
```

### Transformation Flow

```
JSON/File Configuration
    ↓ (unmarshal)
BaseConfig (ImageConfig with Options map)
    ↓ (parse + validate)
Implementation Config (ImageMagickConfig with embedded base)
    ↓ (construct)
Domain Object (imagemagickRenderer with settings)
```

### Implementation Pattern

```go
// Parse function embeds base and adds specific fields
func parseImageMagickConfig(cfg config.ImageConfig) (*ImageMagickConfig, error) {
    imCfg := &ImageMagickConfig{
        Config:     cfg,      // Embed base config
        Background: "white",  // Default implementation-specific
    }

    // Parse Options map into typed fields
    if bg, ok := cfg.Options["background"]; ok {
        bgStr, ok := bg.(string)
        if !ok {
            return nil, fmt.Errorf("background must be a string")
        }
        if bgStr == "" {
            return nil, fmt.Errorf("background cannot be empty")
        }
        imCfg.Background = bgStr
    }

    if b, ok := cfg.Options["brightness"]; ok {
        brightness, ok := b.(int)
        if !ok {
            return nil, fmt.Errorf("brightness must be an integer")
        }
        if brightness < 0 || brightness > 200 {
            return nil, fmt.Errorf("brightness must be 0-200")
        }
        imCfg.Brightness = &brightness
    }

    // Validate and parse other options...

    return imCfg, nil
}

// Constructor creates domain object with complete config
func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()  // Merge defaults

    // Validate universal settings
    if cfg.Format != "png" && cfg.Format != "jpg" {
        return nil, fmt.Errorf("unsupported format: %s", cfg.Format)
    }

    // Parse Options into implementation config (embeds base)
    imCfg, err := parseImageMagickConfig(cfg)
    if err != nil {
        return nil, fmt.Errorf("invalid ImageMagick options: %w", err)
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

// Parameters() returns implementation-specific params
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
    // Access base: r.settings.Config.DPI
    // Access specific: r.settings.Background, r.settings.Brightness
    args := r.buildArgs()
    // ...
}
```

### Key Characteristics

**Base Configuration**:
- Contains universal settings shared across all implementations
- Contains Options map for flexibility
- JSON-serializable for file-based configuration
- Independent of any specific implementation

**Options Map**:
- Flexible container: `map[string]any`
- Accommodates varying implementation requirements
- Validated during parsing (type assertions + business rules)
- Fails fast with clear error messages

**Implementation-Specific Config**:
- Embeds base configuration
- Adds typed fields specific to implementation
- Created by parsing Options map
- Represents "fully realized" configuration

**Domain Object Storage**:
- Stores single `settings` field (implementation-specific config)
- Access base via `settings.Config`
- Access specific via `settings.SpecificField`
- Natural encapsulation of complete configuration

### Benefits

**Enhanced Pattern Benefits**:
- **Single storage field**: One `settings` instead of separate `settings` + `options`
- **Natural encapsulation**: Implementation config is complete configuration
- **Cleaner access patterns**: `r.settings.Config.DPI` for base, `r.settings.Background` for specific
- **Simple interface methods**: `Settings()` returns `r.settings.Config`
- **Clear semantics**: Embedded Config = universal, other fields = specific

**General Pattern Benefits**:
- **JSON-serializable**: File-based and API-based configuration
- **Type safety**: After Options map → typed struct transformation
- **Extensible**: New implementations don't require base config changes
- **Independent evolution**: Implementation configs evolve separately
- **Clear errors**: Type validation during parsing, business validation after
- **Shared dependencies**: Universal fields initialize common dependencies

### When to Use

**Use Configuration Composition Pattern when**:
- Defining interface with multiple implementations
- Implementations have different configuration requirements
- Need JSON-serializable configuration
- Shared dependencies across implementations
- Implementation-specific configuration varies significantly

**Use Enhanced Pattern (Embedded Base) when**:
- Implementation needs both base and specific config during operations
- Configuration is Type 2 (Immutable Runtime Settings)
- Interface requires Settings() and Parameters() methods
- Cleaner encapsulation outweighs slightly increased parsing complexity

### Real-World Examples

**Cache Interface** (Base Pattern):
```go
type CacheConfig struct {
    Logger  LoggerConfig   `json:"logger"`   // Common dependency
    Options map[string]any `json:"options"`  // Implementation-specific
}

type FilesystemCacheConfig struct {
    Directory string
}

func NewFilesystem(c *config.CacheConfig) (Cache, error) {
    // Parse Options → FilesystemCacheConfig
    // Initialize logger from CacheConfig.Logger
    // Create FilesystemCache with parsed config + logger
}
```

**Renderer Interface** (Enhanced Pattern):
```go
type ImageConfig struct {
    Format  string         `json:"format,omitempty"`
    DPI     int            `json:"dpi,omitempty"`
    Options map[string]any `json:"options,omitempty"`
}

type ImageMagickConfig struct {
    Config     ImageConfig  // Embedded base
    Background string
    Brightness *int
}

func NewImageMagickRenderer(cfg ImageConfig) (Renderer, error) {
    // Parse into ImageMagickConfig (embeds base + adds specific)
    // Store single settings field
}
```

---

## Interface-Based Public APIs

### Pattern: Interface as Public Contract

**Principle**: Constructor functions return interfaces. Objects are stored and passed as interfaces. Only interface methods are public; everything else is effectively private.

### Structure

```go
// 1. Define Interface (Public API)
type Renderer interface {
    // Public: Core operations
    Render(input []byte) ([]byte, error)

    // Public: Runtime configuration
    SetBrightness(value int) error
    SetContrast(value int) error

    // Public: Introspection
    FileExtension() string
    MimeType() string
}

// 2. Define Implementation (Private Details)
type imagemagickRenderer struct {
    format     string
    brightness int
    contrast   int
    dpi        int
    command    string
}

// 3. Constructor Returns Interface
func NewImageMagickRenderer(cfg ImageConfig) (Renderer, error) {
    cfg.Finalize()

    // Validate
    if cfg.Format != "png" && cfg.Format != "jpg" {
        return nil, fmt.Errorf("unsupported format: %s", cfg.Format)
    }

    // Transform
    return &imagemagickRenderer{
        format:     cfg.Format,
        brightness: derefInt(cfg.Brightness, 0),
        contrast:   derefInt(cfg.Contrast, 0),
        dpi:        cfg.DPI,
        command:    "magick",
    }, nil
}

// 4. Public Methods (Interface)
func (r *imagemagickRenderer) Render(input []byte) ([]byte, error) {
    args := r.buildArgs(input)
    return r.executeCommand(args)
}

func (r *imagemagickRenderer) SetBrightness(value int) error {
    if value < -100 || value > 100 {
        return fmt.Errorf("brightness must be -100 to +100, got %d", value)
    }
    r.brightness = value
    return nil
}

func (r *imagemagickRenderer) FileExtension() string {
    return r.format
}

// 5. Private Methods (Not in Interface)
func (r *imagemagickRenderer) buildArgs(input []byte) []string {
    // Implementation detail, inaccessible to consumers
    args := []string{"-density", strconv.Itoa(r.dpi)}
    if r.brightness != 0 {
        args = append(args, "-brightness-contrast", fmt.Sprintf("%d,0", r.brightness))
    }
    return args
}

func (r *imagemagickRenderer) executeCommand(args []string) ([]byte, error) {
    // Implementation detail, inaccessible to consumers
    cmd := exec.Command(r.command, args...)
    return cmd.CombinedOutput()
}
```

### Usage Pattern

```go
// Consumer code only sees interface
renderer, err := NewImageMagickRenderer(cfg)
if err != nil {
    return err
}

// Can call interface methods (public API)
renderer.SetBrightness(10)           // ✓ Available
ext := renderer.FileExtension()      // ✓ Available
result, err := renderer.Render(data) // ✓ Available

// Cannot call implementation methods (private)
renderer.buildArgs(data)      // ✗ Compile error: not in interface
renderer.executeCommand(args) // ✗ Compile error: not in interface
```

### Benefits

**Explicit Public API**:
- Interface defines exactly what's public
- No accidental exposure of implementation details
- Clear contract for consumers

**Implementation Freedom**:
- Change internal methods without affecting consumers
- Refactor freely as long as interface satisfied
- Add new implementations without changing consumers

**Testing Simplification**:
- Mock interfaces for testing
- No need to mock concrete types
- Test against contract, not implementation

**Enforcement**:
- Compiler enforces interface boundaries
- Cannot accidentally access private methods
- Type system prevents coupling to implementation

---

## Layered Dependency Hierarchy

### Principle

Packages form a dependency hierarchy where:
- **Higher-level packages** depend on **lower-level interfaces**
- **Lower-level packages** know nothing about **higher-level concerns**
- **Each layer** optimizes for **its own domain**

### Hierarchy Flow

```
Application Layer (Highest)
    ↓ depends on
Domain Layer (Business Logic)
    ↓ depends on
Service Layer (Operations)
    ↓ depends on
Infrastructure Layer (Technical Concerns)
    ↓ depends on
Configuration Layer (Lowest - Data Only)
```

### Example: Document Processing Library

```
pkg/document (High: Document Processing)
    ↓ uses image.Renderer interface
pkg/image (Mid: Image Operations)
    ↓ uses config.ImageConfig data
pkg/config (Low: Configuration Data)
```

### Domain Separation

Each layer only knows about its own domain:

**pkg/config** (Lowest):
- Knows: JSON structures, defaults, merging
- Doesn't know: Images, PDFs, rendering algorithms

**pkg/image** (Middle):
- Knows: Image rendering, format conversion, filters
- Doesn't know: PDFs, documents, pages
- **Input**: Raw bytes in some format
- **Output**: Image bytes

**pkg/document** (Highest):
- Knows: PDF structure, page extraction, document processing
- Doesn't know: How images are rendered (uses image.Renderer interface)
- **Input**: Document files
- **Output**: Uses image.Renderer to convert pages

### Code Example

```go
// pkg/config/image.go (Lowest Layer)
package config

type ImageConfig struct {
    Format     string `json:"format"`
    DPI        int    `json:"dpi"`
    Brightness *int   `json:"brightness,omitempty"`
}

// pkg/image/image.go (Middle Layer)
package image

import "yourproject/pkg/config"

type Renderer interface {
    // Render knows nothing about PDFs
    // Just: bytes in, image out
    Render(input []byte, inputFormat string) ([]byte, error)
}

func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    // Transform config to domain object
}

// pkg/document/pdf.go (Highest Layer)
package document

import "yourproject/pkg/image"

type PDFPage struct {
    data     []byte
    format   string
    renderer image.Renderer  // Interface, not concrete type
}

func (p *PDFPage) ToImage() ([]byte, error) {
    // Document layer prepares input
    // Image layer handles rendering
    // Clean separation of concerns
    return p.renderer.Render(p.data, p.format)
}
```

### Benefits

**Library Reusability**:
- `pkg/image` can be used in any context (PDFs, Word docs, raw images)
- Not tied to PDF processing
- Maximizes code reuse

**Clear Boundaries**:
- Each package has well-defined responsibility
- No circular dependencies
- Prevents import cycles

**Independent Evolution**:
- Change PDF processing without affecting image rendering
- Add new image renderers without changing document layer
- Test each layer independently

**Parallel Development**:
- Different teams can work on different layers
- Clear interfaces enable concurrent development
- Integration points well-defined

---

# Part II: Composition Layers

## Layer 1: Package

### Boundary Definition

The **Package** layer is the foundational composition unit in software systems. At this layer, the boundary separates:
- **Data**: Configuration structures, request parameters, options
- **Behavior**: Domain objects, processors, executors with encapsulated logic

**Upward Boundary**: Exposes public interfaces (types, functions, methods) that higher-level code consumes
**Downward Boundary**: Imports and depends on lower-level packages' interfaces
**Transformation Point**: Constructor functions (`New*()`) that accept configuration and return interface-typed domain objects

### Influence Pattern

Higher-level packages configure this layer by:
1. Constructing configuration structures
2. Calling transformation functions with configuration
3. Receiving interface-typed domain objects
4. Interacting only through public interface methods

This layer influences higher layers by:
1. Defining the interface contract (available operations)
2. Enforcing validation rules during transformation
3. Providing behavior through interface implementation
4. Hiding implementation details behind interface boundaries

### Data vs Behavior at Package Boundary

**Data (Configuration)**:
- Serializable structures (JSON/YAML compatible)
- Default value functions
- Merge semantics for composition
- Finalize methods for default application
- NO validation of domain constraints
- NO business logic
- Short-lived (discarded after transformation)

**Behavior (Domain Objects)**:
- Created through transformation/constructor functions
- Validated during construction
- Always in valid state after creation
- Encapsulates business logic
- Interacts through interfaces
- Long-lived (used throughout application lifetime)

### Transformation Pattern

```
Package Configuration (Data)
    ↓
Finalize() - Apply Defaults
    ↓
New*() Transformation Function - Validate Domain Rules
    ↓
Domain Object (Behavior) returned as Interface
    ↓
Higher layer uses Interface methods
```

### Complete Example: Image Rendering Package

This example demonstrates the package layer pattern using the document-context project's image rendering functionality.

#### Layer 1: Configuration (pkg/config)

```go
package config

type ImageConfig struct {
    Format     string `json:"format,omitempty"`
    Quality    int    `json:"quality,omitempty"`
    DPI        int    `json:"dpi,omitempty"`
    Brightness *int   `json:"brightness,omitempty"`
    Contrast   *int   `json:"contrast,omitempty"`
}

func DefaultImageConfig() ImageConfig {
    return ImageConfig{
        Format:  "png",
        Quality: 0,
        DPI:     300,
    }
}

func (c *ImageConfig) Merge(source *ImageConfig) {
    if source.Format != "" {
        c.Format = source.Format
    }
    if source.Quality > 0 {
        c.Quality = source.Quality
    }
    if source.DPI > 0 {
        c.DPI = source.DPI
    }
    if source.Brightness != nil {
        c.Brightness = source.Brightness
    }
    if source.Contrast != nil {
        c.Contrast = source.Contrast
    }
}

func (c *ImageConfig) Finalize() {
    defaults := DefaultImageConfig()
    defaults.Merge(c)
    *c = defaults
}
```

#### Layer 2: Image Service (pkg/image)

```go
package image

import "yourproject/pkg/config"

// Interface defines public API
type Renderer interface {
    Render(input []byte, inputFormat string) ([]byte, error)
    FileExtension() string
    Settings() config.ImageConfig  // Type 2: Expose settings for operations like cache keys
}

// Implementation - Type 2: Store settings directly
type imagemagickRenderer struct {
    settings config.ImageConfig  // Immutable runtime settings
}

// Constructor returns interface
func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()

    // Validate
    if cfg.Format != "png" && cfg.Format != "jpg" {
        return nil, fmt.Errorf("unsupported format: %s", cfg.Format)
    }

    if cfg.DPI < 72 || cfg.DPI > 600 {
        return nil, fmt.Errorf("DPI must be 72-600, got %d", cfg.DPI)
    }

    if cfg.Quality < 1 || cfg.Quality > 100 {
        return nil, fmt.Errorf("quality must be 1-100, got %d", cfg.Quality)
    }

    if cfg.Brightness != nil && (*cfg.Brightness < -100 || *cfg.Brightness > 100) {
        return nil, fmt.Errorf("brightness must be -100 to +100")
    }

    if cfg.Contrast != nil && (*cfg.Contrast < -100 || *cfg.Contrast > 100) {
        return nil, fmt.Errorf("contrast must be -100 to +100")
    }

    // Store config as settings - Type 2 pattern
    return &imagemagickRenderer{
        settings: cfg,  // No field extraction/duplication
    }, nil
}

// Type 2: Expose settings for operations throughout lifetime
func (r *imagemagickRenderer) Settings() config.ImageConfig {
    return r.settings
}

func (r *imagemagickRenderer) Render(input []byte, inputFormat string) ([]byte, error) {
    args := r.buildArgs(inputFormat)
    return r.execute(input, args)
}

func (r *imagemagickRenderer) FileExtension() string {
    return r.settings.Format  // Access from stored settings
}

// Private methods
func (r *imagemagickRenderer) buildArgs(inputFormat string) []string {
    // Use r.settings.DPI, r.settings.Format, r.settings.Brightness, etc.
}

func (r *imagemagickRenderer) execute(input []byte, args []string) ([]byte, error) {
    // Implementation detail
}
```

**Type 3 Example: Mutable Runtime Settings**

```go
// Type 3: Configuration with mutable runtime settings
type ConnectionPool struct {
    mutex    sync.RWMutex
    settings config.PoolConfig  // Private, mutable through setters
    conns    []*Connection
}

func NewConnectionPool(cfg config.PoolConfig) (*ConnectionPool, error) {
    cfg.Finalize()

    // Validate initial configuration
    if cfg.MaxConnections < 1 || cfg.MaxConnections > 1000 {
        return nil, fmt.Errorf("max connections must be 1-1000")
    }

    if cfg.IdleTimeout < 0 {
        return nil, fmt.Errorf("idle timeout must be >= 0")
    }

    pool := &ConnectionPool{
        settings: cfg,
        conns:    make([]*Connection, 0, cfg.MaxConnections),
    }

    return pool, nil
}

// Read access with read lock
func (p *ConnectionPool) MaxConnections() int {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    return p.settings.MaxConnections
}

// Write access with validation and write lock
func (p *ConnectionPool) SetMaxConnections(max int) error {
    // Validate before acquiring lock
    if max < 1 || max > 1000 {
        return fmt.Errorf("max connections must be 1-1000, got %d", max)
    }

    p.mutex.Lock()
    defer p.mutex.Unlock()

    oldMax := p.settings.MaxConnections
    p.settings.MaxConnections = max

    // Resize pool if needed
    if max < oldMax {
        return p.shrinkPool(max)
    } else if max > oldMax {
        return p.expandPool(max)
    }

    return nil
}

// Private resize methods
func (p *ConnectionPool) shrinkPool(newMax int) error {
    // Close excess connections
}

func (p *ConnectionPool) expandPool(newMax int) error {
    // Prepare capacity for more connections
}
```

#### Layer 3: Document Service (pkg/document)

```go
package document

import "yourproject/pkg/image"

type Document interface {
    PageCount() int
    RenderPage(pageNum int, renderer image.Renderer) ([]byte, error)
}

type pdfDocument struct {
    path  string
    pages int
}

func OpenPDF(path string) (Document, error) {
    // Load PDF metadata
    return &pdfDocument{path: path, pages: 10}, nil
}

func (d *pdfDocument) PageCount() int {
    return d.pages
}

func (d *pdfDocument) RenderPage(pageNum int, renderer image.Renderer) ([]byte, error) {
    // Extract page data (PDF-specific logic)
    pageData := d.extractPageBytes(pageNum)

    // Use renderer (doesn't know it's ImageMagick)
    return renderer.Render(pageData, "pdf")
}

// Private
func (d *pdfDocument) extractPageBytes(pageNum int) []byte {
    // PDF extraction logic
    return []byte{}
}
```

#### Usage

```go
package main

import (
    "yourproject/pkg/config"
    "yourproject/pkg/image"
    "yourproject/pkg/document"
)

func main() {
    // 1. Create configuration
    cfg := config.ImageConfig{
        Format: "jpg",
        DPI:    150,
    }

    // 2. Transform to domain object
    renderer, err := image.NewImageMagickRenderer(cfg)
    if err != nil {
        panic(err)
    }

    // 3. Use domain object
    doc, err := document.OpenPDF("report.pdf")
    if err != nil {
        panic(err)
    }

    // 4. Execute with clean boundaries
    img, err := doc.RenderPage(1, renderer)
    if err != nil {
        panic(err)
    }

    // img contains rendered image
}
```

---

## Layer 2: Library (Module)

### Boundary Definition

The **Library (Module)** layer composes multiple packages into a cohesive, versioned unit. At this layer, the boundary separates:
- **Data**: Module metadata (go.mod), version constraints, dependency specifications
- **Behavior**: The compiled, importable library providing a unified API surface

**Upward Boundary**: Exposes the module's public packages and their exported interfaces to consuming applications
**Downward Boundary**: Declares dependencies on other modules (external libraries)
**Transformation Point**: The module build/compilation process that validates dependencies and produces an importable unit

### Influence Pattern

Higher-level applications configure this layer by:
1. Declaring the module as a dependency (in their go.mod)
2. Specifying version constraints
3. Importing the module's public packages
4. Interacting through exported package interfaces

This layer influences higher layers by:
1. Defining the public API surface (which packages are importable)
2. Enforcing version compatibility
3. Bundling related functionality into logical units
4. Providing semantic versioning guarantees

### Data vs Behavior at Library Boundary

**Data (Module Specification)**:
- `go.mod` file declaring module path and dependencies
- Version tags (semantic versioning)
- Package organization and visibility (internal/ directories)
- Build constraints and tags
- NO runtime behavior
- Specification-only, defining structure

**Behavior (Built Module)**:
- Compiled package code
- Executable functionality
- Runtime capabilities
- Type definitions and interfaces
- Validated dependency graph
- Ready for import and use

### Transformation Pattern

```
Module Specification (go.mod + package structure)
    ↓
go build / go mod tidy - Dependency Resolution & Validation
    ↓
Built Module (importable packages)
    ↓
Application imports and uses packages
```

### Example: document-context Module

The document-context project demonstrates library layer organization:

**Module Specification** (`go.mod`):
```go
module github.com/JaimeStill/document-context

go 1.25.4

require (
    github.com/pdfcpu/pdfcpu v0.8.1
)
```

**Package Structure** (Public API Surface):
```
pkg/
  config/      # Public: Configuration structures
  image/       # Public: Image rendering interfaces
  document/    # Public: Document processing
  encoding/    # Public: Data encoding utilities
```

**Internal Boundaries**:
- All packages under `pkg/` are importable by external applications
- Internal implementation details hidden through lowercase types and unexported functions
- Each package maintains its own transformation pattern (Layer 1)

**Usage by Higher Layers**:
```go
import (
    "github.com/JaimeStill/document-context/pkg/config"
    "github.com/JaimeStill/document-context/pkg/image"
    "github.com/JaimeStill/document-context/pkg/document"
)

// Application code uses the library's public packages
cfg := config.DefaultImageConfig()
renderer, err := image.NewImageMagickRenderer(cfg)
// ... use renderer
```

The module acts as a cohesive unit - the library boundary ensures that:
- Versioning is consistent across all packages
- Dependencies are managed centrally
- The public API is well-defined
- Internal changes don't affect consumers (as long as public interfaces remain stable)

---

## Layer 3: Server (Service)

**Note**: This layer was previously documented as "Service-Level Extension: Query-Command Pattern"

### Boundary Definition

The **Server (Service)** layer exposes functionality over network boundaries. At this layer, the boundary separates:
- **Data**: HTTP requests, Query/Command structures, request parameters
- **Behavior**: Running service instances, request handlers, domain executors

**Upward Boundary**: Exposes HTTP/gRPC endpoints that clients can invoke over the network
**Downward Boundary**: Uses library modules and their packages to implement business logic
**Transformation Point**: Request handlers that parse requests into Query/Command structures, transform them into executors, and return responses

### Influence Pattern

Higher-level clients configure this layer by:
1. Sending HTTP requests with parameters
2. Providing query parameters, request bodies (JSON/protobuf)
3. Specifying authentication and authorization context
4. Invoking specific endpoints

This layer influences higher layers by:
1. Defining the service API contract (endpoints, request/response schemas)
2. Enforcing authentication and authorization
3. Providing service-level guarantees (rate limiting, timeouts)
4. Returning structured responses

### Data vs Behavior at Service Boundary

**Data (Requests - Query/Command Structures)**:
- HTTP request bodies (JSON/XML/protobuf)
- Query parameters, path parameters
- Headers (auth tokens, content-type)
- Request validation rules
- NO business logic execution
- Ephemeral (exist only during request lifecycle)

**Behavior (Service Response - Executors)**:
- Running service instance handling requests
- Domain executors performing business operations
- Database queries, external API calls
- Stateful operations (create, update, delete)
- Response generation
- Long-lived (service runs continuously)

### Transformation Pattern

```
HTTP Request
    ↓
Parse into Query/Command Structure (Data)
    ↓
Validate Request Parameters
    ↓
New*Executor() - Transform into Domain Executor (Behavior)
    ↓
Execute() - Perform Business Logic
    ↓
HTTP Response
```

### Query-Command Pattern Implementation

The service layer implements the transformation pattern through Query (read) and Command (write/transform) structures:

### Query Pattern (Read Operations)

**Query**: Request for data, initializes a retrieval executor.

```go
// 1. Query Structure (Request Data)
type GetDocumentQuery struct {
    DocumentID string `json:"document_id"`
    PageNum    int    `json:"page_num"`
    Format     string `json:"format"`
}

// 2. Query Executor Interface
type DocumentRetriever interface {
    Execute(ctx context.Context) (*DocumentResult, error)
}

// 3. Query Transformation Function
func NewDocumentRetriever(query GetDocumentQuery, deps Dependencies) (DocumentRetriever, error) {
    // Validate query parameters
    if query.DocumentID == "" {
        return nil, fmt.Errorf("document_id required")
    }

    if query.PageNum < 1 {
        return nil, fmt.Errorf("page_num must be >= 1")
    }

    validFormats := map[string]bool{"png": true, "jpg": true, "pdf": true}
    if !validFormats[query.Format] {
        return nil, fmt.Errorf("invalid format: %s", query.Format)
    }

    // Transform to executor
    return &documentRetriever{
        documentID: query.DocumentID,
        pageNum:    query.PageNum,
        format:     query.Format,
        docRepo:    deps.DocumentRepository,
        cache:      deps.Cache,
    }, nil
}

// 4. Implementation
type documentRetriever struct {
    documentID string
    pageNum    int
    format     string
    docRepo    DocumentRepository  // Interface dependency
    cache      Cache                // Interface dependency
}

func (r *documentRetriever) Execute(ctx context.Context) (*DocumentResult, error) {
    // Check cache
    if cached, ok := r.cache.Get(r.cacheKey()); ok {
        return cached, nil
    }

    // Retrieve from repository
    doc, err := r.docRepo.GetByID(ctx, r.documentID)
    if err != nil {
        return nil, fmt.Errorf("document not found: %w", err)
    }

    // Extract page
    page, err := doc.GetPage(r.pageNum)
    if err != nil {
        return nil, fmt.Errorf("page not found: %w", err)
    }

    // Format conversion
    result := &DocumentResult{
        DocumentID: r.documentID,
        PageNum:    r.pageNum,
        Data:       page.GetBytes(r.format),
        Format:     r.format,
    }

    // Cache result
    r.cache.Set(r.cacheKey(), result)

    return result, nil
}

// 5. HTTP Handler (Thin Layer)
func HandleGetDocument(deps Dependencies) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var query GetDocumentQuery
        if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
            http.Error(w, "invalid request", http.StatusBadRequest)
            return
        }

        // Transform query → executor
        retriever, err := NewDocumentRetriever(query, deps)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Execute
        result, err := retriever.Execute(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    }
}
```

### Command Pattern (Write/Transform Operations)

**Command**: Request for transformation, initializes an execution pipeline.

```go
// 1. Command Structure (Request Data)
type RenderPageCommand struct {
    DocumentPath string              `json:"document_path"`
    PageNum      int                 `json:"page_num"`
    ImageConfig  config.ImageConfig  `json:"image_config"`
    CacheKey     string              `json:"cache_key,omitempty"`
}

// 2. Command Executor Interface
type PageRenderExecutor interface {
    Execute(ctx context.Context) (*RenderResult, error)
}

// 3. Command Transformation Function
func NewPageRenderExecutor(cmd RenderPageCommand, deps Dependencies) (PageRenderExecutor, error) {
    // Validate command
    if cmd.DocumentPath == "" {
        return nil, fmt.Errorf("document_path required")
    }

    if cmd.PageNum < 1 {
        return nil, fmt.Errorf("page_num must be >= 1")
    }

    // Transform image config → renderer
    renderer, err := image.NewImageMagickRenderer(cmd.ImageConfig)
    if err != nil {
        return nil, fmt.Errorf("invalid image config: %w", err)
    }

    // Transform to executor
    return &pageRenderExecutor{
        documentPath: cmd.DocumentPath,
        pageNum:      cmd.PageNum,
        cacheKey:     cmd.CacheKey,
        renderer:     renderer,      // Domain object
        docLoader:    deps.DocumentLoader,  // Interface dependency
        cache:        deps.Cache,           // Interface dependency
    }, nil
}

// 4. Implementation
type pageRenderExecutor struct {
    documentPath string
    pageNum      int
    cacheKey     string
    renderer     image.Renderer       // Interface
    docLoader    DocumentLoader       // Interface
    cache        Cache                // Interface
}

func (e *pageRenderExecutor) Execute(ctx context.Context) (*RenderResult, error) {
    // Load document
    doc, err := e.docLoader.Load(ctx, e.documentPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load document: %w", err)
    }
    defer doc.Close()

    // Extract page
    page, err := doc.GetPage(e.pageNum)
    if err != nil {
        return nil, fmt.Errorf("failed to get page: %w", err)
    }

    // Render (transformation)
    imageData, err := e.renderer.Render(page.Bytes(), "pdf")
    if err != nil {
        return nil, fmt.Errorf("failed to render: %w", err)
    }

    result := &RenderResult{
        ImageData:  imageData,
        Format:     e.renderer.FileExtension(),
        PageNum:    e.pageNum,
        CachedAt:   time.Now(),
    }

    // Cache if requested
    if e.cacheKey != "" {
        e.cache.Set(e.cacheKey, result)
    }

    return result, nil
}

// 5. HTTP Handler (Thin Layer)
func HandleRenderPage(deps Dependencies) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var cmd RenderPageCommand
        if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
            http.Error(w, "invalid request", http.StatusBadRequest)
            return
        }

        // Transform command → executor
        executor, err := NewPageRenderExecutor(cmd, deps)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Execute
        result, err := executor.Execute(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    }
}
```

### Service Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│ HTTP Layer (Thin)                                           │
│ - Parse request                                             │
│ - Transform to Query/Command                                │
│ - Execute                                                   │
│ - Serialize response                                        │
└────────────────────────┬────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ Executor Layer (Business Logic)                             │
│ - Validate Query/Command                                    │
│ - Coordinate dependencies                                   │
│ - Execute business operations                               │
│ - Return domain results                                     │
└────────────────────────┬────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ Domain Layer (Core Logic)                                   │
│ - Renderers, Processors, Validators                         │
│ - Interface-based interactions                              │
│ - No HTTP knowledge                                         │
└────────────────────────┬────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ Infrastructure Layer (External Dependencies)                │
│ - Repositories, Caches, File Systems                        │
│ - Interface implementations                                 │
└─────────────────────────────────────────────────────────────┘
```

### Benefits

**Clear Boundaries**:
- HTTP layer knows nothing about business logic
- Business logic knows nothing about HTTP
- Domain layer completely independent

**Consistent Pattern**:
- All operations: Request → Transform → Validate → Execute → Response
- Same pattern from config to commands
- Predictable structure

**Testability**:
- HTTP handlers are trivial (parse, transform, execute)
- Executors testable without HTTP infrastructure
- Domain objects testable in isolation
- Interface mocking for all dependencies

**Maintainability**:
- Each layer has single responsibility
- Changes localized to appropriate layer
- Clear contracts between layers

---

## Layer 4: Image (Service Host)

### Boundary Definition

The **Image (Service Host)** layer packages a service and its runtime environment into a self-contained, deployable unit. At this layer, the boundary separates:
- **Data**: Image build specifications, configuration files, environment setup scripts
- **Behavior**: Built container image containing service binary, runtime, and dependencies

**Upward Boundary**: Exposes the image as a runnable artifact that containers can instantiate
**Downward Boundary**: Uses base images, system packages, and includes service binaries
**Transformation Point**: The image build process that layers filesystem changes and produces an immutable artifact

### Influence Pattern

Higher-level container runtimes configure this layer by:
1. Referencing the image by name/tag
2. Providing runtime environment variables
3. Mounting volumes for persistent data
4. Specifying network and resource constraints

This layer influences higher layers by:
1. Defining the service's runtime environment (OS, libraries, tools)
2. Establishing filesystem layout and entrypoint
3. Specifying default environment variables and configuration
4. Providing runtime dependencies and system utilities

### Data vs Behavior at Image Boundary

**Data (Build Specification)**:
- Dockerfile or build configuration
- Base image references
- File copy operations (COPY, ADD)
- Build-time variables (ARG)
- Environment setup commands (RUN)
- NO runtime execution
- Immutable specification

**Behavior (Built Image)**:
- Layered filesystem with all dependencies
- Executable service binary
- Runtime environment (OS, libraries)
- Default configuration and entrypoint
- Ready to instantiate as container
- Immutable artifact (identified by hash)

### Transformation Pattern

```
Image Build Specification (Dockerfile)
    ↓
Build Process - Layer filesystem, install dependencies
    ↓
Validate Build - Ensure all artifacts present
    ↓
Built Image (immutable artifact)
    ↓
Container runtime instantiates image
```

### Conceptual Example: document-context Service Image

If document-context were deployed as a service, the image would encapsulate:

**Build Specification** (conceptual):
- Base image with Go runtime and ImageMagick
- Service binary built from Layer 3 (Server)
- Configuration file templates
- System dependencies (fonts, libraries)
- Entrypoint script to start service

**Built Image Characteristics**:
- Self-contained unit with all dependencies
- Service binary at known path
- ImageMagick and PDF processing tools installed
- Default configuration embedded
- Versioned and tagged (e.g., `document-context:v1.2.3`)

**Boundary Properties**:
- Image is immutable once built (configuration is data frozen into the image)
- Runtime configuration happens at Layer 5 (Container) through environment variables and volume mounts
- Multiple containers can instantiate from the same image

---

## Layer 5: Container (Image Runtime)

### Boundary Definition

The **Container (Image Runtime)** layer instantiates an image into a running, isolated process. At this layer, the boundary separates:
- **Data**: Container configuration (runtime parameters, environment variables, volume mounts)
- **Behavior**: Running container instance with active process, network, and filesystem

**Upward Boundary**: Exposes the container as a schedulable unit to orchestration platforms
**Downward Boundary**: Uses an image as the template and container runtime for execution
**Transformation Point**: Container creation/start process that applies runtime configuration to image

### Influence Pattern

Higher-level platforms configure this layer by:
1. Specifying which image to run
2. Providing environment variables and secrets
3. Mounting volumes for data persistence
4. Configuring network (ports, DNS, service discovery)
5. Setting resource limits (CPU, memory, I/O)

This layer influences higher layers by:
1. Providing running service instance
2. Exposing network ports for communication
3. Reporting health and readiness status
4. Consuming resources (CPU, memory, storage)
5. Generating logs and metrics

### Data vs Behavior at Container Boundary

**Data (Container Configuration)**:
- Runtime parameters (environment variables, command overrides)
- Volume mount specifications
- Network configuration (port mappings, network attachments)
- Resource constraints (CPU limits, memory limits)
- NO running processes
- Short-lived (exists only during configuration phase)

**Behavior (Running Container)**:
- Active process executing service code
- Network stack handling requests
- Filesystem (image layers + writable layer)
- Resource consumption (CPU, memory, I/O)
- Log output stream
- Health check responses
- Long-lived (runs until stopped/crashes)

### Transformation Pattern

```
Container Configuration (run parameters)
    ↓
Image Pull/Verify - Ensure image available
    ↓
Container Create - Apply configuration to image
    ↓
Container Start - Initialize process, network, filesystem
    ↓
Running Container (active service instance)
```

### Conceptual Example: document-context Service Container

**Container Configuration** (runtime specification):
- Image: `document-context:v1.2.3`
- Environment: `PORT=8080`, `LOG_LEVEL=info`, `CACHE_DIR=/data/cache`
- Volumes: `/data/cache` mounted for persistent caching
- Network: Port 8080 exposed for HTTP requests
- Resources: 2 CPU cores, 4GB memory limit

**Running Container Behavior**:
- Service listening on port 8080
- Processing document rendering requests
- Writing cache data to mounted volume
- Logging to stdout/stderr
- Responding to health checks
- Consuming allocated resources

**Boundary Properties**:
- Single container is one instance of the service
- Multiple containers can run from same image (scaling)
- Each container has isolated filesystem (writable layer)
- Configuration transforms image (data) into running service (behavior)

---

## Layer 6: Platform (Container Host)

### Boundary Definition

The **Platform (Container Host)** layer orchestrates multiple containers into a cohesive, managed system. At this layer, the boundary separates:
- **Data**: Declarative manifests specifying desired state (deployments, services, configuration)
- **Behavior**: Running platform maintaining actual state (pods, services, load balancers)

**Upward Boundary**: Exposes management APIs and service endpoints to operators and external clients
**Downward Boundary**: Manages container lifecycle, scheduling, networking, and storage
**Transformation Point**: Reconciliation loop that continuously transforms desired state into actual state

### Influence Pattern

Higher-level operators configure this layer by:
1. Declaring desired state through manifests
2. Updating deployments with new images/configuration
3. Defining scaling policies and resource requirements
4. Configuring ingress and service exposure

This layer influences higher layers by:
1. Ensuring service availability (health checks, restarts)
2. Providing service discovery and load balancing
3. Managing secrets and configuration distribution
4. Enforcing resource quotas and policies
5. Exposing monitoring and logging infrastructure

### Data vs Behavior at Platform Boundary

**Data (Declarative Manifests)**:
- Desired state specifications (number of replicas, image versions)
- Service definitions (load balancer configuration, ports)
- ConfigMaps and Secrets (externalized configuration)
- Resource requests and limits
- Scaling policies
- NO running containers
- Specification-only

**Behavior (Platform Actual State)**:
- Running pods/containers across cluster
- Active load balancers routing traffic
- Service discovery mechanisms (DNS, IP tables)
- Configuration injected into containers
- Health monitoring and automatic remediation
- Resource scheduling and allocation
- Long-lived (platform continuously reconciles)

### Transformation Pattern

```
Declarative Manifest (desired state)
    ↓
Platform Validation - Check resource availability, quotas
    ↓
Reconciliation Loop - Compare desired vs actual state
    ↓
Container Orchestration - Create/update/delete containers
    ↓
Running Platform State (actual state)
    ↓
Continuous Reconciliation - Maintain desired state
```

### Conceptual Example: document-context Service Platform Deployment

**Declarative Manifest** (desired state specification):
- Deployment: 3 replicas of `document-context:v1.2.3`
- Service: Load balancer exposing port 80 → container port 8080
- ConfigMap: Shared configuration for all instances
- PersistentVolume: Shared cache storage
- Resource requests: 2 CPU, 4GB memory per replica

**Platform Actual State** (running behavior):
- 3 running pods distributed across cluster nodes
- Load balancer routing HTTP requests to healthy pods
- ConfigMap mounted in each pod at `/config`
- PersistentVolume mounted at `/data/cache`
- Platform monitors pod health, restarts failures
- Automatic scale-up/down based on load

**Boundary Properties**:
- Declarative specification (data) continuously reconciled to actual state (behavior)
- Platform autonomously maintains desired state
- Changes to manifests trigger reconciliation
- The transformation never "completes" - continuous reconciliation loop
- This is the highest composition layer where configuration → behavior transformation occurs

---

# Part III: Unified Patterns

## Cross-Layer Pattern Consistency

### The Repeating Pattern

At every composition layer, we observe the same fundamental pattern:

```
Data (Specification/Configuration)
    ↓
Transformation (Validation + Construction)
    ↓
Behavior (Running Entity/Active System)
```

### Layer-by-Layer Manifestation

| Layer | Data | Transformation | Behavior |
|-------|------|----------------|----------|
| **Package** | Configuration struct | `New*()` constructor | Domain object implementing interface |
| **Library** | `go.mod` + package layout | `go build` / compilation | Importable module with public packages |
| **Server** | Query/Command struct | `New*Executor()` | Running executor handling requests |
| **Image** | Dockerfile/build spec | Image build process | Immutable container image |
| **Container** | Runtime configuration | Container create/start | Running container process |
| **Platform** | Declarative manifest | Reconciliation loop | Cluster actual state |

### Configuration Flows Downward

Higher layers configure lower layers by providing specifications:

```
Platform Manifest
    ↓ configures
Container Runtime Parameters
    ↓ instantiates from
Image Artifact
    ↓ contains
Server Binary
    ↓ uses
Library Modules
    ↓ compose
Packages with Interfaces
```

### Interfaces Flow Upward

Lower layers expose capabilities to higher layers through interfaces:

```
Package Interfaces (domain objects)
    ↑ used by
Library Public API
    ↑ imported by
Server Request Handlers
    ↑ packaged in
Image Entrypoint
    ↑ instantiated as
Container Process
    ↑ orchestrated by
Platform Service
```

### Validation at Every Boundary

Each transformation point enforces validation:

- **Package**: Domain rules (format valid, values in range)
- **Library**: Dependency compatibility, import cycles
- **Server**: Request parameters, authentication, authorization
- **Image**: Build success, artifact availability
- **Container**: Resource availability, image existence
- **Platform**: Quota limits, cluster capacity, manifest validity

### Consistency Benefits

By applying the same pattern at every layer:

1. **Predictable Architecture**: Same mental model from package to platform
2. **Clear Failure Points**: Validation happens at transformation boundaries
3. **Explicit Contracts**: Data/behavior separation makes interfaces clear
4. **Safe Initialization**: Entities always created in valid state
5. **Independent Testing**: Each layer testable in isolation
6. **Natural Boundaries**: No confusion about where to place logic
7. **Unified Understanding**: Team members recognize pattern regardless of layer

---

## Complete System Example

### Scenario: Document Rendering Service

A web service that accepts document rendering requests, processes them through a rendering pipeline, and returns images.

### Architecture Overview

```
pkg/config/          # Configuration data structures
pkg/image/           # Image rendering domain
pkg/document/        # Document processing domain
pkg/cache/           # Caching infrastructure
pkg/repository/      # Document storage
internal/query/      # Query executors
internal/command/    # Command executors
internal/http/       # HTTP handlers
```

### Complete Flow

#### 1. Configuration Layer

```go
// pkg/config/image.go
package config

type ImageConfig struct {
    Format     string `json:"format,omitempty"`
    DPI        int    `json:"dpi,omitempty"`
    Brightness *int   `json:"brightness,omitempty"`
}

func DefaultImageConfig() ImageConfig {
    return ImageConfig{Format: "png", DPI: 300}
}

func (c *ImageConfig) Finalize() {
    defaults := DefaultImageConfig()
    defaults.Merge(c)
    *c = defaults
}

func (c *ImageConfig) Merge(source *ImageConfig) {
    // ... merge logic
}
```

#### 2. Domain Layer

```go
// pkg/image/image.go
package image

type Renderer interface {
    Render(input []byte, inputFormat string) ([]byte, error)
}

func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()
    // validate and transform
    return &imagemagickRenderer{ /* ... */ }, nil
}

// pkg/document/document.go
package document

type Processor interface {
    ExtractPage(pageNum int) ([]byte, error)
}

func NewPDFProcessor(path string) (Processor, error) {
    // validate and create
    return &pdfProcessor{path: path}, nil
}
```

#### 3. Infrastructure Layer

```go
// pkg/repository/document.go
package repository

type DocumentRepository interface {
    Get(ctx context.Context, id string) (*Document, error)
}

// pkg/cache/cache.go
package cache

type Cache interface {
    Get(key string) ([]byte, bool)
    Set(key string, value []byte) error
}
```

#### 4. Command Layer

```go
// internal/command/render.go
package command

type RenderPageCommand struct {
    DocumentID  string              `json:"document_id"`
    PageNum     int                 `json:"page_num"`
    ImageConfig config.ImageConfig  `json:"image_config"`
}

type RenderExecutor interface {
    Execute(ctx context.Context) (*RenderResult, error)
}

func NewRenderExecutor(
    cmd RenderPageCommand,
    docRepo repository.DocumentRepository,
    cache cache.Cache,
) (RenderExecutor, error) {
    // Validate command
    if cmd.DocumentID == "" {
        return nil, fmt.Errorf("document_id required")
    }

    // Transform image config → renderer
    renderer, err := image.NewImageMagickRenderer(cmd.ImageConfig)
    if err != nil {
        return nil, err
    }

    // Create executor
    return &renderExecutor{
        documentID: cmd.DocumentID,
        pageNum:    cmd.PageNum,
        renderer:   renderer,
        docRepo:    docRepo,
        cache:      cache,
    }, nil
}

type renderExecutor struct {
    documentID string
    pageNum    int
    renderer   image.Renderer
    docRepo    repository.DocumentRepository
    cache      cache.Cache
}

func (e *renderExecutor) Execute(ctx context.Context) (*RenderResult, error) {
    // Get document
    doc, err := e.docRepo.Get(ctx, e.documentID)
    if err != nil {
        return nil, err
    }

    // Process with document processor
    processor, err := document.NewPDFProcessor(doc.Path)
    if err != nil {
        return nil, err
    }

    pageData, err := processor.ExtractPage(e.pageNum)
    if err != nil {
        return nil, err
    }

    // Render with image renderer
    imageData, err := e.renderer.Render(pageData, "pdf")
    if err != nil {
        return nil, err
    }

    return &RenderResult{
        ImageData: imageData,
        PageNum:   e.pageNum,
    }, nil
}
```

#### 5. HTTP Layer

```go
// internal/http/handlers.go
package http

type Dependencies struct {
    DocRepo repository.DocumentRepository
    Cache   cache.Cache
}

func HandleRenderPage(deps Dependencies) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Parse request
        var cmd command.RenderPageCommand
        if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
            http.Error(w, "invalid request", 400)
            return
        }

        // 2. Transform to executor
        executor, err := command.NewRenderExecutor(cmd, deps.DocRepo, deps.Cache)
        if err != nil {
            http.Error(w, err.Error(), 400)
            return
        }

        // 3. Execute
        result, err := executor.Execute(r.Context())
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // 4. Respond
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    }
}
```

#### 6. Main Application

```go
// cmd/server/main.go
package main

func main() {
    // Initialize infrastructure
    docRepo := repository.NewFileSystemRepository("/var/documents")
    cache := cache.NewMemoryCache()

    // Create dependencies
    deps := http.Dependencies{
        DocRepo: docRepo,
        Cache:   cache,
    }

    // Setup HTTP routes
    mux := http.NewServeMux()
    mux.HandleFunc("/render", http.HandleRenderPage(deps))

    // Start server
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

#### 7. Request Flow

```
1. HTTP Request:
   POST /render
   {
     "document_id": "doc123",
     "page_num": 1,
     "image_config": {
       "format": "png",
       "dpi": 150,
       "brightness": 10
     }
   }

2. Handler parses → RenderPageCommand

3. NewRenderExecutor:
   - Validates command
   - Creates image.Renderer from ImageConfig
   - Returns RenderExecutor interface

4. executor.Execute():
   - Gets document from repository
   - Creates document.Processor
   - Extracts page data
   - Renders with image.Renderer
   - Returns result

5. Handler responds:
   {
     "image_data": "base64...",
     "page_num": 1
   }
```

---

## Design Principles Summary

### 1. Universal Data vs Behavior Separation

At every composition layer, the boundary separates specification (data) from execution (behavior):

**Data (Specifications/Configurations)**:
- Package: Configuration structures with serialization
- Library: Module metadata (go.mod), package layout
- Server: Query/Command structures, HTTP requests
- Image: Build specifications (Dockerfile)
- Container: Runtime parameters, environment variables
- Platform: Declarative manifests (desired state)

**Behavior (Execution/Runtime)**:
- Package: Domain objects implementing interfaces
- Library: Compiled, importable modules
- Server: Running service with request handlers
- Image: Built, immutable container image
- Container: Active process with network/filesystem
- Platform: Cluster actual state with reconciliation

### 2. Universal Transformation Pattern

The same pattern repeats at every layer:

```
Specification (Data) → Transformation (Validation) → Runtime (Behavior)
```

- **Package**: Config → `New*()` → Domain Object
- **Library**: go.mod → `go build` → Importable Module
- **Server**: Request → `New*Executor()` → Running Handler
- **Image**: Dockerfile → Build Process → Image Artifact
- **Container**: Config → Container Start → Running Process
- **Platform**: Manifest → Reconciliation → Actual State

### 3. Additive Composition

Each layer adds an execution context to the layer below:

- **Library** = Package + Module Context (versioning, dependencies)
- **Server** = Library + Entrypoint (`main()` initializing packages)
- **Image** = Server Binary + Runtime Environment (OS, dependencies)
- **Container** = Image + Process Context (isolation, resources)
- **Platform** = Containers + Orchestration (scheduling, networking)

**Key Insight**: A server binary is an executable library - the same packages with a `main()` entrypoint. Each higher layer doesn't replace the lower layer, it *adds context*.

### 4. Bidirectional Flow

**Configuration Flows Downward** (higher layers configure lower):
```
Platform Manifest → Container Config → Image Selection → Server Binary → Library Imports → Package Configs
```

**Interfaces Flow Upward** (lower layers expose capabilities):
```
Package APIs → Library Public Surface → Server Endpoints → Image Entrypoint → Container Process → Platform Service
```

### 5. Validation at Every Boundary

Each transformation point enforces its domain's validation:

- **Package**: Domain rules (value ranges, format compatibility)
- **Library**: Dependency compatibility, import cycle detection
- **Server**: Request parameters, authentication, authorization
- **Image**: Build success, artifact availability
- **Container**: Resource availability, image existence
- **Platform**: Quotas, cluster capacity, manifest validity

Validation failures prevent invalid states from propagating upward.

### 6. Interface-Based Boundaries

Lower layers expose capabilities through interfaces, hiding implementation:

- **Package**: Domain interfaces (Renderer, Processor, etc.)
- **Library**: Public package APIs (exported types/functions)
- **Server**: HTTP/gRPC endpoints (request/response contracts)
- **Image**: Entrypoint and exposed ports
- **Container**: Network ports, health check endpoints
- **Platform**: Service discovery, load balancer endpoints

Interface stability allows lower layers to evolve without breaking higher layers.

### 7. Consistent Mental Model

By applying the same pattern at every layer:

1. **Predictable Architecture**: Same thinking from package to platform
2. **Clear Failure Points**: Validation at transformation boundaries
3. **Explicit Contracts**: Data/behavior separation defines interfaces
4. **Safe Initialization**: Entities always created in valid state
5. **Independent Testing**: Each layer testable in isolation
6. **Natural Boundaries**: No confusion about logic placement
7. **Unified Understanding**: Pattern recognition across all layers
8. **Scalable Complexity**: Same principles from library to distributed systems

---

## Conclusion

**Layered Composition Architecture** provides a unified framework for building software systems at any scale. By recognizing that modern systems compose through six distinct layers - Package, Library, Server, Image, Container, and Platform - and applying the same transformation pattern at each boundary, we achieve:

**Safety Through Validation**:
- Every entity constructed in validated state
- Transformation boundaries enforce domain rules
- Invalid configurations rejected before creating behavior
- Failures localized to specific transformation points
- Type systems and compilers enforce contracts

**Clarity Through Separation**:
- Explicit data vs behavior boundaries at every layer
- Configuration flows downward, interfaces flow upward
- Each layer has well-defined responsibilities
- No ambiguity about where logic belongs
- Predictable patterns from package to platform

**Flexibility Through Interfaces**:
- Lower layers hide implementation behind interfaces
- Higher layers depend on contracts, not concret types
- New implementations added without breaking consumers
- Each layer evolves independently
- Interface stability isolates change impact

**Testability Through Isolation**:
- Each layer testable without requiring higher layers
- Interface mocking enables unit testing
- Package tests don't need servers
- Server tests don't need containers
- Container tests don't need platforms
- Fast, focused test suites at every level

**Scalability Through Consistency**:
- Same mental model from library to cloud platform
- Pattern recognition across architectural decisions
- New team members understand structure at all layers
- Principles apply regardless of system scale
- Complexity managed through layered abstraction

**Additive Composition**:
- Each layer adds execution context to the layer below
- Libraries extend packages with versioning
- Servers add entrypoints to libraries
- Images add runtime environments to servers
- Containers add process isolation to images
- Platforms add orchestration to containers

### Application

This architecture is not prescriptive about specific technologies (Docker, Kubernetes, etc.) but rather describes the **natural boundaries** that emerge in modern software systems. Whether you're:

- Designing a Go package with configuration
- Building a module with multiple packages
- Creating a web service
- Containerizing an application
- Deploying to a platform

The same principles apply: separate data from behavior, validate at transformation boundaries, expose interfaces upward, and maintain consistency across all layers.

By keeping your vision aligned to these naturally emerging architectural patterns, you build systems that are safe, clear, flexible, testable, and scalable - from the smallest package to the largest distributed platform.
