package cmd

import (
	"context"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"k8s.io/kubectl/pkg/util/templates"
	"os"
)

var (
	cfgFile string
)

//Get New Kubenx Command
func NewKubenxCommand(out, err io.Writer) *cobra.Command {
	cobra.OnInitialize(initConfig)
	rootCmd := &cobra.Command{
		Use:   "kubenx",
		Short: "A brief description of your application",
		Long: `Kubenx is command line tool for kubernetes users with amazon eks.

You can find more information in https://github.com/GwonsooLee/kubenx`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	//Group by commands
	groups := templates.CommandGroups{
		{
			Message: "Get Information of kubernetes cluster",
			Commands: []*cobra.Command{
				NewCmdGet(),
				NewCmdSearch(),
				NewCmdInspect(),
			},
		},
		{
			Message: "manager EKS Cluster",
			Commands: []*cobra.Command{
				NewCmdCluster(),
				NewCmdConfig(),
			},
		},
	}

	groups.Add(rootCmd)

	rootCmd.AddCommand(NewCmdPortForward())
	rootCmd.AddCommand(NewCmdNamespace())
	rootCmd.AddCommand(NewCmdContext())
	rootCmd.AddCommand(NewCmdCompletion())
	rootCmd.AddCommand(NewCmdVersion())

	templates.ActsAsRootCommand(rootCmd, nil, groups...)

	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kubenx" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kubenx")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func alwaysSucceedWhenCancelled(ctx context.Context, err error) error {
	// if the context was cancelled act as if all is well
	if err != nil && ctx.Err() == context.Canceled {
		return nil
	}
	return err
}
