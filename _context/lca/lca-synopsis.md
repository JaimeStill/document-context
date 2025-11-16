# Layered Composition Architecture - Synopsis

## What Is This?

**Layered Composition Architecture** (LCA) is a unified framework for understanding how modern software systems compose from packages to cloud platforms. It reveals that the same fundamental pattern repeats at every layer of the technology stack, providing a consistent mental model for building robust, maintainable systems.

## The Core Insight

Whether you're writing a Go package, building a library, deploying a service, or managing a Kubernetes cluster, you're doing the same thing at different scales:

**Taking a specification (data) and transforming it into runtime behavior**

```
Configuration/Specification → Validation → Runtime Behavior
```

This pattern appears everywhere:
- Package config → validated domain object
- go.mod → compiled library
- HTTP request → running handler
- Dockerfile → container image
- Runtime params → active process
- K8s manifest → cluster state

## The Six Composition Layers

Modern systems compose through six distinct layers:

| Layer | You Know This As... | Data (Input) | Behavior (Output) |
|-------|---------------------|--------------|-------------------|
| **1. Package** | Import statements, interfaces | Config structs | Domain objects |
| **2. Library** | go.mod, NPM packages | Module metadata | Compiled code |
| **3. Server** | REST APIs, microservices | HTTP requests | Running handlers |
| **4. Image** | Dockerfiles, containers | Build specs | Container images |
| **5. Container** | docker run, docker-compose | Runtime config | Active processes |
| **6. Platform** | Kubernetes, orchestration | Manifests | Cluster state |

### How Layers Relate

Each layer **adds execution context** to the layer below:
- Library = Packages + versioning/dependencies
- Server = Library + entrypoint (`main()`)
- Image = Server + runtime environment
- Container = Image + process isolation
- Platform = Containers + orchestration

**Key Point**: A server binary is just an executable library. A container is just an image with runtime parameters. Each layer builds on the one below.

## Two Fundamental Flows

### Configuration Flows Downward
Higher layers configure lower layers:
```
Platform Manifest (K8s YAML)
    ↓ specifies
Container Config (environment vars, volumes)
    ↓ references
Image (Dockerfile built artifact)
    ↓ contains
Server Binary (compiled Go code)
    ↓ imports
Library Packages
    ↓ use
Package Configurations
```

### Interfaces Flow Upward
Lower layers expose capabilities to higher layers:
```
Package Interfaces (Renderer, Processor)
    ↑ compose into
Library Public API
    ↑ used by
Server Endpoints (HTTP, gRPC)
    ↑ packaged in
Image Entrypoint
    ↑ instantiated as
Container Process
    ↑ orchestrated by
Platform Service
```

## Why This Matters

### If You Work at the Package Layer
You're defining the foundation. Your interfaces flow upward through every layer above. A well-designed package API makes services easier to build and containers easier to configure.

**Your decisions affect**: Service design, deployment complexity, testing at all layers

### If You Work at the Library Layer
You're creating reusable units. Your module boundaries determine what services can compose together. Your dependency choices cascade through images and containers.

**Your decisions affect**: Binary size, deployment time, runtime dependencies, service capabilities

### If You Work at the Server Layer
You're the bridge between code and infrastructure. Your endpoints become container entrypoints. Your request validation is the last line of defense before business logic.

**Your decisions affect**: Image design, container configuration, platform deployment patterns

### If You Work at the Image Layer
You're defining the runtime environment. Your image size affects deployment speed. Your base image choices determine security posture.

**Your decisions affect**: Container startup time, resource consumption, platform scaling patterns

### If You Work at the Container Layer
You're managing process isolation. Your resource limits affect platform capacity. Your networking config determines service connectivity.

**Your decisions affect**: Platform scheduling efficiency, cluster resource utilization

### If You Work at the Platform Layer
You're orchestrating the entire stack. Your manifests configure containers, which instantiate images, which run servers, which use libraries, which depend on packages.

**Your decisions affect**: Service availability, scaling behavior, operational complexity

## The Universal Pattern: Data vs Behavior

At every layer, there's a boundary between **specification** (data) and **execution** (behavior):

**Data (Specification)**:
- Describes what you want
- Serializable (JSON, YAML, source code)
- Validated during transformation
- Short-lived or immutable
- Examples: Config structs, go.mod, Dockerfiles, K8s manifests

**Behavior (Runtime)**:
- Does what you specified
- Active execution (processes, objects, services)
- Created from validated data
- Long-lived and stateful
- Examples: Domain objects, running services, active containers

**Transformation Point**:
- Where validation happens
- Where invalid specs are rejected
- Where data becomes behavior
- Examples: Constructors, compilers, build processes, container start, reconciliation loops

## Key Architectural Benefits

### Safety
Every entity is constructed in a validated state. Invalid configurations are caught at transformation boundaries, not at runtime.

### Clarity
Explicit separation between data and behavior makes responsibilities clear. You know where logic belongs at every layer.

### Flexibility
Interface-based boundaries allow implementations to change without breaking consumers. Lower layers evolve independently.

### Testability
Each layer is testable without requiring higher layers. Package tests don't need servers. Server tests don't need containers.

### Consistency
The same pattern at every layer means the same mental model from code to cloud. New team members recognize the structure immediately.

## Practical Implications

### When Designing Packages
- Separate configuration (data) from domain objects (behavior)
- Use transformation functions (`New*()`) that validate and return interfaces
- Configuration should never leak into business logic

### When Building Libraries
- Module boundaries should be clear and stable
- Public APIs should expose interfaces, not concrete types
- Version compatibility is a contract with all higher layers

### When Creating Services
- Request structures (Query/Command) are data that transform into executors (behavior)
- Validation happens at the service boundary, not in domain logic
- Services are just libraries with a `main()` entrypoint

### When Containerizing
- Image build specs are data; built images are behavior
- Runtime configuration (env vars, volumes) happens at container layer, not image layer
- Each container is an instance of the image with specific runtime context

### When Deploying to Platforms
- Manifests are data describing desired state
- Platform maintains actual state through continuous reconciliation
- The transformation never "completes" - it's a continuous loop

## Reading the Full Document

The complete [Layered Composition Architecture](./layered-composition-architecture.md) document provides:

- **Part I: Foundational Concepts** - Deep dive into data vs behavior, transformation patterns, interfaces, and dependencies
- **Part II: Composition Layers** - Detailed analysis of all six layers with real examples
- **Part III: Unified Patterns** - Cross-layer consistency, design principles, and complete system examples

## Who Should Care?

- **Software Engineers**: Understand how your code fits into the larger system
- **DevOps Engineers**: See how package decisions affect deployment
- **Platform Engineers**: Understand what you're orchestrating all the way down
- **Architects**: Get a unified framework for system design across all layers
- **Technical Leaders**: Understand the complete stack with one mental model

## The Bottom Line

Modern software systems compose through six natural layers. At each boundary, the same pattern repeats: specification transforms into runtime. Configuration flows down, interfaces flow up. By recognizing and respecting these natural boundaries, you build systems that are safe, clear, flexible, testable, and scalable - from the smallest package to the largest cloud platform.

**Your work at one layer affects every layer above and below.** Understanding the complete stack helps you make better decisions wherever you operate.

---

*For the complete architectural framework with detailed examples and implementation guidance, see [Layered Composition Architecture](./layered-composition-architecture.md).*
