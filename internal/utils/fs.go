package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureDir ensures a directory exists.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0750)
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file %s: %w", src, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat source file %s: %w", src, err)
	}
	mode := info.Mode()

	dstDir := filepath.Dir(dst)
	if err := EnsureDir(dstDir); err != nil {
		return fmt.Errorf("creating destination directory %s: %w", dstDir, err)
	}

	// Use atomic write pattern: temp file + rename
	tmp, err := os.CreateTemp(dstDir, ".copyfile.*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dstDir, err)
	}
	tmpPath := tmp.Name()

	// Clean up temp file on any error
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpPath)
		}
	}()

	// Set the correct mode on the temp file
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("setting file mode: %w", err)
	}

	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("copying content: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("syncing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	// Atomic rename
	if runtime.GOOS == "windows" {
		if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing existing destination file for atomic rename: %w", err)
		}
	}
	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("renaming temp file to %s: %w", dst, err)
	}
	success = true
	return nil
}

// CopyDir copies a directory recursively.
func CopyDir(src, dst string) error {
	err := CopyDirWithOpts(src, dst, false)
	if err == ErrSourceMissing {
		return nil
	}
	return err
}

// ErrSourceMissing indicates the source directory does not exist.
var ErrSourceMissing = fmt.Errorf("source directory does not exist")

// CopyDirWithOpts copies a directory with options.
func CopyDirWithOpts(src, dst string, required bool) error {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			if required {
				return fmt.Errorf("required directory missing: %s", src)
			}
			// If not required and missing, return specific error so caller can decide to log or ignore
			return ErrSourceMissing
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	if err := EnsureDir(dst); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return EnsureDir(target)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("reading symlink %s: %w", path, err)
			}
			// Security check: ensure symlink doesn't point/traverse outside the target root
			// We only allow relative links that stay within the tree.
			if filepath.IsAbs(linkTarget) {
				return fmt.Errorf("insecure symlink %s -> %s: absolute links not allowed in archives", path, linkTarget)
			}
			// Check for traversal
			if strings.HasPrefix(linkTarget, "..") || strings.Contains(linkTarget, "/../") || strings.Contains(linkTarget, "\\..\\") {
				return fmt.Errorf("insecure symlink %s -> %s: directory traversal not allowed", path, linkTarget)
			}

			// Replicate the symlink
			if err := os.Symlink(linkTarget, target); err != nil {
				return fmt.Errorf("creating symlink %s -> %s: %w", target, linkTarget, err)
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			// Skip other non-regular files silently to prevent data leaks and ensure portability
			return nil
		}
		return CopyFile(path, target)
	})
}
