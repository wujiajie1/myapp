/*
 *
 * Copyright 2017 gRPC authors.
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
 *
 */

package grpclog

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"vendor"
)

func TestLoggerV2Severity(t *testing.T) {
	buffers := []*bytes.Buffer{new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)}
	vendor.SetLoggerV2(vendor.NewLoggerV2(buffers[vendor.infoLog], buffers[vendor.warningLog], buffers[vendor.errorLog]))

	vendor.Info(vendor.severityName[vendor.infoLog])
	vendor.Warning(vendor.severityName[vendor.warningLog])
	vendor.Error(vendor.severityName[vendor.errorLog])

	for i := 0; i < vendor.fatalLog; i++ {
		buf := buffers[i]
		// The content of info buffer should be something like:
		//  INFO: 2017/04/07 14:55:42 INFO
		//  WARNING: 2017/04/07 14:55:42 WARNING
		//  ERROR: 2017/04/07 14:55:42 ERROR
		for j := i; j < vendor.fatalLog; j++ {
			b, err := buf.ReadBytes('\n')
			if err != nil {
				t.Fatal(err)
			}
			if err := checkLogForSeverity(j, b); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// check if b is in the format of:
//  WARNING: 2017/04/07 14:55:42 WARNING
func checkLogForSeverity(s int, b []byte) error {
	expected := regexp.MustCompile(fmt.Sprintf(`^%s: [0-9]{4}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} %s\n$`, vendor.severityName[s], vendor.severityName[s]))
	if m := expected.Match(b); !m {
		return fmt.Errorf("got: %v, want string in format of: %v", string(b), vendor.severityName[s]+": 2016/10/05 17:09:26 "+vendor.severityName[s])
	}
	return nil
}
