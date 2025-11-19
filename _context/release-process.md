# Release Process

This document defines the versioning strategy, publishing workflow, and CHANGELOG conventions for the document-context library.

## Version Numbering Strategy

The document-context library follows [Semantic Versioning 2.0.0](https://semver.org/) with a deliberate pre-release phase to validate API design through real-world usage before committing to long-term stability.

### Pre-Release Versions (v0.x.x)

**Initial release**: `v0.1.0`

**Version format**: `v0.MINOR.PATCH`

**Breaking changes**: Allowed between minor versions during pre-release phase

**Purpose**: Validate API design and gather feedback before stability commitment

**Duration**: Until supplemental package integration validates all core abstractions

### Release Candidates (v1.0.0-rc.x)

**Version format**: `v1.0.0-rc.1`, `v1.0.0-rc.2`, etc.

**Purpose**: Signal approaching API stability and final opportunity for breaking changes

**Criteria**: All v1.0.0 features complete, comprehensive validation performed

### Stable Release (v1.0.0+)

**Version format**: `vMAJOR.MINOR.PATCH`

**API stability**: No breaking changes within major version

**Breaking changes**: Require major version increment (v1.x.x → v2.0.0)

**Commitment**: Backward compatibility guaranteed within major version

## Pre-Release Philosophy

### Purpose of Pre-Release Phase

- Validate API design through integration with go-agents and supplemental packages
- Gather feedback from real-world usage before stabilization
- Identify missing capabilities and awkward patterns
- Refine interfaces based on integration experience
- Build confidence in architectural decisions

### Development Through Usage

The pre-release phase enables validation through integration:
- **go-agents vision API**: Validates document → image → data URI pipeline
- **go-agents-document-context** (if separated): Validates standalone document processing patterns
- **Real-world applications**: Validates configuration, caching, and filter usage

This approach:
- Exercises the public API from a consumer perspective
- Reveals integration friction and missing abstractions
- Validates that the library delivers on its document processing promise
- Demonstrates practical usage patterns through examples
- Builds experience with Go package lifecycle management

### Breaking Change Policy (v0.x.x)

During pre-release, breaking changes are acceptable but managed carefully:
- Breaking changes documented in CHANGELOG.md with upgrade guidance
- Migration examples provided for significant API changes
- Breaking changes batched when possible to minimize disruption
- Clear communication about stability expectations
- Community feedback actively sought before major restructuring

### Promotion Criteria

The library graduates to v1.0.0 when:
- Integration with go-agents validated and stable
- No significant API concerns identified during usage
- Community feedback addressed and incorporated
- API surface complete for document processing scope
- Documentation comprehensive and accurate
- Test coverage maintained above 80%
- Minimum 2-3 months of pre-release usage without major issues
- Confidence in long-term API stability commitment

## CHANGELOG Format

**CHANGELOG.md** tracks version history focusing on public API changes only.

### Purpose

Document library API changes for users to understand what's available in each version.

### Format Structure

**Version heading**:
```markdown
## [vX.Y.Z] - YYYY-MM-DD
```

**Category headings** (use only categories that apply):
- `**Added**:` - New packages, types, functions, methods, capabilities
- `**Changed**:` - Modifications to existing public API
- `**Deprecated**:` - Features marked for future removal
- `**Removed**:` - Deleted features (breaking changes)
- `**Fixed**:` - Bug fixes affecting public API behavior

**Entry format**:
- Package-level bullet with concise description
- Sub-bullets for package capabilities when helpful
- Focus on what's available, not how it's implemented

### What to Include

- New packages and their purpose
- New types, interfaces, functions, methods
- Interface changes (new methods, signature changes)
- New configuration options
- Breaking changes with migration guidance
- Public API behavior fixes

### What to Exclude

- Implementation details
- Internal refactoring
- Documentation updates
- Test additions
- Bug fix implementation details (unless API behavior changed)
- Performance improvements (unless API changed)

### Example Format

```markdown
## [v0.2.0] - 2025-11-25

**Added**:

- `pkg/ocr` - OCR-based text extraction from images

  Provides OCR interface for extracting text from image-based documents. Includes TesseractOCR implementation with confidence scoring, language detection, and configurable preprocessing options.

**Changed**:

- `Page.ToImage()` now requires `image.Renderer` parameter instead of `ImageOptions` struct

  Migration: Replace `page.ToImage(opts)` with `renderer, _ := image.NewImageMagickRenderer(opts.ToConfig()); page.ToImage(renderer)`

**Fixed**:

- `PDFDocument.PageCount()` now correctly handles encrypted PDFs without pages
```

### Version Sections

**Initial Pre-Release** (v0.1.0):
- Single version section with all initial features
- Use `**Added**:` category exclusively
- Comprehensive package descriptions

**Subsequent Versions**:
- Focus on changes since previous version
- Use appropriate categories (Added, Changed, Deprecated, Removed, Fixed)
- Reference related versions for context when helpful

## Publishing Workflow

### Pre-Publication Checklist

Before publishing any version, ensure:

**Code Quality**:
- [ ] All unit tests passing with 80%+ coverage
- [ ] No critical bugs or known issues
- [ ] Code review completed for all changes since last version
- [ ] `go vet` and `go fmt` clean
- [ ] All deprecated features documented with migration path

**Documentation**:
- [ ] README.md updated with current capabilities and examples
- [ ] README.md examples verified manually
- [ ] Complete package documentation (godoc)
- [ ] ARCHITECTURE.md reflects current implementation
- [ ] PROJECT.md updated with current status
- [ ] CHANGELOG.md updated with version changes
- [ ] Migration guide provided for breaking changes (if applicable)

**Repository State**:
- [ ] All changes committed to main branch
- [ ] Git repository clean (no uncommitted changes)
- [ ] Git tags follow semantic versioning
- [ ] `.github/` directory content appropriate (if exists)

**Communication**:
- [ ] Pre-release status clearly marked in README
- [ ] Breaking change policy documented
- [ ] Expected stability communicated clearly
- [ ] Feedback channels established (GitHub issues)

### Go Module Publishing Steps

After merging to main and completing the checklist:

**1. Ensure Clean Repository State**

```bash
# Pull latest main
git checkout main
git pull origin main

# Verify no uncommitted changes
git status
```

**2. Tag the Release**

```bash
# Create and push the version tag
git tag v0.1.0
git push origin v0.1.0
```

**3. Create GitHub Release**

1. Navigate to `https://github.com/JaimeStill/document-context/releases/new`
2. Select tag: `v0.1.0`
3. Title: `v0.1.0 - [Brief heading from CHANGELOG]`
4. Description: Copy from CHANGELOG.md version section
5. Check: "Set as a pre-release" (for v0.x.x versions)
6. Click: "Publish release"

**4. Verify Publication**

```bash
# Verify the module is indexed (may take a few minutes)
go list -m -versions github.com/JaimeStill/document-context

# Should show: v0.1.0
```

**5. Test Installation**

```bash
# In a separate test directory
mkdir test-install && cd test-install
go mod init test
go get github.com/JaimeStill/document-context@v0.1.0

# Verify it appears in go.mod
cat go.mod
```

### Version Discovery

**Module Proxy**: Go module proxy indexes all tagged versions automatically

**Browse versions**:
```bash
go list -m -versions github.com/JaimeStill/document-context
```

**Install specific version**:
```bash
go get github.com/JaimeStill/document-context@v0.2.0
```

**Install latest**:
```bash
go get github.com/JaimeStill/document-context@latest
```

### Module Maintenance

**Version Immutability**: Each version is immutable once published

**Breaking Changes**: Require new version (v0.2.0, v0.3.0, or v1.0.0 depending on scope)

**CHANGELOG**: Maintained with detailed version history

**GitHub Releases**: Created for each version with release notes

**Migration Guides**: Provided for breaking changes in CHANGELOG and release notes

## Communication Strategy

### Pre-Release Transparency

The library is clearly marked as pre-release in README.md:

```markdown
## Status: Pre-Release (v0.x.x)

**document-context** is currently in pre-release development. The API may change between minor versions until v1.0.0 is released. Production use is supported, but be prepared for potential breaking changes.

We actively seek feedback on API design, missing capabilities, and integration challenges. Please open issues for any concerns or suggestions.
```

### Version Communication

**Each release includes**:
- Detailed changelog with all public API changes
- Breaking changes highlighted prominently
- Migration examples provided inline
- Deprecation warnings with removal version noted
- Community notification through GitHub releases

### Feedback Channels

**GitHub Issues**: Bug reports and feature requests

**GitHub Discussions**: API design conversations and usage questions

**Issue Templates**: Structured feedback collection

**Integration Review**: Regular review of usage patterns in dependent projects

## Post-1.0.0 Commitment

Once v1.0.0 is released:

**API Stability**: Guaranteed within major version

**Breaking Changes**: Only in major version increments

**Backward Compatibility**: Maintained for v1.x.x series

**Deprecation Cycle**: Minimum one minor version before removal

**Predictable Upgrades**: Clear upgrade path for consumers

**Long-term Support**: Commitment to API stability and maintenance
