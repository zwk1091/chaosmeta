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

package injector

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/crclient"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/log"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/storage"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/cmdexec"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/errutil"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/user"
	"runtime/debug"
	"strings"
)

type IInjector interface {
	SetCommonArgs(info *BaseInfo)
	OptionToExp(args, r interface{}) (*storage.Experiment, error)
	LoadInjector(exp *storage.Experiment, argsVar, rVar interface{}) error
	DelayRecover(ctx context.Context, timeout int64) error

	GetArgs() interface{}
	GetRuntime() interface{}

	SetOption(cmd *cobra.Command)
	SetDefault()
	Validator(ctx context.Context) error
	Inject(ctx context.Context) error
	Recover(ctx context.Context) error
}

/*=======================================Base Injector===================================================*/

type BaseInjector struct {
	Info BaseInfo
}

type BaseInfo struct {
	// basic information
	Uid     string `json:"uid"`
	Creator string `json:"creator"`
	// state information
	Status  string `json:"status"`
	Error   string `json:"error"`
	Timeout string `json:"timeout"`
	// configuration information
	Target string `json:"target"`
	Fault  string `json:"fault"`
	// container information
	ContainerId      string `json:"container_id"`
	ContainerRuntime string `json:"container_runtime"`
	//ContainerNs      []string `json:"container_ns"`
}

func (i *BaseInjector) GetArgs() interface{} {
	var empty interface{}
	return empty
}

func (i *BaseInjector) GetRuntime() interface{} {
	var empty interface{}
	return empty
}

func (i *BaseInjector) SetCommonArgs(info *BaseInfo) {
	if info == nil {
		return
	}

	if info.Uid != "" {
		i.Info.Uid = info.Uid
	}

	if info.Target != "" {
		i.Info.Target = info.Target
	}

	if info.Fault != "" {
		i.Info.Fault = info.Fault
	}

	if info.Creator != "" {
		i.Info.Creator = info.Creator
	}

	if info.Timeout != "" {
		i.Info.Timeout = info.Timeout
	}

	if info.ContainerRuntime != "" {
		i.Info.ContainerRuntime = info.ContainerRuntime
	}

	if info.ContainerId != "" {
		i.Info.ContainerId = info.ContainerId
	}
}

func (i *BaseInjector) SetOption(cmd *cobra.Command) {
	//cmd.Flags().StringVarP(&i.Info.Timeout, "timeout", "t", "", "experiment's duration（default need to stop manually）")
	//cmd.Flags().StringVar(&i.Info.Creator, "creator", "", "experiment's creator（default the cmd exec user）")
}

func (i *BaseInjector) Inject(ctx context.Context) error {
	//i.Info.Status = core.StatusSuccess
	return fmt.Errorf("not implemented")
}

func (i *BaseInjector) Recover(ctx context.Context) error {
	if i.Info.Status == utils.StatusDestroyed || i.Info.Status == utils.StatusError {
		return nil
	}

	return fmt.Errorf("not implemented")
}

func (i *BaseInjector) SetDefault() {
	if i.Info.Creator == "" {
		i.Info.Creator = user.GetUser()
	}

	if i.Info.Uid == "" {
		i.Info.Uid = utils.NewUid()
	}

	if i.Info.Status == "" {
		i.Info.Status = utils.StatusCreated
	}

	if i.Info.ContainerId != "" && i.Info.ContainerRuntime == "" {
		i.Info.ContainerRuntime = crclient.CrDocker
	}
}

func (i *BaseInjector) Validator(ctx context.Context) error {
	if i.Info.ContainerRuntime != "" {
		if i.Info.ContainerId == "" {
			return fmt.Errorf("\"container-id\" is empty")
		}

		client, err := crclient.GetClient(ctx, i.Info.ContainerRuntime)
		if err != nil {
			return fmt.Errorf("create container runtime client[%s] error: %s", i.Info.ContainerRuntime, err.Error())
		}

		if _, err := client.GetPidById(ctx, i.Info.ContainerId); err != nil {
			return fmt.Errorf("check container error: %s", err.Error())
		}
	}

	if err := utils.IsValidUid(i.Info.Uid); err != nil {
		return fmt.Errorf("\"uid\" format error: %s", err.Error())
	}

	if i.Info.Timeout == "" {
		return nil
	}

	if _, err := utils.GetTimeSecond(i.Info.Timeout); err != nil {
		return fmt.Errorf("\"timeout\" is not valid: %s", err.Error())
	}

	return nil
}

