package main

import (
	"fmt"
	"os"

	"github.com/nii-cloud/nblibram/internal/cells"
	"github.com/nii-cloud/nblibram/internal/filter"
	"github.com/nii-cloud/nblibram/internal/mutate"
	"github.com/nii-cloud/nblibram/internal/outputs"
	"github.com/nii-cloud/nblibram/internal/pkl"
	"github.com/nii-cloud/nblibram/internal/section"
	"github.com/nii-cloud/nblibram/internal/toc"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	// query
	case "toc":
		err = toc.Run(args)
	case "section":
		err = section.Run(args)
	case "cells":
		err = cells.Run(args)
	case "outputs":
		err = outputs.Run(args)
	// mutate
	case "insert":
		err = mutate.RunInsert(args)
	case "update":
		err = mutate.RunUpdate(args)
	case "delete":
		err = mutate.RunDelete(args)
	// filter
	case "filter":
		err = filter.Run(args)
	case "init-config":
		err = filter.InitConfig()
	// pkl
	case "pkl":
		err = pkl.Run(args)
	// utility
	case "hash":
		err = mutate.RunHash(args)
	case "version":
		fmt.Printf("nblibram %s\n", version)
		return
	case "-h", "--help", "help":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "nblibram: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `nblibram <command> [options]

query:
  toc       extract table of contents
  section   extract sections by heading
  cells     extract cell sets
  outputs   extract cell outputs

mutate:
  insert    insert a new cell
  update    update cell content
  delete    delete a cell

filter:
  filter      sanitize sensitive information
  init-config create default filter config

pkl:
  pkl         read pickled kernel output logs

utility:
  hash      compute cell hashes
  version   show version`)
}
