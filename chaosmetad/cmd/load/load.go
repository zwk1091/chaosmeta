package load

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/loader"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/errutil"
)

func NewLoadCommand() *cobra.Command {
	loadArgs := &loader.Args{}
	var loadCmd = &cobra.Command{
		Use:   "load",
		Short: "load ranger module  command",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := utils.GetCtxWithTraceId(context.Background(), utils.TraceId)
			if len(args) != 0 {
				errutil.SolveErr(ctx, errutil.BadArgsErr, fmt.Sprintf("unknown args: %s, please add -h to get more info", args))
			}
			loaderCur := loader.NewLoader(loadArgs)
			code, msg := loader.ProcessLoad(ctx, loaderCur)
			errutil.SolveErr(ctx, code, msg)
		},
	}

	loadCmd.PersistentFlags().StringVarP(&loadArgs.Info.Timeout, "timeout", "t", "", "experiment's duration, support unit: \"s、m、h\"(default s)")
	loadCmd.Flags().StringVarP(&loadArgs.Load, "load", "l", "", "load module")
	loadCmd.Flags().StringVarP(&loadArgs.SetConfig, "setConfig", "s", "", "")

	loadCmd.PersistentFlags().StringVar(&loadArgs.Info.Creator, "creator", "", "experiment's creator（default the cmd exec user）")

	loadCmd.PersistentFlags().StringVar(&loadArgs.Info.ContainerRuntime, "container-runtime", "", "if attack a container of local host, need to provide the container runtime of target container")
	loadCmd.PersistentFlags().StringVar(&loadArgs.Info.ContainerId, "container-id", "", "if attack a container of local host, need to provide the container id of target container")

	loadCmd.PersistentFlags().StringVar(&loadArgs.Info.Uid, "uid", "", "if not provide, it will automatically generate an uid")

	return loadCmd
}
