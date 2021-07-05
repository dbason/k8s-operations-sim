// Package opnisim contains the root of the Opni K8s simulator command tree
package opnisim

import (
	"context"
	"os"

	"github.com/dbason/k8s-operations-sim/pkg/opnisim/commands"
	"github.com/rancher/opni/pkg/opnictl/common"
	"github.com/spf13/cobra"
)

func BuildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "opnisim",
	}

	rootCmd.AddCommand(commands.BuildRunCmd())

	return rootCmd
}

func Execute() {
	common.LoadDefaultClientConfig()
	if err := BuildRootCmd().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}
