## Summary

<!-- Provide a brief, high-level summary of what this PR implements -->

This PR implements [brief description of the main changes].

## Key Features Added

<!-- Organize features by category/domain -->

### [Category/Domain Name]

- **Feature 1**: Description of what was added
- **Feature 2**: Description of what was added
- **Feature 3**: Description of what was added

### [Another Category/Domain Name]

- **Feature 1**: Description of what was added
- **Feature 2**: Description of what was added

### Infrastructure Improvements

- **Improvement 1**: Description
- **Improvement 2**: Description

## Architecture Highlights

<!-- Describe architectural decisions, patterns used, or design considerations -->

**Design Patterns:**
- Pattern or approach used (e.g., Domain-Driven Design, Clean Architecture)
- Key architectural decisions

**Key Components:**
- Component 1: Description
- Component 2: Description

**Security & Best Practices:**
- Security considerations
- Best practices applied

## Type of Change

<!-- Mark the relevant option with an 'x' -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Refactoring (no functional changes)
- [ ] Performance improvement
- [ ] Test addition or update

## Related Issue

<!-- Link to the related issue (if applicable) -->
Closes #

## Test Plan

<!-- Checklist of testing activities -->

- [ ] Build project: `go build ./...`
- [ ] Run unit tests: `go test ./...`
- [ ] Run integration tests (if applicable)
- [ ] Manual testing performed
- [ ] All API endpoints tested
- [ ] Error handling verified
- [ ] Edge cases tested
- [ ] Performance tested (if applicable)
- [ ] Security checks performed
- [ ] No sensitive data in commits

### Test Steps

<!-- Provide specific steps to test the changes -->

1. Step 1
2. Step 2
3. Step 3

## Files Changed

<!-- Update with actual numbers after creating PR -->
- X files changed
- X insertions(+)
- X deletions(-)

## Database Migration

<!-- If applicable, describe any database changes or migrations needed -->

N/A

<!-- Example if migration needed:
After deployment, run migration:
```bash
go run cmd/cli/migrate-indexes/main.go
```
-->

## Dependencies

<!-- Note any new dependencies added -->

- [ ] No new dependencies added
- [ ] New dependencies added (list below)
- [ ] Dependencies updated (list below)

<!-- If dependencies changed, list them:
- `package-name`: version - reason for addition/update
-->

Check `go.mod` for new dependencies and ensure they're reviewed for security.

## Checklist

<!-- Mark completed items with an 'x' -->

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings or errors
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published
- [ ] All method names, comments, and documentation are in English (as per `.cursorrules`)
- [ ] OpenAPI specification updated (if API changes were made)
- [ ] Logging added for audit purposes (if applicable)

## Additional Notes

<!-- Any additional information, context, or notes for reviewers -->

<!-- Screenshots, diagrams, or other visual aids can be added here -->
