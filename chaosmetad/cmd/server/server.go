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

package server

import (
	"context"
	"fmt"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/log"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/errutil"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils/process"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/web"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func watchSignal(ctx context.Context) {
	logger := log.GetLogger(ctx)
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGCHLD, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	for s := range c {
		switch s {
		case syscall.SIGCHLD:
			logger.Debugf("receive signal of children process")
			process.WaitDefunctProcess(ctx)
		default:
			// TODO: The atomicity of the task needs to be guaranteed, and the program can really exit without the task in the processing flow
			// TODO: Exiting the server with 'ctrl-c' will cause the tasks in progress to also be exited(but not exited with 'kill' and 'normal terminate')
			errutil.SolveErr(ctx, errutil.NoErr, "server exit")
		}
	}
}

// NewServerCommand serverCmd represents the server command
func NewServerCommand() *cobra.Command {
	var addr, port string
	//var cert, key string
	var isPprof bool
	cmd := &cobra.Command{
		Use:   "server",
		Short: "start up daemon service",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := utils.GetCtxWithTraceId(context.Background(), "system")
			go watchSignal(ctx)

			//if cert != "" && key != "" {
			//	startHTTPSServer(addr, port, isPprof, cert, key)
			//} else {
			startHTTPService(ctx, addr, port, isPprof)
			//}
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", "0.0.0.0", "service bind addr")
	cmd.Flags().StringVarP(&port, "port", "p", "29595", "service bind port")
	cmd.Flags().BoolVar(&isPprof, "enable-pprof", true, "if open pprof service")
	//cmd.Flags().StringVarP(&cert, "cert", "c", "", "path to certificate file")
	//cmd.Flags().StringVarP(&key, "key", "k", "", "path to private key file")
	// HTTPS
	//      --CA string            path to a PEM encoded CA's certificate file
	//      --cert string          path to a PEM encoded certificate file
	//   	--key string           path to a PEM encoded private key file
	return cmd
}

func startHTTPService(ctx context.Context, addr string, port string, isPprof bool) {
	logger := log.GetLogger(ctx)
	logger.Infof("HTTP Service Listen on %s:%s, pprof: %t", addr, port, isPprof)
	router := web.NewRouter(ctx, isPprof)

	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", addr, port), router); err != nil {
		logger.Fatalf("start http service fail: %s", err.Error())
	}
}

//func startHTTPSServer(addr string, port string, isPprof bool, cert, key string) {
//	logger := tools.GetLogger()
//	logger.Infof("HTTPS Service Listen on %s:%s, pprof: %t, cert: %s, key: %s", addr, port, isPprof, cert, key)
//	router := web.NewRouter(isPprof)
//
//	if err := http.ListenAndServeTLS(fmt.Sprintf("%s:%s", addr, port), cert, key, router); err != nil {
//		logger.Fatalf("start https service fail: %s", err.Error())
//	}
//}
