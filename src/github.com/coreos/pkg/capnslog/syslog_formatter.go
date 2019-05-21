// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// +build !windows

package capnslog

import (
	"fmt"
	"github.com/coreos"
	"log/syslog"
)

func NewSyslogFormatter(w *syslog.Writer) coreos.Formatter {
	return &syslogFormatter{w}
}

func NewDefaultSyslogFormatter(tag string) (coreos.Formatter, error) {
	w, err := syslog.New(syslog.LOG_DEBUG, tag)
	if err != nil {
		return nil, err
	}
	return NewSyslogFormatter(w), nil
}

type syslogFormatter struct {
	w *syslog.Writer
}

func (s *syslogFormatter) Format(pkg string, l coreos.LogLevel, _ int, entries ...interface{}) {
	for _, entry := range entries {
		str := fmt.Sprint(entry)
		switch l {
		case coreos.CRITICAL:
			s.w.Crit(str)
		case coreos.ERROR:
			s.w.Err(str)
		case coreos.WARNING:
			s.w.Warning(str)
		case coreos.NOTICE:
			s.w.Notice(str)
		case coreos.INFO:
			s.w.Info(str)
		case coreos.DEBUG:
			s.w.Debug(str)
		case coreos.TRACE:
			s.w.Debug(str)
		default:
			panic("Unhandled loglevel")
		}
	}
}

func (s *syslogFormatter) Flush() {
}
