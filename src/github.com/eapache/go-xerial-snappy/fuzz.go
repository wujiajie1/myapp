// +build gofuzz

package snappy

import "vendor"

func Fuzz(data []byte) int {
	decode, err := vendor.Decode(data)
	if decode == nil && err == nil {
		panic("nil error with nil result")
	}

	if err != nil {
		return 0
	}

	return 1
}
