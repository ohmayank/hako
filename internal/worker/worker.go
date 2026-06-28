package worker

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ohmayank/hako/internal/transfer"
)

const defaultInterval = 30 * time.Second

type Options struct {
	RemoteDir string
	QueueDir  string
	Interval  time.Duration
	Logger    *log.Logger
}

type Result struct {
	Attempted int
	Succeeded int
	Failed    int
}

func Run(ctx context.Context, options Options) error {
	if options.Interval <= 0 {
		options.Interval = defaultInterval
	}

	if _, err := RetryOnce(options); err != nil {
		return err
	}

	ticker := time.NewTicker(options.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := RetryOnce(options); err != nil {
				return err
			}
		}
	}
}

func RetryOnce(options Options) (Result, error) {
	if options.QueueDir == "" {
		return Result{}, fmt.Errorf("queue directory is required")
	}
	if options.RemoteDir == "" {
		return Result{}, fmt.Errorf("remote directory is required")
	}

	entries, err := os.ReadDir(options.QueueDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, nil
		}
		return Result{}, fmt.Errorf("scan queue directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var result Result
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".tmp") {
			continue
		}

		result.Attempted++
		queuePath := filepath.Join(options.QueueDir, entry.Name())
		finalName := originalFileName(entry.Name())

		logf(options, "retry started file=%q final=%q", entry.Name(), finalName)
		if _, err := transfer.CopyAtomicAs(queuePath, options.RemoteDir, finalName); err != nil {
			result.Failed++
			logf(options, "retry failed file=%q error=%q", entry.Name(), err)
			continue
		}

		if err := os.Remove(queuePath); err != nil {
			result.Failed++
			logf(options, "retry cleanup failed file=%q error=%q", entry.Name(), err)
			continue
		}
		if err := syncDir(options.QueueDir); err != nil {
			return result, fmt.Errorf("sync queue directory: %w", err)
		}

		result.Succeeded++
		logf(options, "retry succeeded file=%q final=%q", entry.Name(), finalName)
	}

	return result, nil
}

func originalFileName(queueName string) string {
	parts := strings.SplitN(queueName, "_", 3)
	if len(parts) != 3 {
		return queueName
	}
	if _, err := time.Parse("20060102T150405", parts[0]); err != nil {
		return queueName
	}
	if len(parts[1]) != 8 {
		return queueName
	}
	if _, err := hex.DecodeString(parts[1]); err != nil {
		return queueName
	}
	if parts[2] == "" {
		return queueName
	}
	return filepath.Base(parts[2])
}

func syncDir(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()

	return directory.Sync()
}

func logf(options Options, format string, args ...any) {
	if options.Logger == nil {
		return
	}
	options.Logger.Printf(format, args...)
}