func (i *BaseInjector) DelayRecover(ctx context.Context, timeout int64) error {
	return cmdexec.StartSleepRecover(ctx, timeout, i.Info.Uid)
}

func (i *BaseInjector) LoadInjector(exp *storage.Experiment, argsVar, rVar interface{}) error {
	if err := json.Unmarshal([]byte(exp.Args), argsVar); err != nil {
		return fmt.Errorf("load args from experiment error: %s", err.Error())
	}

	if err := json.Unmarshal([]byte(exp.Runtime), rVar); err != nil {
		return fmt.Errorf("load runtime from experiment error: %s", err.Error())
	}

	i.Info.Uid = exp.Uid
	i.Info.Target = exp.Target
	i.Info.Fault = exp.Fault
	i.Info.Status = exp.Status
	i.Info.Error = exp.Error
	i.Info.Creator = exp.Creator
	i.Info.Timeout = exp.Timeout
	i.Info.ContainerRuntime = exp.ContainerRuntime
	i.Info.ContainerId = exp.ContainerId

	return nil
}

func (i *BaseInjector) OptionToExp(args, r interface{}) (*storage.Experiment, error) {
	argsByte, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("args convert to string error: %s", err.Error())
	}
	runtimeByte, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("runtime convert to string error: %s", err.Error())
	}

	exp := &storage.Experiment{
		Uid:              i.Info.Uid,
		Target:           i.Info.Target,
		Fault:            i.Info.Fault,
		Status:           i.Info.Status,
		Creator:          i.Info.Creator,
		Timeout:          i.Info.Timeout,
		Error:            i.Info.Error,
		Args:             string(argsByte),
		Runtime:          string(runtimeByte),
		ContainerRuntime: i.Info.ContainerRuntime,
		ContainerId:      i.Info.ContainerId,
	}

	return exp, nil
}

/*=======================================Main Process===================================================*/

func ProcessInject(ctx context.Context, i IInjector) (code int, msg string) {
	logger := log.GetLogger(ctx)
	defer func() {
		if err := recover(); err != any(nil) {
			logger.Debug(string(debug.Stack()))
			code, msg = errutil.UnknownErr, fmt.Sprintf("ProcessInject Exception: %v", err)
		}
	}()

	i.SetDefault()

	if err := i.Validator(ctx); err != nil {
		return errutil.BadArgsErr, fmt.Sprintf("args error: %s", err.Error())
	}

	db, err := storage.GetExperimentStore()
	if err != nil {
		return errutil.DBErr, fmt.Sprintf("connect db error: %s", err.Error())
	}

	exp, err := i.OptionToExp(i.GetArgs(), i.GetRuntime())
	if err != nil {
		return errutil.BadArgsErr, fmt.Sprintf("create experiment error: %s", err.Error())
	}

	if err := db.Insert(exp); err != nil {
		return errutil.DBErr, fmt.Sprintf("insert new experiment error: %s", err.Error())
	}

	logger.Infof("uid: %s", exp.Uid)
	logger.Infof("args: %s", exp.Args)

	if err := i.Inject(ctx); err != nil {
		errMsg := fmt.Sprintf("inject error: %s", err.Error())
		if err := db.UpdateStatusAndErr(exp.Uid, utils.StatusError, errMsg); err != nil {
			logger.Warnf("update status[%s] for experiment[%s] error: %s", utils.StatusError, exp.Uid, errMsg)
		}

		return errutil.InjectErr, errMsg
	}

	exp, _ = i.OptionToExp(i.GetArgs(), i.GetRuntime())
	exp.Status = utils.StatusSuccess
	if err := db.Update(exp); err != nil {
		// update fails, runtime will be lost, so it must roll back
		if err := i.Recover(ctx); err != nil {
			logger.Warnf("recover error: %s", err.Error())
		}
		return errutil.DBErr, fmt.Sprintf("update status[%s] for experiment[%s] error: %s", exp.Status, exp.Uid, err.Error())
	}

	logger.Info("inject success")

	if exp.Timeout != "" {
		timeSecond, _ := utils.GetTimeSecond(exp.Timeout)
		if err := i.DelayRecover(ctx, timeSecond); err != nil {
			logger.Warnf("inject success but auto delay recover cmd exec error: %s, please execute [chaosmetad recover -u %s] manually to recover", err.Error(), exp.Uid)
		}
	}

	return errutil.NoErr, "success"
}

