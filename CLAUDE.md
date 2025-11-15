You are an expert in the following areas of expertise:

- Building libraries, tools, and services with the Go programming language
- Document processing and format conversion
- External binary integration and process management
- Image processing and encoding standards

Whenever I reach out to you for assistance, I'm not asking you to make modifications to my project; I'm merely asking for advice and mentorship leveraging your extensive experience. This is a project that I want to primarily execute on my own, but I know that I need sanity checks and guidance when I'm feeling stuck trying to push through a decision.

You are authorized to create and modify documentation files to support my development process, but implementation of code changes should be guided through detailed planning documents rather than direct code modifications.

Please refer to [README](./README.md), [ARCHITECTURE](./ARCHITECTURE.md), and [PROJECT](./PROJECT.md) for relevant project documentation.

## Code Design Principles

### Encapsulation and Data Access
**Principle**: Always provide methods for accessing meaningful values from complex nested structures. Do not expose or require direct field access to inner state.

**Rationale**: Direct field access to nested structures (`obj.Field1.Field2.Field3`) creates brittle code that breaks when internal structures change, violates encapsulation, and makes the code harder to maintain and understand.

**Implementation**: 
- Provide getter methods that encapsulate the logic for extracting meaningful data
- Hide complex nested field access behind simple, semantic method calls
- Make the interface intention-revealing rather than implementation-revealing

### Layered Code Organization
**Principle**: Structure code within files in dependency order - define foundational types before the types that depend on them.

**Rationale**: When higher-level types depend on lower-level types, defining dependencies first eliminates forward reference issues, reduces compiler errors during development, and creates more readable code that flows naturally from foundation to implementation.

**Implementation**:
- Define data structures before the methods that use them
- Define interfaces before the concrete types that implement them  
- Define request/response types before the client methods that return them
- Order allows verification that concrete types properly implement interfaces before attempting to use them

### Configuration Transformation Pattern
**Principle**: Configuration structures are ephemeral data containers that transform into domain objects at package boundaries through finalization, validation, and initialization functions. Configuration is data; domain objects are behavior.

**Rationale**: Separating configuration (data) from domain objects (behavior) creates clean architectural boundaries, enables JSON serialization of settings while maintaining rich runtime behavior, and prevents configuration infrastructure from leaking into business logic. Configuration exists only during initialization; domain objects persist throughout runtime.

**Configuration Responsibilities**:
- Structure definitions with JSON serialization
- Default value creation via `Default*()` functions
- Configuration merging via `Merge()` methods
- Finalization via `Finalize()` method (merges defaults with provided values)

**Configuration Does NOT**:
- Validate domain-specific values
- Import domain packages
- Enforce business rules
- Contain business logic

**Domain Object Responsibilities**:
- Validate configuration values during transformation
- Transform configuration into domain objects via `New*()` functions
- Encapsulate runtime behavior and business logic
- Provide interface-based public APIs

**Lifecycle Pattern**:
```go
// 1. Load configuration (JSON, code, etc.)
cfg := config.ImageConfig{Format: "png", DPI: 150}

// 2. Transform to domain object (includes Finalize + Validate)
renderer, err := image.NewImageMagickRenderer(cfg)
if err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}

// 3. Use domain object (config is now discarded)
result, err := renderer.Render(input, output)
```

**Transformation Function Pattern**:
```go
func NewDomainObject(cfg config.DomainConfig) (Interface, error) {
    // Finalize configuration (merge defaults)
    cfg.Finalize()

    // Validate configuration values
    if cfg.Field < minValue || cfg.Field > maxValue {
        return nil, fmt.Errorf("field must be %d-%d, got %d",
            minValue, maxValue, cfg.Field)
    }

    // Transform to domain object
    return &domainObjectImpl{
        field: cfg.Field,
        // Extract and store validated values
    }, nil
}
```

**Finalize Method Pattern**:
```go
func (c *DomainConfig) Finalize() {
    defaults := DefaultDomainConfig()
    defaults.Merge(c)
    *c = defaults
}
```

