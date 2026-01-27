## Review Style

This review follows the Conventional Comments specification:
https://conventionalcomments.org/

## Issues

## Suggestions
- Consider renaming ResolvePackage to better reflect its recursive behavior
- Consider renaming `NpmPackageVersion` to clarify that it is not a direct representation of registry data, but our resolver constructed tree node
- Sanitise and validate user inputs

## Questions
- Is output ordering important for the user?
- Should all ResolvePackage errors be treated as internal server errors?

