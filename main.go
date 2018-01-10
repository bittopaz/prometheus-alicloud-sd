package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tinglinux/prometheus-alicloud-sd/exporter"
)

func main() {
	var filePath string
	var exporterType string
	flag.StringVar(&filePath, "f", "", "Output filename")
	flag.StringVar(&exporterType, "t", "", "exporter type(node/mysql)")
	flag.Parse()

	if filePath == "" {
		fmt.Fprintf(os.Stderr, "required arguments -f must pass in.")
		os.Exit(1)
	}

	if exporterType == "node" {
		exporter.DiscoveryAlicloudNode(filePath, exporterType)
	} else if exporterType == "mysql" {
		exporter.DiscoveryAlicloudMysql(filePath, exporterType)
	} else if exporterType == "" {
		fmt.Fprintf(os.Stderr, "required arguments -t must pass in.")
		os.Exit(1)
	}
}
