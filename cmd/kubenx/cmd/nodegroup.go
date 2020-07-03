package cmd

import (
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
			Red("Too many Arguments")
			os.Exit(1)
		}

		_get_detail_info_of_nodegroup()
	},
	Aliases: []string{"ng"},
}
