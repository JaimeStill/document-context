# Development Process

This document describes the development workflow for the document-context project, including session planning, implementation guides, validation, and documentation phases.

## Session Workflow Overview

Development sessions follow a structured workflow to maintain clarity and enable the developer to execute implementations independently with guidance.

**Workflow Steps**:
1. **Planning Phase**: Collaborative discussion to explore implementation approaches and align on architectural decisions
2. **Plan Presentation**: Present outline of implementation guide structure for final review
3. **Plan Approval**: Developer approves implementation guide outline
4. **Implementation Guide Creation**: Create detailed step-by-step implementation guide
5. **Developer Execution**: Developer implements code following the guide
6. **Validation Phase**: Validate implementation, run tests, verify correctness
7. **Documentation Phase**: Update documentation to reflect implemented features
8. **Session Closeout**: Create summary and clean up temporary documents

## Planning Phase

### Collaborative Exploration

The planning phase is a **collaborative discussion** where we explore different implementation approaches together. This is not a one-shot plan presentation, but an iterative conversation to align on the best solution.

**Discussion Focus**:
- Explore different architectural approaches
- Discuss trade-offs between implementation strategies
- Align on design patterns and abstractions
- Identify integration points with existing code
- Work through technical challenges and edge cases
- Consider testing strategies and validation approaches

**Key Principle**: Meaningful discussion and alignment must happen BEFORE presenting an execution plan.

### Plan Mode Protocol

When in plan mode:

1. **Engage in Collaborative Discussion**
   - Explore implementation options together
   - Ask clarifying questions as needed
   - Discuss architectural trade-offs
   - Align on best approach through iteration

2. **Present Implementation Guide Outline** (via ExitPlanMode tool)
   - After alignment through discussion
   - Outline structure of the implementation guide
   - High-level task breakdown
   - Expected deliverables
   - NOT full implementation details yet

3. **Wait for Approval**
   - Developer reviews outline
   - Requests modifications if needed
   - Approves outline structure

4. **Create Detailed Implementation Guide**
   - After approval, generate full step-by-step guide
   - Include specific file paths, signatures, logic flow
   - Ready for developer execution

5. **Never Make Code Changes** While in Plan Mode
   - Planning is for discussion and guide creation
   - Code changes happen during developer execution phase

### Plan Presentation Format

The plan outline presented via ExitPlanMode should include:
- Brief overview (2-3 sentences)
- Core components to be created
- Key architectural decisions made during discussion
- High-level task breakdown (not detailed steps yet)
- Expected deliverables
- Estimated complexity

**What NOT to Include**:
- Detailed implementation steps (those go in the implementation guide)
- Specific code blocks (those go in the implementation guide)
- Testing details (those go in the validation phase)

## Implementation Guides

### Purpose

Provide comprehensive step-by-step instructions for code implementation that the developer executes independently.

### Storage Location

`_context/##-[guide-title].md`
- Numbered sequentially (01, 02, 03, etc.)
- Descriptive title reflecting session goal
- Example: `_context/01-configuration-foundation.md`

### Content Guidelines

Implementation guides contain:
- **Pure implementation steps only**
- File paths, function signatures, struct definitions
- Exact code structure and logic flow
- Task breakdown with clear goals

Implementation guides do NOT contain:
- Code comments or documentation strings (AI assistant adds these after validation)
- Testing instructions (handled in validation phase)
- Documentation updates (handled in documentation phase)

### Structure Template

```markdown
# Session Title

## Overview
Brief description of what will be implemented

## Task Breakdown

### Task 1: [Name]

**Goal**: Clear objective for this task

**Steps**:
1. Create file X at path Y
2. Define struct Z with fields A, B, C
3. Implement method M with logic:
   - Step detail
   - Step detail

### Task 2: [Name]

**Goal**: Clear objective for this task

**Steps**:
1. Update file A
2. Add method B
3. Integrate with C

## Integration Notes

Any important notes about how components work together
```

### Lifecycle

Implementation guides are **temporary working documents** removed after session completion. They are replaced by session summaries that document what was accomplished.

## Developer Execution Phase

During this phase:
- Developer implements code structure following the implementation guide
- Developer can ask clarifying questions
- Implementation may reveal edge cases or improvements
- Developer validates code compiles and basic functionality works

## Validation Phase

After developer completes implementation, perform validation:

### Verification Steps

1. **Code Structure Review**
   - Verify code structure matches implementation guide
   - Check adherence to design principles (see CLAUDE.md)
   - Validate proper encapsulation and interfaces

2. **Testing**
   - Run all tests: `go test ./...`
   - Check test coverage: `go test ./... -cover`
   - Verify coverage targets met (typically 80%+)

3. **Build Verification**
   - Verify code builds: `go build ./...`
   - Check for any compilation warnings

4. **Code Quality**
   - Review error handling patterns
   - Check adherence to project conventions
   - Validate proper use of configuration patterns

5. **Implementation Completeness**
   - Verify all implementation guide tasks completed
   - Check that code structure matches guide specifications
   - Confirm no placeholder or incomplete implementations

### Outcome

Report findings with specific recommendations:
- **Pass**: Implementation meets all criteria, ready for documentation phase
- **Revisions Needed**: List specific issues to address before proceeding
- **Improvements Suggested**: Optional enhancements for consideration

