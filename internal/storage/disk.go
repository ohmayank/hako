package storage

import (
	"fmt"
	"os"
	"syscall"
)

const bytesPerMegabyte = 1024 * 1024

func EnsureFreeSpace(path string, minFreeMB int) error {
	if minFreeMB <= 0 {
		return nil
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create directory for disk check: %w", err)
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return fmt.Errorf("check free disk space: %w", err)
	}

	freeBytes := stat.Bavail * uint64(stat.Bsize)
	requiredBytes := uint64(minFreeMB) * bytesPerMegabyte
	if freeBytes < requiredBytes {
		return fmt.Errorf("not enough free disk space: have %d MB, need %d MB", freeBytes/bytesPerMegabyte, minFreeMB)
	}

	return nil
}
