## Review Style

This review follows the Conventional Comments specification:
https://conventionalcomments.org/

## Issues
- Cyclic dependencies not handled
- Potential infinite recursion

## Suggestions
- Caching of resolved dependencies
- Expand testing for large graphs and cycles
- Consider renaming ResolvePackage to better reflect its recursive behavior
- Consider renaming `NpmPackageVersion` to clarify that it is not a direct representation of registry data, but our resolver constructed tree node

## Questions
- Is output ordering important for the user?
- Should we fail if any version strings are invalid in `ResolveHighestVersion`?
- Should all ResolvePackage errors be treated as internal server errors?