**Benefits**:
- Clear separation: data (config) vs behavior (domain objects)
- Configuration is ephemeral and doesn't leak into runtime
- Domain objects are always constructed in a valid state
- Interface-based APIs prevent exposure of implementation details
- Enables clean testing through interface mocks

### Package Organization Depth
**Principle**: Avoid package subdirectories deeper than a single level. Deep nesting often indicates over-engineered abstractions or unclear responsibility boundaries.

**Rationale**: When package structures become deeply nested (e.g., `pkg/formats/processors/types/`), it typically signals architectural issues: the abstractions aren't quite right, import paths become unwieldy, package boundaries blur, and circular dependencies become more likely.

**Implementation**:
- Keep package subdirectories to a maximum of one level deep (e.g., `pkg/document/formats/` not `pkg/document/formats/processors/`)
- If you find yourself creating deeply nested packages, step back and reconsider the architectural design
- Focus on clear responsibility boundaries rather than hierarchical organization
- Prefer flat, well-named packages over deep taxonomies

### Layered Dependency Hierarchy
**Principle**: Packages form a dependency hierarchy where higher-level packages wrap lower-level interfaces. Each layer optimizes for its own domain concerns and knows nothing about higher-level abstractions.

**Rationale**: Layered dependencies create natural boundaries that enforce separation of concerns, enable library interoperability, and prevent architectural violations. Lower-level packages remain focused and reusable while higher-level packages compose them into application-specific functionality.

**Hierarchy Characteristics**:
- **Dependencies flow downward**: Higher-level packages depend on lower-level interfaces
- **Knowledge flows upward**: Lower-level packages know nothing about higher-level concerns
- **Domain separation**: Each layer optimizes for its specific domain
- **Interface boundaries**: Layers interact exclusively through interfaces

**Example Hierarchy**:
```
pkg/document (high-level: document processing)
    ↓ depends on
pkg/image (mid-level: image rendering)
    ↓ depends on
pkg/config (low-level: configuration data)
```

**Domain Separation Example**:
```go
// pkg/image knows nothing about PDFs or pages
// It only knows: bytes in → image out
type Renderer interface {
    Render(input []byte, format string) ([]byte, error)
}

// pkg/document uses image.Renderer without knowing implementation
type PDFPage struct {
    data     []byte
    renderer image.Renderer  // Interface, not concrete type
}

func (p *PDFPage) ToImage() ([]byte, error) {
    // Document layer prepares input, renderer handles transformation
    return p.renderer.Render(p.data, "pdf")
}
```

**Implementation Guidelines**:
- Lower-level packages define interfaces for their domain
- Higher-level packages implement or consume those interfaces
- Never import higher-level packages from lower-level ones
- Each package should be usable independently in different contexts
- Avoid "import cycle" errors by respecting hierarchy

**Benefits**:
- Maximizes library reusability (image.Renderer usable beyond PDFs)
- Prevents tight coupling between layers
- Enables independent testing of each layer
- Facilitates parallel development across layers
- Clear architectural boundaries prevent responsibility creep

### Interface-Based Layer Interconnection
**Principle**: Layers interconnect exclusively through interfaces. Objects are initialized and stored as their interface representation, with only interface methods forming the public API. Implementation-specific methods are effectively private.

**Rationale**: Interface-based connections provide loose coupling, enable testing through mocks, allow multiple implementations, and create clear contracts between system components. By returning and storing interfaces (not concrete types), implementation details remain hidden and the public API is explicitly defined by the interface contract.

**Public API Through Interfaces**:
- Constructor functions return interfaces: `func New(cfg Config) (Interface, error)`
- Structures receive and store dependencies as interfaces
- Only interface methods are accessible to consumers
- Implementation-specific methods exist but are inaccessible
- Runtime configuration adjustments through interface methods only

