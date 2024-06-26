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

package log

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/traas-stack/chaosmeta/chaosmetad/pkg/utils"
	"sync"

	"os"
)

var (
	Level  string
	Path   string
	logger *logrus.Logger
	mutex  sync.Mutex
)

const (
	Debug = "debug"
	Info  = "info"
	Warn  = "warn"
	Error = "error"

	TimeFormat = "2006-01-02 15:04:05"
)

func GetLogger(ctx context.Context) *logrus.Entry {
	if logger == nil {
		mutex.Lock()
		if logger == nil {
			setLogger()
		}
		mutex.Unlock()
	}

	traceId := utils.GetTraceId(ctx)
	if traceId != "" {
		return logger.WithFields(logrus.Fields{utils.CtxTraceId: traceId})
	}

	return logger.WithFields(logrus.Fields{})
}

//func WithUid(uid string) *logrus.Entry {
//	return GetLogger().WithFields(logrus.Fields{
//		"uid": uid,
//	})
//}

//func WithTraceId(ctx context.Context) *logrus.Entry {
//	return GetLogger().WithFields(logrus.Fields{
//		utils.CtxTraceId: utils.GetTraceId(ctx),
//	})
//}

func setLogger() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: TimeFormat,
		FullTimestamp:   true,
	})
	logger.SetLevel(getLogLevel(Level))
	if Path == "" {
		logger.SetOutput(os.Stdout)
	} else {
		f, err := getLogPathFile()
		if err != nil {
			// unacceptable exception
			panic(any(fmt.Sprintf("get logger path file error: %s", err.Error())))
		}
		logger.SetOutput(f)
	}
}

func getLogPathFile() (*os.File, error) {
	f, err := os.OpenFile(Path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func getLogLevel(lev string) logrus.Level {
	switch lev {
	case Info:
		return logrus.InfoLevel
	case Debug:
		return logrus.DebugLevel
	case Error:
		return logrus.ErrorLevel
	case Warn:
		return logrus.WarnLevel
	default:
		return logrus.InfoLevel
	}
}
