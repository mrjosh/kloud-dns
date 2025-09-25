package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloud-team/dns/config"
	"github.com/kloud-team/dns/forwarder"
	"github.com/spf13/cobra"
)

type ServerCommandFlags struct {
	Host           string
	Port           int32
	ConfigFilePath string
}

type ServerCommand struct {
	m *Meta
}

func (c *ServerCommand) cmd() *cobra.Command {
	cFlags := new(ServerCommandFlags)
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run a kloud-dns server node instance",
		RunE: func(cmd *cobra.Command, args []string) error {

			cmd.Println(c.m.RootCmd.Long)
			log := c.m.Logger.Sugar()

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Load configfile
			cfg, err := config.Load(cFlags.ConfigFilePath)
			if err != nil {
				return err
			}

			if cFlags.Host != "" {
				cfg.Host = cFlags.Host
			}

			if cFlags.Port != 0 {
				cfg.Port = cFlags.Port
			}

			log.Infof("config file [%s] loaded successfully.", cFlags.ConfigFilePath)

			msg := make(chan error)
			go func() {
				msg <- forwarder.Start(ctx, cfg, log)
			}()

			go func() {
				c := make(chan os.Signal, 1)
				signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
				msg <- fmt.Errorf("%s", <-c)
			}()

			return <-msg

		},
	}

	cmd.SuggestionsMinimumDistance = 1

	cmd.PersistentFlags().StringVarP(&cFlags.ConfigFilePath, "config-file", "c", "server.hcl", "Using this flag you can specify the config file path.")
	cmd.PersistentFlags().StringVarP(&cFlags.Host, "host", "H", "", "Using this flag you can specify the host addr to listen on")
	cmd.PersistentFlags().Int32VarP(&cFlags.Port, "port", "P", 0, "Using this flag you can specify the host port to listen on")
	cmd.MarkFlagsOneRequired("config-file")
	return cmd
}

// Add the current command to cobra interface
func (c *ServerCommand) AddCommandToCobra(rootCmd *cobra.Command) {
	rootCmd.AddCommand(c.cmd())
}
