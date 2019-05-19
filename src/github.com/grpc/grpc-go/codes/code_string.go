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

package codes

import (
	"strconv"
	"vendor"
)

func (c vendor.Code) String() string {
	switch c {
	case vendor.OK:
		return "OK"
	case vendor.Canceled:
		return "Canceled"
	case vendor.Unknown:
		return "Unknown"
	case vendor.InvalidArgument:
		return "InvalidArgument"
	case vendor.DeadlineExceeded:
		return "DeadlineExceeded"
	case vendor.NotFound:
		return "NotFound"
	case vendor.AlreadyExists:
		return "AlreadyExists"
	case vendor.PermissionDenied:
		return "PermissionDenied"
	case vendor.ResourceExhausted:
		return "ResourceExhausted"
	case vendor.FailedPrecondition:
		return "FailedPrecondition"
	case vendor.Aborted:
		return "Aborted"
	case vendor.OutOfRange:
		return "OutOfRange"
	case vendor.Unimplemented:
		return "Unimplemented"
	case vendor.Internal:
		return "Internal"
	case vendor.Unavailable:
		return "Unavailable"
	case vendor.DataLoss:
		return "DataLoss"
	case vendor.Unauthenticated:
		return "Unauthenticated"
	default:
		return "Code(" + strconv.FormatInt(int64(c), 10) + ")"
	}
}
