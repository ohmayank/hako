package queue

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const copyBufferSize = 256 * 1024

func Enqueue(srcPath, queueDir string) (string, error) {
	sourceFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open source for queue: %w", err)
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return "", fmt.Errorf("stat source for queue: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return "", fmt.Errorf("invalid source file: %s is not a regular file", srcPath)
	}

	if err := os.MkdirAll(queueDir, 0o755); err != nil {
		return "", fmt.Errorf("create queue directory: %w", err)
	}

	queuedName, queuedPath, err := reserveQueuePath(srcPath, queueDir)
	if err != nil {
		return "", err
	}
	tempPath := queuedPath + ".tmp"

	queueFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("create temporary queue file: %w", err)
	}

	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	copyBuffer := make([]byte, copyBufferSize)
	if _, err := io.CopyBuffer(queueFile, sourceFile, copyBuffer); err != nil {
		_ = queueFile.Close()
		return "", fmt.Errorf("copy to queue file: %w", err)
	}

	if err := queueFile.Sync(); err != nil {
		_ = queueFile.Close()
		return "", fmt.Errorf("sync queue file: %w", err)
	}

	if err := queueFile.Close(); err != nil {
		return "", fmt.Errorf("close queue file: %w", err)
	}

	if err := os.Rename(tempPath, queuedPath); err != nil {
		return "", fmt.Errorf("rename queue file: %w", err)
	}

	if err := syncDir(queueDir); err != nil {
		return "", fmt.Errorf("sync queue directory: %w", err)
	}

	removeTemp = false
	return queuedName, nil
}

func reserveQueuePath(srcPath, queueDir string) (string, string, error) {
	for range 10 {
		queuedName, err := makeQueueName(filepath.Base(srcPath))
		if err != nil {
			return "", "", err
		}

		queuedPath := filepath.Join(queueDir, queuedName)
		if _, err := os.Stat(queuedPath); errors.Is(err, os.ErrNotExist) {
			return queuedName, queuedPath, nil
		} else if err != nil {
			return "", "", fmt.Errorf("check queue destination: %w", err)
		}
	}

	return "", "", fmt.Errorf("reserve queue file: too many filename collisions")
}

func makeQueueName(fileName string) (string, error) {
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate queue suffix: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102T150405")
	return fmt.Sprintf("%s_%s_%s", timestamp, hex.EncodeToString(randomBytes), fileName), nil
}

func syncDir(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()

	return directory.Sync()
}
