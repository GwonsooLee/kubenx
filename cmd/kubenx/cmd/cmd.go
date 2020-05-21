package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"io"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"os"
)

//Get New Kubenx Command
func NewKubenxCommand(out, err io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kubenx",
		Short: "A brief description of your application",
		Long: `Kubenx is command line tool for kubernetes users with amazon eks.

You can find more information in https://github.com/GwonsooLee/kubenx`,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.Root().SilenceUsage = true

			return nil
		},
	}

	// Group by commands
	groups := templates.CommandGroups{
		{
			Message: "Get Information of kubernetes cluster",
			Commands: []*cobra.Command{
				//NewCmdGet(),
			},
		},
	}

	groups.Add(rootCmd)

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubenx.yaml)")
	rootCmd.PersistentFlags().StringP("region", "r", "ap-northeast-2", "AWS region for service")
	rootCmd.PersistentFlags().StringP("cluster", "c", "", "Name of EKS Cluster")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Get Current Namespace")
	rootCmd.PersistentFlags().StringP("securitygroup", "s", "", "Name of Security Group")
	rootCmd.PersistentFlags().BoolP("all", "A", false, "All Namespace")

	// Viper Binding
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("cluster", rootCmd.PersistentFlags().Lookup("cluster"))
	viper.BindPFlag("securitygroup", rootCmd.PersistentFlags().Lookup("securitygroup"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("all", rootCmd.PersistentFlags().Lookup("all"))


	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

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
