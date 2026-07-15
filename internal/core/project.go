package core

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var projectIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ProjectID is a stable, human-readable project identifier.
type ProjectID string

// ParseProjectID validates and returns a project identifier.
func ParseProjectID(value string) (ProjectID, error) {
	if !projectIDPattern.MatchString(value) {
		return "", fmt.Errorf("project.id %q is invalid: use lowercase letters, digits, and single hyphens", value)
	}
	return ProjectID(value), nil
}

// SlugifyProjectID derives a project identifier for explicit init operations.
// Use builder to decrease allocation
// TODO: This function is bad when translate Vietnamese due to unicode.MaxASCII - no blocker
func SlugifyProjectID(value string) (ProjectID, error) {
	var b strings.Builder
	separator := false
	for _, r := range strings.ToLower(value) {
		if r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			if separator && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			separator = false
			continue
		}
		separator = true
	}
	return ParseProjectID(b.String())
}

// RegisteredProject describes one project in the machine registry.
type RegisteredProject struct {
	ID   ProjectID
	Name string
	Path string
}
