# Code Review Area 9: API Design & Code Organization

## Why This Matters

Clean package boundaries and minimal exported APIs make code easier to maintain and harder to misuse. This review checks that responsibilities are well-separated and naming is consistent.

## Files to Review

- All packages - exported types and functions
- `internal/cli/` - 23 command files
- Package import graphs

## What to Check

- **Exported vs unexported**: is each package's public API minimal and intentional?
- **Naming consistency**: do similar operations use similar names across packages?
  - Store pattern: Create/Get/Update/Delete/List
  - Consistent use of ID types, error variables, struct field names
- **Package boundaries**: are responsibilities clearly separated?
  - Does `cli/` contain business logic that belongs in domain packages?
  - Does `daemon/` reach into other packages' internals?
  - Are there circular dependency risks?
- **Code duplication**: are patterns like atomic writes, store operations, or config loading duplicated or shared?
- **File organization**: are 23 CLI command files in one package manageable? Should they be grouped?
- **Struct design**: are types well-designed with clear ownership?
  - Mutable vs immutable fields
  - Optional vs required fields
  - JSON serialization tags consistency

## Severity Estimate

**LOW** - Organizational issues are about maintainability, not correctness. Still valuable for long-term health.