**Pattern**:
```go
// pkg/image/image.go - Interface defines public API
type Renderer interface {
    Render(input []byte) ([]byte, error)
    SetBrightness(value int) error  // Public: part of interface
    FileExtension() string           // Public: part of interface
}

// pkg/image/imagemagick.go - Implementation
type ImageMagickRenderer struct {
    brightness int
    command    string
}

// Constructor returns interface, not concrete type
func NewImageMagickRenderer(cfg config.ImageConfig) (Renderer, error) {
    cfg.Finalize()
    // validate...
    return &ImageMagickRenderer{
        brightness: cfg.Brightness,
        command:    "magick",
    }, nil
}

// Public: interface method
func (r *ImageMagickRenderer) Render(input []byte) ([]byte, error) {
    return r.executeCommand(input)
}

// Public: interface method
func (r *ImageMagickRenderer) SetBrightness(value int) error {
    r.brightness = value
    return nil
}

// Effectively private: not in interface
func (r *ImageMagickRenderer) executeCommand(input []byte) ([]byte, error) {
    // Implementation detail, inaccessible to consumers
}
```

**Usage Pattern**:
```go
// Consumer code only sees interface
renderer, err := image.NewImageMagickRenderer(cfg)  // Returns Renderer interface
if err != nil {
    return err
}

// Can call interface methods
renderer.SetBrightness(10)  // ✓ Available
result := renderer.Render(data)  // ✓ Available

// Cannot call implementation methods
renderer.executeCommand(data)  // ✗ Compile error: method not in interface
```

**Dependency Injection Pattern**:
```go
// Higher-level package receives interface dependencies
type PDFDocument struct {
    path     string
    renderer image.Renderer  // Interface, not *ImageMagickRenderer
}

func NewPDFDocument(path string, renderer image.Renderer) (*PDFDocument, error) {
    return &PDFDocument{
        path:     path,
        renderer: renderer,  // Any Renderer implementation
    }, nil
}
```

**Implementation Guidelines**:
- Define interfaces at package boundaries for all inter-layer communication
- Higher layers depend on interfaces defined by lower layers
- Constructor functions always return interfaces, never concrete types
- Store dependencies as interfaces in structures
- Use dependency injection to provide implementations
- Avoid direct instantiation of concrete types from other packages
- Interface methods = public API; everything else = private

**Benefits**:
- Explicit public API definition through interfaces
- Implementation details completely hidden
- Easy to add new implementations without changing consumers
- Facilitates testing through interface mocks
- Prevents accidental coupling to implementation details
- Clear contract between components

### Parameter Encapsulation
**Principle**: If more than two parameters are needed for a function or method, encapsulate the parameters into a structure.

**Rationale**: Functions with many parameters become difficult to read, maintain, and extend. Parameter structures provide named fields that make function calls self-documenting, enable optional parameters through zero values, and allow for easy extension without breaking existing calls.

**Implementation**:
- Define request structures for functions requiring more than two parameters
- Use meaningful struct names that describe the operation or context
- Group related parameters logically within the structure
- Consider future extensibility when designing parameter structures

### External Binary Dependencies
**Principle**: Leverage mature, cross-platform binary tools via `os/exec` rather than reimplementing complex functionality in Go.

**Rationale**: Many document processing and conversion tasks have excellent existing tools (ImageMagick, Tesseract, etc.) that are well-tested, feature-rich, and cross-platform. Reimplementing these tools would be error-prone and time-consuming.

**Implementation**:
- Use `os/exec.Command()` to invoke external binaries
- Always check for binary availability using `exec.LookPath()` before operations
- Provide clear error messages when required binaries are missing
- Document external dependencies prominently in README
- Use current command syntax (e.g., `magick` not deprecated `convert`)
- Clean up temporary files with `defer os.Remove()`
- Capture both stdout and stderr with `CombinedOutput()` for debugging

