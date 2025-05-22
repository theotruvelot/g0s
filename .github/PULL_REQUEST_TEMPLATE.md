# [Feature/Bug/Hotfix]/[ID]: Title

## Description

Brief description of the changes and the problem they solve.

### Key Components

List the main components that were added/modified:

1. **Component Name**
   - Key change/feature 1
   - Key change/feature 2
   - Testing approach

### Technical Details

- **Language**: Go
- **Test Coverage**: X% across modified packages
- **Dependencies Added**:
  - `dependency/name`: Purpose
  - `dependency/name`: Purpose

### Testing

Coverage report for modified packages:
```
pkg/example1: XX.X% coverage
pkg/example2: XX.X% coverage
```

Tests can be run using:
- `make test`: Run all tests
- `make test-nocache`: Run tests without cache
- `make test-coverage`: Run tests with coverage report

### Usage

If applicable, provide usage examples:
```bash
# Example command
make example-command ARG1=value1 ARG2=value2
```

### Configuration

List any new or modified configuration options:
- `OPTION_NAME`: Description and type
- `OPTION_NAME`: Description and type

## Breaking Changes
- List any breaking changes here
- Include migration steps if necessary

## Related Issues
- Closes #X: Issue description
- Related to #Y: Issue description

## Checklist
- [ ] Code follows project standards
- [ ] Tests added for all new functionality
- [ ] Documentation updated
- [ ] All tests passing
- [ ] Code coverage maintained/improved
- [ ] No linting errors
- [ ] Rebased on latest main/master
- [ ] All conflicts resolved

## Screenshots
If applicable, add screenshots to help explain your changes.

## Additional Notes
Any additional information that might be helpful for reviewers. 
