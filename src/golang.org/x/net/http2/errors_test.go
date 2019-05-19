// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http2

import (
	"testing"
	"vendor"
)

func TestErrCodeString(t *testing.T) {
	tests := []struct {
		err  vendor.ErrCode
		want string
	}{
		{vendor.ErrCodeProtocol, "PROTOCOL_ERROR"},
		{0xd, "HTTP_1_1_REQUIRED"},
		{0xf, "unknown error code 0xf"},
	}
	for i, tt := range tests {
		got := tt.err.String()
		if got != tt.want {
			t.Errorf("%d. Error = %q; want %q", i, got, tt.want)
		}
	}
}