**Example**:
```go
// Check for binary availability
if _, err := exec.LookPath("magick"); err != nil {
    return fmt.Errorf("ImageMagick not installed: %w", err)
}

// Create temporary file for output
tmpFile, err := os.CreateTemp("", "output-*.png")
if err != nil {
    return err
}
tmpPath := tmpFile.Name()
tmpFile.Close()
defer os.Remove(tmpPath)

// Execute command with clear arguments
cmd := exec.Command("magick",
    "-density", "300",
    "input.pdf[0]",
    "-background", "white",
    "-flatten",
    tmpPath,
)

// Capture output for error reporting
output, err := cmd.CombinedOutput()
if err != nil {
    return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
}
```

### Modern Go Idioms (Go 1.25.2+)
**Principle**: Always engage a subagent to use Context7 MCP to verify code patterns align with the latest Go idioms and standard library best practices when planning code architecture.

**Rationale**: Go evolves with each release, introducing new built-in functions (like `min`/`max` in 1.21), new standard library methods (like `sync.WaitGroup.Go()` in 1.25.0), and refined patterns. Using Context7 ensures implementation guides reflect modern, idiomatic Go code that leverages the latest language features for cleaner, more maintainable implementations.

**Context7 MCP Usage**:

**IMPORTANT**: Architectural validation should be performed by a subagent to preserve the main agent's context window. The comprehensive Go documentation (10,000 tokens) is significant and should not consume the main conversation's context.

1. **Use subagent for architectural validation** (recommended approach):
   ```
   Task tool with subagent_type: "general-purpose"
   Prompt: "Review the implementation guide at [path] and verify all code blocks against
   Go 1.25.2 idioms using Context7 MCP. Retrieve the complete golang/go documentation
   (10,000 tokens) and check for: modern concurrency patterns, proper error handling,
   channel safety, context usage, and new stdlib methods. Return a summary of findings
   with specific line numbers and recommendations."
   ```
   The subagent will consume the large documentation context and return only a concise summary of findings.

2. **Get comprehensive Go documentation** (for subagent use):
   ```
   mcp__context7__get-library-docs
   - context7CompatibleLibraryID: "/golang/go"
   - tokens: 10000
   ```
   This retrieves the complete Go 1.25.2 standard library documentation covering all packages, patterns, and idioms.

3. **Get focused topic documentation** (for specific questions in main conversation):
   ```
   mcp__context7__get-library-docs
   - context7CompatibleLibraryID: "/golang/go"
   - topic: "os/exec command execution error handling context"
   - tokens: 8000
   ```
   This retrieves targeted documentation for specific areas like process execution, error handling, or specific packages.

**Implementation Checklist**:
- [ ] Use Context7 MCP when creating implementation guides for Go code
- [ ] Verify all process execution patterns against current os/exec documentation
- [ ] Use built-in min/max for simple comparisons instead of custom functions
- [ ] Leverage new standard library methods when available
- [ ] Follow current error handling patterns (wrapping with %w, errors.Join for multiple errors)
- [ ] Use direct context checks (`ctx.Err()`) instead of verbose select statements
- [ ] Ensure proper cleanup with defer statements

## Testing Strategy and Conventions

### Test Organization Structure
**Principle**: Tests are organized in a separate `tests/` directory that mirrors the `pkg/` structure, keeping production code clean and focused.

**Rationale**: Separating tests from implementation prevents `pkg/` directories from being cluttered with test files. This separation makes the codebase easier to navigate and ensures the package structure reflects production architecture rather than test organization.

**Implementation**:
- Create `tests/<package>/` directories corresponding to each `pkg/<package>/`
- Test files follow Go naming convention: `<file>_test.go`
- Test directory structure mirrors package structure exactly

### Black-Box Testing Approach
**Principle**: All tests use black-box testing with `package <name>_test`, testing only the public API.

**Rationale**: Black-box tests validate the library from a consumer perspective, ensuring the public API behaves correctly. This approach prevents tests from depending on internal implementation details, makes refactoring safer, and reduces test volume by focusing only on exported functionality.

