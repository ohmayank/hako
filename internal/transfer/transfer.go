package transfer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const copyBufferSize = 256 * 1024

var (
	ErrInvalidSource = errors.New("invalid source file")
	ErrCollision     = errors.New("destination already exists")
	ErrTempExists    = errors.New("temporary destination already exists")
)

// CopyAtomic streams srcPath into remoteDir using a temporary file, then renames
// it into place only after the copy has completed successfully.
func CopyAtomic(srcPath, remoteDir string) (string, error) {
	return CopyAtomicAs(srcPath, remoteDir, filepath.Base(srcPath))
}

// CopyAtomicAs streams srcPath into remoteDir under fileName using a temporary
// file, then renames it into place only after the copy has completed.
func CopyAtomicAs(srcPath, remoteDir, fileName string) (string, error) {
	if fileName == "" || fileName != filepath.Base(fileName) {
		return "", fmt.Errorf("%w: invalid destination filename %q", ErrInvalidSource, fileName)
	}

	sourceFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open source: %w", err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return "", fmt.Errorf("stat source: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return "", fmt.Errorf("%w: %s is not a regular file", ErrInvalidSource, srcPath)
	}

	if err := os.MkdirAll(remoteDir, 0o755); err != nil {
		return "", fmt.Errorf("create remote directory: %w", err)
	}

	finalPath := filepath.Join(remoteDir, fileName)
	tempPath := finalPath + ".tmp"

	if _, err := os.Stat(finalPath); err == nil {
		return "", fmt.Errorf("%w: %s", ErrCollision, finalPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("check destination: %w", err)
	}

	tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", fmt.Errorf("%w: %s", ErrTempExists, tempPath)
		}
		return "", fmt.Errorf("create temporary destination: %w", err)
	}

	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	copyBuffer := make([]byte, copyBufferSize)
	if _, err := io.CopyBuffer(tempFile, sourceFile, copyBuffer); err != nil {
		_ = tempFile.Close()
		return "", fmt.Errorf("copy to temporary destination: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return "", fmt.Errorf("sync temporary destination: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close temporary destination: %w", err)
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", fmt.Errorf("rename temporary destination: %w", err)
	}

	if err := syncDir(remoteDir); err != nil {
		return "", fmt.Errorf("sync remote directory: %w", err)
	}

	removeTemp = false
	return finalPath, nil
}

func syncDir(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()

	return directory.Sync()
}
