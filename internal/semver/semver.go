package semver

import (
	"errors"
	"fmt"
	"iter"

	"github.com/Masterminds/semver/v3"
)

// ResolveHighestVersion resolves the highest version, from the versions list, that satisfies the constraint.
// If if there is no such version, an error is returned.
func ResolveHighestVersion(constraint *semver.Constraints, versions iter.Seq[string]) (string, error) {
	var (
		errs    error
		highest *semver.Version
	)

	for version := range versions {
		v, err := semver.StrictNewVersion(version)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("version %s: %w", version, err))
			continue
		}
		if constraint.Check(v) && (highest == nil || v.GreaterThan(highest)) {
			highest = v
		}
	}
	// question: should we fail if any version strings are invalid?
	// If a valid highest version was found, we continue parsing all versions
	// and any parsing error causes the function to return an error even though
	// we have the highest version

	if errs != nil {
		return "", errs
	}

	if highest == nil {
		return "", errors.New("no compatible versions found")
	}

	return highest.String(), nil
}