## Documentation Phase

After successful validation, update project documentation:

### Documentation Updates

After successful validation, AI assistant updates:

1. **Code Documentation**
   - Add package documentation (doc.go files)
   - Add struct and method godoc comments
   - Add interface documentation with usage examples
   - Ensure all exported symbols are documented

2. **ARCHITECTURE.md**
   - Add new components and interfaces
   - Document design patterns used
   - Update package structure diagrams
   - Add implementation details for new features

3. **PROJECT.md**
   - Mark completed checklist items
   - Update current status
   - Note any deviations from roadmap

4. **README.md** (if user-facing changes)
   - Update prerequisites or installation steps
   - Add or update usage examples
   - Update configuration documentation

5. **Session Summary**
   - Create comprehensive record in `_context/sessions/##-[title].md` documenting:
     - What was implemented (detailed component breakdown)
     - Architectural decisions made and rationale
     - Challenges encountered and solutions
     - Test coverage achieved with metrics
     - Documentation updates (code and project docs)
     - Files created and modified
     - Git commit reference
     - Success criteria met

### Documentation Principle

**Documentation happens AFTER implementation and validation, not during.**

This ensures documentation reflects actual implementation, not planned implementation.

## Session Closeout Process

After completing the documentation phase, perform session closeout:

### Closeout Steps

1. **Create Session Summary**

   Document the completed session at `_context/sessions/##-[title].md` with:
   - What was implemented (actual outcomes)
   - Key architectural decisions made
   - Challenges encountered and solutions
   - Test coverage achieved (with metrics)
   - Documentation updates made
   - Links to relevant commits

2. **Remove Implementation Guide**

   Delete the temporary implementation guide at `_context/##-[guide-title].md`
   - Guide served its purpose
   - Session summary is permanent record
   - Avoids maintaining duplicate documentation

3. **Update Core Documents**

   Ensure all core documents reflect the completed session:
   - **ARCHITECTURE.md**: Current implementation state
   - **PROJECT.md**: Updated roadmap and status
   - **README.md**: User-facing changes (if any)

### Closeout Principle

**Session summaries are permanent documentation; implementation guides are temporary scaffolding.**

## Session Summaries

### Purpose

Document completed development sessions for future reference and project history.

### Storage Location

`_context/sessions/##-[summary-title].md`
- Numbered to match implementation guide
- Created AFTER session completion
- Permanent documentation
- Example: `_context/sessions/01-configuration-foundation.md`

### Content Structure

```markdown
# Session ##: [Title]

## Overview
What was implemented and why

## Implementation Summary
- Package/files created
- Interfaces defined
- Key implementations
- Integration points

## Architectural Decisions
Key decisions made during planning and implementation

## Challenges and Solutions
Problems encountered and how they were resolved

## Test Coverage
- Coverage percentage
- Test approach
- Key test cases

## Documentation Updates
- ARCHITECTURE.md changes
- PROJECT.md changes
- README.md changes
- Code documentation added

## Commits
- Links to relevant commits
- PR references (if applicable)

## Lessons Learned
Insights for future sessions
```

### Lifecycle

Permanent documentation maintained throughout project lifetime.

## Session Structure Guidelines

### Session Goals

Each session should have:
- **Clear, focused goal statement**: What specific capability will be implemented?
- **Specific deliverables**: What artifacts will exist after completion?
- **Success criteria**: How will we know the session succeeded?
- **Estimated effort**: Rough time estimate for planning (2-3 hours, 4-5 hours, etc.)

### Task Organization

Break sessions into focused tasks:
- Each task has a clear objective
- Tasks ordered by dependency (foundation first)
- Include specific file paths and function signatures
- Provide logic flow for complex implementations
- Keep tasks focused (30-60 minutes of work each)

### Deliverable Tracking

Track deliverables with checkboxes:
- ✅ Package structures created
- ✅ Interfaces defined
- ✅ Implementations complete
- ✅ Tests written with coverage targets met
- ✅ Documentation updated

## Phase and Milestone Planning

### Phase Structure

Large features are broken into phases:
- Each phase has a clear goal (e.g., "Complete v0.1.0 release")
- Phases are broken into multiple focused sessions
- Sessions are ordered by dependency (low-level → high-level)
- Each session can be executed independently

### Session Dependencies

Document dependencies between sessions:
- **Sequential dependencies**: Sessions that must run in order (e.g., cache interface before cache implementation)
- **Parallel opportunities**: Sessions that can run concurrently (e.g., filter integration and cache integration)
- **Dependency diagrams**: Visual representation of session ordering in PROJECT.md

### Milestone Tracking

Track progress against milestones:
- Mark completed sessions with ✅
- Track completion percentage for each phase
- Document any deviations from plan
- Update estimates based on actual effort
- Adjust roadmap as needed based on lessons learned

## Completed Sessions

For completed development sessions, see:
- `_context/sessions/01-configuration-foundation.md`
- `_context/sessions/02-cache-logging-infrastructure.md`
- `_context/sessions/03-cache-registry-infrastructure.md`
- `_context/sessions/04-filesystem-cache-implementation.md`
- `_context/sessions/05-imagemagick-filter-integration.md` (pending)

For upcoming sessions, see PROJECT.md "v0.1.0 Completion Roadmap" section.
