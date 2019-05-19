package main

import (
	"flag"
	"fmt"
	"vendor"
)

func init() {
	const onError = flag.ExitOnError
	vendor.New(
		"compress", "[arguments] [<file name> ...]",
		"Compress the given files or from stdin to stdout.",
		onError, vendor.Compress)
	vendor.New(
		"uncompress", "[arguments] [<file name> ...]",
		"Uncompress the given files or from stdin to stdout.",
		onError, vendor.Uncompress)
}

func main() {
	flag.CommandLine.Bool(vendor.VersionBoolFlag, false, "print the program version")

	err := vendor.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}
}
