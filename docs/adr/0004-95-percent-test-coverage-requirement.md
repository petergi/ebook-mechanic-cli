# ADR 0004: 95% Test Coverage Requirement

## Status

Accepted

## Context

The ebook-mechanic-cli needs to be reliable and maintainable. To ensure quality, we need to establish:
- Minimum acceptable test coverage
- Types of tests required (unit, integration, e2e)
- Testing strategy for TUI components
- Coverage measurement and enforcement

### Considered Alternatives

1. **No Coverage Requirement**
   - Pros: Faster initial development
   - Cons: Technical debt, hard to refactor, bugs in production

2. **80% Coverage**
   - Pros: Industry standard, achievable
   - Cons: Too low for a CLI tool with critical file operations

3. **95% Coverage**
   - Pros: High confidence, forces good design, catches edge cases
   - Cons: More upfront effort, may encourage coverage-chasing

4. **100% Coverage**
   - Pros: Complete testing
   - Cons: Unrealistic, diminishing returns, encourages trivial tests

## Decision

We will require **95% test coverage** with the following breakdown:

- **Unit Tests**: 95%+ coverage of all packages
- **Integration Tests**: Cover all major user flows
- **TUI Tests**: Cover all models and state transitions
- **E2E Tests**: Cover complete use cases (validate, repair, batch)

Coverage will be measured using Go's built-in coverage tools and enforced in CI.

## Rationale

### Why 95% (not 80% or 100%)

**95% is high enough to**:
- Catch most bugs before they reach users
- Force thinking about edge cases during development
- Enable confident refactoring
- Serve as living documentation

**95% is realistic enough to**:
- Allow for genuinely untestable code (external dependencies)
- Avoid busywork tests just for coverage
- Balance development speed with quality

### File Operations Criticality

This CLI performs file modifications (repair operations):
- Data loss is unacceptable
- High test coverage reduces risk
- Validates backup mechanisms
- Ensures error handling works

### TUI Testing Challenges

TUI code is often undertested, but with Bubbletea's MVU architecture:
- Models are pure functions (easy to test)
- Update logic is deterministic
- Views are just string rendering
- Messages can be simulated

High coverage is achievable and valuable.

### Refactoring Confidence

95% coverage enables:
- Safe refactoring of complex operations
- Architectural changes without fear
- Performance optimizations with regression protection

## Consequences

### Positive

1. **Quality Assurance**: High confidence in code correctness
2. **Regression Prevention**: Breaking changes caught immediately
3. **Design Pressure**: Forces testable architecture
4. **Documentation**: Tests serve as usage examples
5. **Refactoring Safety**: Can modify with confidence

### Negative

1. **Development Time**: More upfront test writing
2. **Maintenance**: Tests need maintenance along with code
3. **False Security**: Coverage doesn't guarantee correctness

### Mitigation

- Write tests alongside code (not after)
- Use table-driven tests to reduce boilerplate
- Focus on behavior, not implementation details
- Use test helpers to reduce duplication

## Implementation Strategy

### Coverage Measurement

Use Go's built-in coverage tools:

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total
```

### CI Enforcement

Add coverage check to CI pipeline:

```yaml
# .github/workflows/test.yml
- name: Run tests with coverage
  run: go test -coverprofile=coverage.out ./...

- name: Check coverage
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 95.0" | bc -l) )); then
      echo "Coverage is below 95%: $COVERAGE%"
      exit 1
    fi
