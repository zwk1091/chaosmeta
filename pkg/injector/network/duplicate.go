/*
 * Copyright 2022-2023 Chaos Meta Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package network

import (
	"context"
	"fmt"
	"github.com/ChaosMetaverse/chaosmetad/pkg/injector"
	"github.com/ChaosMetaverse/chaosmetad/pkg/utils"
	"github.com/ChaosMetaverse/chaosmetad/pkg/utils/cmdexec"
	"github.com/ChaosMetaverse/chaosmetad/pkg/utils/namespace"
	"github.com/ChaosMetaverse/chaosmetad/pkg/utils/net"
	"github.com/spf13/cobra"
)

func init() {
	injector.Register(TargetNetwork, FaultDuplicate, func() injector.IInjector { return &DuplicateInjector{} })
}

type DuplicateInjector struct {
	injector.BaseInjector
	Args    DuplicateArgs
	Runtime DuplicateRuntime
}

type DuplicateArgs struct {
	Interface string `json:"interface"`
	Percent   int    `json:"percent"`
	Direction string `json:"direction"`
	Mode      string `json:"mode"`
	SrcIp     string `json:"src_ip,omitempty"`
	DstIp     string `json:"dst_ip,omitempty"`
	SrcPort   string `json:"src_port,omitempty"`
	DstPort   string `json:"dst_port,omitempty"`
	Force     bool   `json:"force,omitempty"`
}

type DuplicateRuntime struct{}

func (i *DuplicateInjector) GetArgs() interface{} {
	return &i.Args
}

func (i *DuplicateInjector) GetRuntime() interface{} {
	return &i.Runtime
}

func (i *DuplicateInjector) SetDefault() {
	i.BaseInjector.SetDefault()

	if i.Args.Direction == "" {
		i.Args.Direction = DirectionOut
	}

	if i.Args.Mode == "" {
		i.Args.Mode = net.ModeNormal
	}
}

func (i *DuplicateInjector) SetOption(cmd *cobra.Command) {
	// i.BaseInjector.SetOption(cmd)
	cmd.Flags().IntVarP(&i.Args.Percent, "percent", "p", 0, "packets duplicate percent, an integer in (0,100] without \"%\", eg: \"30\" means \"30%\"")

	cmd.Flags().StringVarP(&i.Args.Direction, "direction", "d", "", fmt.Sprintf("flow direction to inject, support: %s（default %s）", DirectionOut, DirectionOut))
	cmd.Flags().StringVarP(&i.Args.Mode, "mode", "m", "", fmt.Sprintf("inject mode, support: %s（default）、%s(means white list mode)", net.ModeNormal, net.ModeExclude))
	cmd.Flags().BoolVarP(&i.Args.Force, "force", "f", false, "force will overwrite the network rule if old rule exist")

	cmd.Flags().StringVarP(&i.Args.Interface, "interface", "i", "", "filter condition: network interface. eg: lo")
	cmd.Flags().StringVar(&i.Args.SrcIp, "src-ip", "", "filter condition: source ip. eg: 10.10.0.0/16,192.168.2.5,192.168.1.0/24")
	cmd.Flags().StringVar(&i.Args.DstIp, "dst-ip", "", "filter condition: destination ip. eg: 10.10.0.0/16,192.168.2.5,192.168.1.0/24")
	cmd.Flags().StringVar(&i.Args.SrcPort, "src-port", "", "filter condition: source port. eg: 8080,9090,12000/8")
	cmd.Flags().StringVar(&i.Args.DstPort, "dst-port", "", "filter condition: destination port. eg: 8080,9090,12000/8")
}

func (i *DuplicateInjector) getCmdExecutor(method, args string) *cmdexec.CmdExecutor {
	return &cmdexec.CmdExecutor{
		ContainerId:      i.Info.ContainerId,
		ContainerRuntime: i.Info.ContainerRuntime,
		ContainerNs:      []string{namespace.NET},
		ToolKey:          NetworkExec,
		Method:           method,
		Fault:            FaultDuplicate,
		Args:             args,
	}
}

// Validator Only one tc network failure can be executed at the same time
func (i *DuplicateInjector) Validator(ctx context.Context) error {
	if err := i.BaseInjector.Validator(ctx); err != nil {
		return err
	}
	return i.getCmdExecutor(utils.MethodValidator, fmt.Sprintf("'%s' '%s' '%s' '%s' '%s' '%s' '%s' %v %d",
		i.Args.Interface, i.Args.Direction, i.Args.Mode, i.Args.SrcIp, i.Args.DstIp, i.Args.SrcPort, i.Args.DstPort,
		i.Args.Force, i.Args.Percent)).ExecTool(ctx)
}

func (i *DuplicateInjector) Inject(ctx context.Context) error {
	return i.getCmdExecutor(utils.MethodInject, fmt.Sprintf("'%s' '%s' '%s' '%s' '%s' '%s' %v %d",
		i.Args.Interface, i.Args.Mode, i.Args.SrcIp, i.Args.DstIp, i.Args.SrcPort, i.Args.DstPort,
		i.Args.Force, i.Args.Percent)).ExecTool(ctx)
}

func (i *DuplicateInjector) Recover(ctx context.Context) error {
	if i.BaseInjector.Recover(ctx) == nil {
		return nil
	}
	return i.getCmdExecutor(utils.MethodRecover, i.Args.Interface).ExecTool(ctx)
}
