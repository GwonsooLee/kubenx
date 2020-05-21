package app

import (
	"context"
	"io"
	"github.com/GwonsooLee/kubenx/cmd/kubenx/cmd"
)

func Run(out, stderr io.Writer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	catchCtrlC(cancel)

	c := cmd.NewKubenxCommand(out, stderr)
	return c.ExecuteContext(ctx)
}