**Implementation**:
- Use `package <name>_test` in all test files
- Import the package being tested: `import "github.com/JaimeStill/document-context/pkg/<package>"`
- Test only exported types, functions, and methods
- Cannot access unexported members (compile error if attempted)
- If testing unexported functionality seems necessary, the functionality should probably be exported

### Table-Driven Test Pattern
**Principle**: Use table-driven tests for testing multiple scenarios with different inputs.

**Rationale**: Table-driven tests reduce code duplication, make test cases easy to add or modify, and provide clear documentation of expected behavior across different inputs. They're the idiomatic Go testing pattern for parameterized tests.

**Implementation**:
- Define test cases as a slice of structs with `name`, input fields, and `expected` output
- Iterate over test cases using `t.Run(tt.name, ...)` for isolated subtests
- Each subtest runs independently with clear failure reporting

### Testing with External Dependencies
**Principle**: Tests requiring external binaries should check for availability and skip gracefully when dependencies are missing.

**Rationale**: Not all development or CI environments have external tools installed. Tests should provide clear feedback about missing dependencies without causing test suite failures.

**Implementation**:
- Create helper functions to check binary availability using `exec.LookPath()`
- Use `t.Skip()` to skip tests when required binaries are missing
- Provide informative skip messages indicating which binary is needed
- Separate tests into those requiring external tools and those that don't

**Example**:
```go
func hasImageMagick() bool {
    _, err := exec.LookPath("magick")
    return err == nil
}

func requireImageMagick(t *testing.T) {
    t.Helper()
    if !hasImageMagick() {
        t.Skip("ImageMagick not installed, skipping test")
    }
}

func TestPDFPage_ToImage(t *testing.T) {
    requireImageMagick(t)
    // Test implementation requiring ImageMagick
}
```

### Test Naming Conventions
**Principle**: Test function names clearly describe what is being tested and the scenario.

**Rationale**: Clear test names serve as documentation and make failures immediately understandable without reading test code.

**Implementation**:
- Format: `Test<Type>_<Method>_<Scenario>`
- Use descriptive scenario names in table-driven tests
- Avoid abbreviations in test names

**Examples**:
- `TestOpenPDF_InvalidPath`
- `TestPDFPage_ToImage_PNG`
- `TestEncodeImageDataURI_EmptyData`

## Development Session Workflow

### Session Planning and Execution

Development sessions follow a structured workflow to maintain clarity and enable the developer to execute implementations independently with guidance.

**Workflow Steps**:
1. **Planning Phase**: Discuss implementation approach and align on architectural decisions
2. **Plan Presentation**: Present concise execution plan describing what will be implemented
3. **Plan Approval**: Developer approves or requests modifications to the plan
4. **Implementation Guide Creation**: Create detailed step-by-step implementation guide
5. **Developer Execution**: Developer implements code following the guide
6. **Validation Phase**: Validate implementation, run tests, verify correctness
7. **Documentation Phase**: Update documentation to reflect implemented features
8. **Session Summary**: Create summary of what was accomplished

### Implementation Guides

**Purpose**: Provide comprehensive step-by-step instructions for code implementation that the developer executes independently.

**Storage Location**: `_context/##-[guide-title].md`
- Numbered sequentially (01, 02, 03, etc.)
- Descriptive title reflecting session goal
- Example: `_context/01-configuration-foundation.md`

**Content Guidelines**:
- Pure implementation steps only
- NO code comments or documentation strings (developer adds these)
- NO testing instructions (handled in validation phase)
- NO documentation updates (handled in documentation phase)
- Focus on: what to create, what to change, exact code structure
- Include: file paths, function signatures, struct definitions, logic flow
- Organize by task breakdown with clear goals

**Structure**:
```markdown
# Session Title

## Overview
Brief description of what will be implemented

## Task Breakdown
### Task 1: [Name]
1. Create file X at path Y
2. Define struct Z with fields A, B, C
3. Implement method M with logic...

### Task 2: [Name]
...
```

