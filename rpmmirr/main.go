package main

import (
	"fmt"
	"github.com/jmahler/rpmmirr"
	"github.com/pborman/getopt/v2"
	"os"
)

func main() {

	var (
		help   bool
		config string
		rpm    string
		path   string
	)

	getopt.Flag(&help, 'h', "help")
	getopt.Flag(&config, 'c', "config file (e.g. config.json)")
	getopt.Flag(&rpm, 'r', "rpm name (e.g. patch)")
	getopt.Flag(&path, 'C', "path to git repo for rpm")
	getopt.Parse()

	if help {
		getopt.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	_, err := rpmmirr.LoadConfig(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
