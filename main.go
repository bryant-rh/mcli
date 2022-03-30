package main

import (
	"os"

	cmd "github.com/bryant-rh/mcli/cmd/mcli"
	"github.com/bryant-rh/mcli/pkg/util"

	"github.com/google/go-containerregistry/pkg/logs"
)

//var imagelist []string

func init() {
	logs.Warn.SetOutput(os.Stderr)
	logs.Progress.SetOutput(os.Stderr)
}

func main() {
	command := cmd.New()
	util.CheckErr(command.Execute())
}

