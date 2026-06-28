package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ohmayank/hako/internal/queue"
	"github.com/ohmayank/hako/internal/storage"
	"github.com/ohmayank/hako/internal/transfer"
	"github.com/ohmayank/hako/internal/worker"
	"github.com/spf13/cobra"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newRootCommand().ExecuteContext(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		fmt.Fprintf(os.Stderr, "failed: %v\n", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hako",
		Short: "Reliable local-to-remote file ingestion",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newIngestCommand(), newWorkerCommand())
	return cmd
}

func newIngestCommand() *cobra.Command {
	var remoteDir string
	var queueDir string
	var dbPath string
	var minFreeMB int

	cmd := &cobra.Command{
		Use:   "ingest <source-file>",
		Short: "Copy a file to the remote destination atomically",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = dbPath

			sourcePath := args[0]
			if err := validateSource(sourcePath); err != nil {
				return err
			}

			deliveredPath, err := transfer.CopyAtomic(sourcePath, remoteDir)
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "delivered: %s\n", filepath.Base(deliveredPath))
				return nil
			}

			if errors.Is(err, transfer.ErrCollision) || errors.Is(err, transfer.ErrTempExists) {
				return err
			}

			fmt.Fprintf(cmd.ErrOrStderr(), "remote transfer failed: %v\n", err)

			if err := storage.EnsureFreeSpace(queueDir, minFreeMB); err != nil {
				return err
			}

			queuedName, queueErr := queue.Enqueue(sourcePath, queueDir)
			if queueErr != nil {
				return fmt.Errorf("queue fallback failed after remote transfer failed: %w", queueErr)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "queued: %s\n", queuedName)
			return nil
		},
	}

	cmd.Flags().StringVar(&remoteDir, "remote", "./remote", "remote destination directory")
	cmd.Flags().StringVar(&queueDir, "queue", "./queue", "local queue directory")
	cmd.Flags().StringVar(&dbPath, "db", "./hako.db", "SQLite metadata path reserved for later phases")
	cmd.Flags().IntVar(&minFreeMB, "min-free-mb", 1024, "minimum free local disk space before queueing")
	return cmd
}

func newWorkerCommand() *cobra.Command {
	var remoteDir string
	var queueDir string
	var intervalSeconds int

	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Retry queued files until delivered",
		RunE: func(cmd *cobra.Command, args []string) error {
			if intervalSeconds <= 0 {
				return fmt.Errorf("interval-seconds must be positive")
			}

			logger := log.New(cmd.ErrOrStderr(), "", log.LstdFlags)
			return worker.Run(cmd.Context(), worker.Options{
				RemoteDir: remoteDir,
				QueueDir:  queueDir,
				Interval:  time.Duration(intervalSeconds) * time.Second,
				Logger:    logger,
			})
		},
	}

	cmd.Flags().StringVar(&remoteDir, "remote", "./remote", "remote destination directory")
	cmd.Flags().StringVar(&queueDir, "queue", "./queue", "local queue directory")
	cmd.Flags().IntVar(&intervalSeconds, "interval-seconds", 30, "seconds between queue retry scans")
	return cmd
}

func validateSource(sourcePath string) error {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return fmt.Errorf("%w: %s is not a regular file", transfer.ErrInvalidSource, sourcePath)
	}

	return nil
}
