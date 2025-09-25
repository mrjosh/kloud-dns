package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kloud-team/dns/cmd"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	BranchName string
	GitCommit  string
	CompiledBy string
	BuildTime  string
)

func main() {

	defer cmd.PanicHandler()
	rootCmd := &cobra.Command{
		Use:   "kdns",
		Short: "the go-kloud dns command line interface",
		Long: `
   _    _                 _ 
  | | _| | ___  _   _  __| |
  | |/ / |/ _ \| | | |/ _' |  
  |   <| | (_) | |_| | (_| |  Version: v1.0
  |_|\_\_|\___/ \__,_|\__,_|  DNS Forwarder

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Create Meta for CMDs
	zapLoggerOpts := []zap.Option{
		zap.WithClock(zapcore.DefaultClock),
		zap.WithCaller(false),
	}

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC822)
	logger, err := config.Build(zapLoggerOpts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rootCmd.SetArgs(os.Args[1:])

	if err := cmd.Start(rootCmd, logger); err != nil {
		logger.Sugar().Error(err)
		os.Exit(1)
	}

}