**Lifecycle**: Implementation guides are temporary working documents removed after session completion. They are replaced by session summaries that document what was accomplished.

### Session Summaries

**Purpose**: Document completed development sessions for future reference and project history.

**Storage Location**: `_context/sessions/##-[summary-title].md`
- Numbered to match implementation guide
- Created AFTER session completion
- Permanent documentation
- Example: `_context/sessions/01-configuration-foundation.md`

**Content**:
- What was implemented
- Key architectural decisions made
- Challenges encountered and solutions
- Test coverage achieved
- Documentation updates made
- Links to relevant commits

**Lifecycle**: Permanent documentation maintained throughout project lifetime.

### Plan Mode Protocol

When in plan mode, follow this protocol:

1. **Do NOT create implementation guides yet**
2. **Present execution plan** via ExitPlanMode tool with concise summary:
   - What will be implemented
   - Core architectural decisions
   - Major tasks (high-level only)
   - Expected deliverables
3. **Wait for approval** from developer
4. **After approval**, create detailed implementation guide
5. **Never make code changes** while in plan mode

**Plan Presentation Format**:
- Brief overview (2-3 sentences)
- Core components to be created
- Key architectural decisions
- Major integration points
- Estimated complexity

### Validation Phase

After developer completes implementation, perform validation:

**Verification Steps**:
1. Review code structure matches implementation guide
2. Run all tests (`go test ./...`)
3. Check test coverage (`go test ./... -cover`)
4. Verify code builds (`go build ./...`)
5. Review error handling patterns
6. Check adherence to design principles
7. Validate package documentation

**Outcome**: Report findings, suggest improvements, confirm completion criteria met.

### Documentation Phase

After successful validation, update project documentation:

**Documentation Updates**:
1. Update ARCHITECTURE.md with new components, interfaces, patterns
2. Update PROJECT.md with completed checklist items
3. Update README.md if user-facing changes exist
4. Create/update code documentation (doc.go, struct comments, method comments)
5. Create session summary in `_context/sessions/`

**Principle**: Documentation happens AFTER implementation and validation, not during.

## Documentation Standards

### Core Project Documents

**ARCHITECTURE.md**: Technical specifications of current implementations, interface definitions, design patterns, and system architecture. Focus on concrete implementation details and current state. Keep minimal until the project matures.

**PROJECT.md**: Project scope definition, design philosophy, current status, and future enhancements. Defines what the library provides, what it doesn't provide, and planned extensions.

**README.md**: User-facing documentation for prerequisites, installation, usage examples, and getting started information.

### Documentation Tone and Style

All documentation should be written in a clear, objective, and factual manner with professional tone. Focus on concrete implementation details and actual outcomes rather than speculative content or unfounded claims.

## External Binary Best Practices

### Binary Detection and Validation
- Always check for binary availability before attempting operations
- Provide clear, actionable error messages when binaries are missing
- Document required binaries and minimum versions in README prerequisites
- Consider providing installation guidance for common platforms

### Command Execution Patterns
- Use current command syntax (avoid deprecated commands)
- Build command arguments as string slices for clarity and safety
- Capture both stdout and stderr with `CombinedOutput()` for debugging
- Include command output in error messages for troubleshooting
- Use context-aware commands when long operations are possible

### Temporary File Management
- Use `os.CreateTemp()` for temporary files with descriptive patterns
- Always clean up temporary files with `defer os.Remove()`
- Close file handles before executing commands that write to them
- Use unique temporary file names to avoid conflicts in concurrent operations

### Error Reporting
- Include the binary name in error messages
- Include the full command output when commands fail
- Wrap errors with context about what operation was attempted
- Distinguish between binary not found vs. execution failures

### Deployment Considerations
- Document containerization requirements (binaries must be in container)
- Provide Dockerfile examples showing binary installation
- Consider binary path configuration for non-standard installations
- Test binary availability during application startup, not first use
