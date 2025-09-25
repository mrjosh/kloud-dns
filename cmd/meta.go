package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Meta are the meta-options that are available on all or most commands.
type Meta struct {

	// Root cmd is the root of the command line
	// Good enough explanation!)
	RootCmd *cobra.Command

	// WorkingDir is the working directory of the command.
	WorkingDir string

	// Logger is the logger to use for the command.
	Logger *zap.Logger

	// RunningInAutomation indicates that commands are being run by an
	// automated system rather than directly at a command prompt.
	RunningInAutomation bool

	// A context.Context provided by the caller -- typically "package main" --
	// which might be carrying telemetry-related metadata and so should be
	// used when creating downstream traces, etc.
	//
	// This isn't guaranteed to be set, so use [Meta.CommandContext] to
	// safely create a context for the entire execution of a command, which
	// will be connected to this parent context if it's present.
	CallerContext context.Context

	// When this channel is closed, the command will be cancelled.
	ShutdownCh <-chan struct{}
}
