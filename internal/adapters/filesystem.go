package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/namdang-fdp/pivot/internal/core"
)

// HostFilesystem implements Slice 1 path observations against the local host.
type HostFilesystem struct{}

// NewHostFilesystem creates a host filesystem adapter.
func NewHostFilesystem() *HostFilesystem { return &HostFilesystem{} }

// CurrentDirectory returns the process working directory.
func (*HostFilesystem) CurrentDirectory() (string, error) {
	value, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("determine current directory: %w", err)
	}
	return value, nil
}

// CanonicalDirectory resolves an existing directory to a canonical absolute path.
func (*HostFilesystem) CanonicalDirectory(value string) (string, error) {
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve directory %q: %w", value, err)
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve directory %q: %w", value, err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return "", fmt.Errorf("inspect directory %q: %w", canonical, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path %q is not a directory", canonical)
	}
	return filepath.Clean(canonical), nil
}

// DirectoryExists reports whether a path resolves to a directory.
func (*HostFilesystem) DirectoryExists(value string) bool {
	info, err := os.Stat(value)
	return err == nil && info.IsDir()
}

// FileExists reports whether a path resolves to a regular file.
func (*HostFilesystem) FileExists(value string) bool {
	info, err := os.Stat(value)
	return err == nil && info.Mode().IsRegular()
}

// FindManifestRoot searches the start directory and its ancestors for a manifest.
func (fs *HostFilesystem) FindManifestRoot(start string) (string, bool, error) {
	current, err := fs.CanonicalDirectory(start)
	if err != nil {
		return "", false, err
	}
	for {
		if fs.FileExists(filepath.Join(current, core.ManifestFilename)) {
			return current, true, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}

func validatePathWithinRoot(root, field, relative string) error {
	target := filepath.Join(root, filepath.FromSlash(relative))
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("%s %q escapes the project root", field, relative)
	}

	current := root
	for _, component := range strings.Split(filepath.Clean(filepath.FromSlash(relative)), string(filepath.Separator)) {
		current = filepath.Join(current, component)
		_, err := os.Lstat(current)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return fmt.Errorf("inspect %s %q: %w", field, relative, err)
		}
		resolved, err := filepath.EvalSymlinks(current)
		if err != nil {
			return fmt.Errorf("resolve %s %q: %w", field, relative, err)
		}
		resolvedRel, err := filepath.Rel(root, resolved)
		if err != nil || resolvedRel == ".." || strings.HasPrefix(resolvedRel, ".."+string(filepath.Separator)) || filepath.IsAbs(resolvedRel) {
			return fmt.Errorf("%s %q resolves outside the project root", field, relative)
		}
	}
	return nil
}
