package loader

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/injector/jvm"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/log"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/cmdexec"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/errutil"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/filesys"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/namespace"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/process"
)

const (
	RangerExecutor = "load.sh"
	LoadConfig     = "setConfig"
	Load           = "Load"
	Attach         = "Attach"
	ProcessKey     = "javaagent"
)

type Loader struct {
	Args *Args
}

type LoadModuleArgs struct {
	SetConfig string `json:"setConfig,omitempty"`
	Load      string `json:"load,omitempty"`
}

func NewLoader(args *Args) *Loader {
	loader := &Loader{
		Args: args,
	}
	return loader
}

func NewCmdLoader(args *Args) *cobra.Command {
	loader := NewLoader(args)
	cmd := &cobra.Command{
		Short: fmt.Sprintf("create an load task for %v", args),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := utils.GetCtxWithTraceId(context.Background(), utils.TraceId)
			if len(args) != 0 {
				errutil.SolveErr(ctx, errutil.BadArgsErr, fmt.Sprintf("unknown args: %s, please add -h to get more info", args))
			}
			code, msg := ProcessLoad(ctx, loader)
			errutil.SolveErr(ctx, code, msg)
		},
	}
	return cmd
}

func ProcessLoad(ctx context.Context, loader *Loader) (code int, msg string) {
	if loader.Args.Load != "" {
		err := loader.load(ctx)
		if err != nil {
			fmt.Println(err)
			return errutil.InternalErr, "load fail"
		}
	}
	return 0, "load success"
}

func (i *Loader) SetOption(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&i.Args.Pid, "pid", "p", 0, "target process's pid")
	cmd.Flags().StringVarP(&i.Args.Key, "key", "k", "", "the key used to grep to get target process, the effect is equivalent to \"ps -ef | grep [key]\". if \"pid\" provided, \"key\" will be ignored")
}

func (i *Loader) Action(ctx context.Context, action string) error {
	logger := log.GetLogger(ctx)
	var (
		pidList []int
		err     error
	)

	pidList, err = process.GetPidListByPidOrKeyInContainer(ctx, i.Args.Info.ContainerRuntime, i.Args.Info.ContainerId, i.Args.Pid, ProcessKey)

	if err != nil {
		return fmt.Errorf("get pid list error: %s", err.Error())
	}

	if len(pidList) == 0 {
		return fmt.Errorf("pidList is null, containerId [%s], pid [%s], processKey[%s]", i.Args.Info.ContainerRuntime, i.Args.Pid, ProcessKey)
	}

	dstDir := i.getJVMPackagePath()
	isExist, err := filesys.CheckDir(ctx, i.Args.Info.ContainerRuntime, i.Args.Info.ContainerId, dstDir)
	if err != nil {
		return fmt.Errorf("check dir error: %s", err.Error())
	}
	// 拷贝chaosmeta-jvm包，解压缩
	if !isExist {
		if i.Args.Info.ContainerRuntime != "" {
			src := fmt.Sprintf("%s.tar.gz", utils.GetToolPath(jvm.JVMPackage))
			dst := fmt.Sprintf("%s.tar.gz", dstDir)
			if err := cmdexec.CpContainerFile(ctx, i.Args.Info.ContainerRuntime, i.Args.Info.ContainerId, src, dst); err != nil {
				return fmt.Errorf("cp jvm tool file[%s] to container[%s] [%s] error: %s", src, i.Args.Info.ContainerId, dst, err.Error())
			}
		}

		tarCheckCmd := fmt.Sprintf("export && which tar")
		_, err := cmdexec.ExecCommonWithNS(ctx, i.Args.Info.ContainerRuntime, i.Args.Info.ContainerId, tarCheckCmd, []string{namespace.MNT})
		if err != nil {
			fmt.Println("check tar cmd error:", err.Error())
		}

		tarCmd := fmt.Sprintf("/bin/tar vzxf %s.tar.gz -C %s", dstDir, filesys.GetDirName(dstDir))
		_, err = cmdexec.ExecCommonWithNS(ctx, i.Args.Info.ContainerRuntime, i.Args.Info.ContainerId, tarCmd, []string{namespace.MNT})
		if err != nil {
			return fmt.Errorf("tar JVM tool error: %s", err.Error())
		}
	}
	for _, unitPid := range pidList {
		if action == Attach {
			attachCmd := fmt.Sprintf("%s/%s attach %d", dstDir, RangerExecutor, unitPid)
			_, err := cmdexec.ExecCommonWithNS(ctx, i.Args.Info.ContainerRuntime, i.Args.Info.ContainerId, attachCmd, []string{namespace.MNT, namespace.ENV, namespace.PID, namespace.IPC, namespace.UTS})
			if err != nil {
				logger.Error("load for %d error: %s", unitPid, err.Error())
			}
		} else if action == Load {
			loadCmd := fmt.Sprintf("%s/%s load %d %s", dstDir, RangerExecutor, unitPid, i.Args.Load)
			_, err := cmdexec.ExecCommonWithNS(ctx, i.Args.Info.ContainerRuntime, i.Args.Info.ContainerId, loadCmd, []string{namespace.MNT, namespace.ENV, namespace.PID, namespace.IPC, namespace.UTS})
			if err != nil {
				logger.Error("load for %d error: %s", unitPid, err.Error())
			}
		}
	}
	return err
}

func (i *Loader) attach(ctx context.Context) error {
	return i.Action(ctx, Attach)
}

func (i *Loader) load(ctx context.Context) error {
	return i.Action(ctx, Load)
}

func (i *Loader) getJVMPackagePath() string {
	if i.Args.Info.ContainerRuntime == "" {
		return utils.GetToolPath(jvm.JVMPackage)
	} else {
		return fmt.Sprintf("%s/%s", jvm.ContainerJVMDir, jvm.JVMPackage)
	}
}