func ProcessRecover(ctx context.Context, uid string) (code int, msg string) {
	logger := log.GetLogger(ctx)

	defer func() {
		if err := recover(); err != any(nil) {
			logger.Debug(string(debug.Stack()))
			code, msg = errutil.UnknownErr, fmt.Sprintf("ProcessRecover Exception: %v", err)
		}
	}()

	logger.Debugf("uid: %s", uid)

	db, err := storage.GetExperimentStore()
	if err != nil {
		return errutil.DBErr, fmt.Sprintf("connect db error: %s", err.Error())
	}

	exp, err := db.GetByUid(uid)
	if err != nil {
		return errutil.DBErr, fmt.Sprintf("query experiment by uid[%s] error: %s", uid, err.Error())
	}

	i, err := NewInjector(exp.Target, exp.Fault)
	if err != nil {
		return errutil.InternalErr, fmt.Sprintf("find injector by target[%s] and fault[%s] error: %s", exp.Target, exp.Fault, err.Error())
	}

	if err := i.LoadInjector(exp, i.GetArgs(), i.GetRuntime()); err != nil {
		return errutil.InternalErr, fmt.Sprintf("load experiment to injector error: %s", err.Error())
	}

	if err := i.Recover(ctx); err != nil {
		return errutil.RecoverErr, fmt.Sprintf("recover error: %s", err.Error())
	}

	logger.Info("recover success")

	if err := db.UpdateStatus(uid, utils.StatusDestroyed); err != nil {
		logger.Warnf("update status[%s] for experiment[%s] error: %s", utils.StatusDestroyed, uid, err.Error())
	}

	return errutil.NoErr, "success"
}

/*=======================================Command Constructor===================================================*/

func NewCmdByTarget(target string, args *BaseInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   target,
		Short: fmt.Sprintf("create an experiment for %s", target),
	}

	faults := GetFaultsByTarget(target)
	for _, fault := range faults {
		cmd.AddCommand(NewCmdByTargetAndFault(target, fault, args))
	}

	return cmd
}

func NewCmdByTargetAndFault(target, fault string, infoArgs *BaseInfo) *cobra.Command {
	i, _ := NewInjector(target, fault)
	var cmd = &cobra.Command{
		Use:   fault,
		Short: fmt.Sprintf("create %s experiment for %s", fault, target),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := utils.GetCtxWithTraceId(context.Background(), utils.TraceId)
			if len(args) != 0 {
				errutil.SolveErr(ctx, errutil.BadArgsErr, fmt.Sprintf("unknown args: %s, please add -h to get more info", args))
			}

			i.SetCommonArgs(infoArgs)
			code, msg := ProcessInject(ctx, i)
			errutil.SolveErr(ctx, code, msg)
		},
	}

	i.SetOption(cmd)

	return cmd
}

/*=======================================ConstructorScheme Function===================================================*/

var constructorScheme = map[string]func() IInjector{}

func Register(target, fault string, f func() IInjector) {
	constructorScheme[getInjectorKey(target, fault)] = f
}

func NewInjector(target, fault string) (IInjector, error) {
	f := constructorScheme[getInjectorKey(target, fault)]
	if f == nil {
		return nil, fmt.Errorf("no such injector")
	}

	i := f()
	i.SetCommonArgs(&BaseInfo{
		Target: target,
		Fault:  fault,
	})

	return i, nil
}

func getInjectorKey(target, fault string) string {
	return fmt.Sprintf("%s%s%s", target, utils.BuilderSplit, fault)
}

func GetTargets() []string {
	set := make(map[string]bool)
	targets := make([]string, 0)
	for k := range constructorScheme {
		kArr := strings.Split(k, utils.BuilderSplit)
		if !set[kArr[0]] {
			set[kArr[0]] = true
			targets = append(targets, kArr[0])
		}
	}

	return targets
}

func GetFaultsByTarget(target string) []string {
	faults := make([]string, 0)
	for k := range constructorScheme {
		kArr := strings.Split(k, utils.BuilderSplit)
		if kArr[0] == target {
			faults = append(faults, kArr[1])
		}
	}

	return faults
}
