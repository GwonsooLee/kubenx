package cmd

import (
	"github.com/GwonsooLee/kubenx/pkg/runner"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

// ngCmd represents the ng command
var ngCmd = &cobra.Command{
	Use:   "nodegroup",
	Short: "Command about node group",
	Long:  `Command about node group`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			utils.Red("Too many Arguments")
			os.Exit(1)
		}

		runner.GetDetailInfoOfNodegroup()
	},
	Aliases: []string{"ng"},
}
