package main

import (
	"flag"
	"github.com/flowbreeze/cert-manager-webhook-ali/pkg/alidns"
	"github.com/flowbreeze/cert-manager-webhook-ali/pkg/log"
	"github.com/flowbreeze/cert-manager-webhook-ali/pkg/option"
	"github.com/flowbreeze/cert-manager-webhook-ali/pkg/util/exit"
	"github.com/spf13/cobra"
	"k8s.io/component-base/logs"
	"os"
	"runtime"
)

func main() {
	defer exit.WaitForExit()

	logs.InitLogs()
	defer logs.FlushLogs()

	ctx := exit.BackgroundCtx()
	stopCh := ctx.Done()
	logger := log.FromContext(ctx)
	webhook := alidns.NewSolver()

	groupName := os.Getenv("GROUP_NAME")
	if groupName == "" {
		logger.V(0).Info("GROUP_NAME must be specified")
		_ = exit.Exit(1)
		return
	}

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	options := option.NewOptions(os.Stdout, os.Stderr, groupName, webhook)
	cmd := &cobra.Command{
		Short: "Launch an ACME solver API server",
		Long:  "Launch an ACME solver API server",
		RunE: func(c *cobra.Command, args []string) error {
			config, err := options.Config()
			if err != nil {
				return err
			}
			server, err := config.Complete().New()
			if err != nil {
				return err
			}
			err = server.GenericAPIServer.PrepareRun().Run(stopCh)
			if err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	options.RecommendedOptions.AddFlags(flags)
	flags.AddGoFlagSet(flag.CommandLine)

	if err := cmd.Execute(); err != nil {
		logger.Error(err, "error executing command")
		_ = exit.Panic(err)
		return
	}
}
