/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubenx",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}
//
var (
	//STATIS VALUE
	KUBENX_HOMEDIR = ".kubenx"
	SSH_DEFAULT_PATH = "ssh"
	TARGET_DEFAULT_PORT = "22"


	//Color Definition
	Red    = color.New(color.FgRed).PrintlnFunc()
	Blue   = color.New(color.FgBlue).PrintlnFunc()
	Green  = color.New(color.FgGreen).PrintlnFunc()
	Yellow = color.New(color.FgYellow).PrintlnFunc()
	Cyan   = color.New(color.FgCyan).PrintlnFunc()

	//OPEN_ID_CA_FINGERPRINT
	CA_FINGERPRINT = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"

	//Error Message
	NO_FILE_EXCEPTION = "No file exists... Please check the file path"
)
//
//// Execute adds all child commands to the root command and sets flags appropriately.
//// This is called by main.main(). It only needs to happen once to the rootCmd.
//func Execute() {
//	if err := rootCmd.Execute(); err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//}
//
//func init() {
//	cobra.OnInitialize(initConfig)
//
//	// Here you will define your flags and configuration settings.
//	// Cobra supports persistent flags, which, if defined here,
//	// will be global for your application.
//	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubenx.yaml)")
//	rootCmd.PersistentFlags().StringP("region", "r", "ap-northeast-2", "AWS region for service")
//	rootCmd.PersistentFlags().StringP("cluster", "c", "", "Name of EKS Cluster")
//	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Get Current Namespace")
//	rootCmd.PersistentFlags().StringP("securitygroup", "s", "", "Name of Security Group")
//	rootCmd.PersistentFlags().BoolP("all", "A", false, "All Namespace")
//
//	// Viper Binding
//	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
//	viper.BindPFlag("cluster", rootCmd.PersistentFlags().Lookup("cluster"))
//	viper.BindPFlag("securitygroup", rootCmd.PersistentFlags().Lookup("securitygroup"))
//	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
//	viper.BindPFlag("all", rootCmd.PersistentFlags().Lookup("all"))
//
//
//	// Cobra also supports local flags, which will only run
//	// when this action is called directly.
//	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
//}