```

### Test Organization

#### 1. Unit Tests (Per Package)

Test individual functions and methods:

```go
// internal/operations/validate_test.go
func TestValidateEPUB(t *testing.T) {
    tests := []struct {
        name    string
        file    string
        want    bool
        wantErr bool
    }{
        {"valid epub", "testdata/valid.epub", true, false},
        {"invalid epub", "testdata/invalid.epub", false, false},
        {"missing file", "nonexistent.epub", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ValidateEPUB(tt.file)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEPUB() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got.IsValid != tt.want {
                t.Errorf("ValidateEPUB() = %v, want %v", got.IsValid, tt.want)
            }
        })
    }
}
```

#### 2. TUI Model Tests

Test Bubbletea models:

```go
// internal/tui/models/menu_test.go
func TestMenuModel_Update(t *testing.T) {
    tests := []struct {
        name    string
        model   MenuModel
        msg     tea.Msg
        want    MenuModel
    }{
        {
            name:  "navigate down",
            model: MenuModel{selected: 0, options: []string{"A", "B", "C"}},
            msg:   tea.KeyMsg{Type: tea.KeyDown},
            want:  MenuModel{selected: 1, options: []string{"A", "B", "C"}},
        },
        {
            name:  "navigate up from top wraps to bottom",
            model: MenuModel{selected: 0, options: []string{"A", "B", "C"}},
            msg:   tea.KeyMsg{Type: tea.KeyUp},
            want:  MenuModel{selected: 2, options: []string{"A", "B", "C"}},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, _ := tt.model.Update(tt.msg)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Update() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### 3. Integration Tests

Test complete flows:

```go
// tests/integration/validation_flow_test.go
func TestValidationFlow(t *testing.T) {
    // Create test EPUB
    testFile := createTestEPUB(t)
    defer os.Remove(testFile)

    // Initialize app
    app := tui.NewApp()

    // Simulate user flow: select file -> validate -> view report
    app.SendMessage(SelectFileMsg{Path: testFile})
    app.SendMessage(ValidateMsg{})

    // Wait for operation to complete
    time.Sleep(100 * time.Millisecond)

    // Verify report was generated
    if !app.HasReport() {
        t.Error("Expected report to be generated")
    }
}
```

#### 4. E2E Tests

Test complete user scenarios:

```go
// tests/integration/e2e_test.go
func TestE2E_ValidateAndRepair(t *testing.T) {
    // Setup test environment
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "broken.epub")
    copyTestFile(t, "testdata/broken.epub", testFile)

    // Run validation
    validateOp := operations.NewValidate()
    report, err := validateOp.Execute(context.Background(), testFile)
    require.NoError(t, err)
    assert.False(t, report.IsValid)

    // Run repair
    repairOp := operations.NewRepair()
    result, err := repairOp.Execute(context.Background(), testFile)
    require.NoError(t, err)
    assert.True(t, result.Success)

    // Verify repair worked
    report2, err := validateOp.Execute(context.Background(), testFile)
    require.NoError(t, err)
    assert.True(t, report2.IsValid)
}
```

### Test Helpers

Create helpers to reduce boilerplate:

```go
// tests/helpers/fixtures.go
func LoadTestEPUB(t *testing.T, name string) string {
    t.Helper()
    path := filepath.Join("testdata", "epub", name)
    if _, err := os.Stat(path); err != nil {
        t.Fatalf("Test file not found: %s", path)
    }
    return path
}

func CreateBrokenEPUB(t *testing.T) string {
    t.Helper()
    tmpDir := t.TempDir()
    path := filepath.Join(tmpDir, "broken.epub")
    // Create intentionally broken EPUB
    return path
}
```

### Coverage Exclusions

Explicitly exclude genuinely untestable code:

```go
// +build !test

// main.go - entry point is hard to test
func main() {
    // ...
}
```

Document why code is excluded:
- External dependencies that can't be mocked
- Platform-specific code
- Main entry points

## Monitoring and Reporting

### Coverage Reports

Generate coverage reports on each CI run:
- HTML report as artifact
- Coverage badge in README
- Trend tracking over time

### Coverage by Package

Track coverage per package:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Expected breakdown:
- `internal/tui/models`: 95%+
- `internal/operations`: 95%+
- `pkg/utils`: 98%+ (simpler code)
- `cmd/ebm`: 60%+ (entry point, harder to test)

## Review Process

### Pull Request Requirements

Every PR must:
1. Maintain or improve overall coverage
2. Include tests for all new code
3. Include integration test for new features
4. Update existing tests if behavior changes

### Coverage Checks

Automated checks will:
- Fail PR if coverage drops below 95%
- Show coverage diff in PR comments
- Highlight uncovered lines

## Testing Best Practices

1. **Test Behavior, Not Implementation**: Focus on what code does, not how
2. **Table-Driven Tests**: Use for multiple scenarios
3. **Test Helpers**: Extract common setup/assertions
4. **Mock External Dependencies**: Use interfaces for testability
5. **Fast Tests**: Unit tests should run in milliseconds
6. **Isolated Tests**: No shared state between tests
7. **Clear Test Names**: Describe what is being tested

## References

- [Go Testing Documentation](https://go.dev/doc/tutorial/add-a-test)
- [Go Coverage Tool](https://go.dev/blog/cover)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testing Best Practices](https://go.dev/doc/effective_go#testing)
