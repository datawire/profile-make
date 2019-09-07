package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"text/template"

	"github.com/pkg/errors"

	"github.com/datawire/profile-make/internal/runmake"
	"github.com/datawire/profile-make/internal/runshell"
)

var usageTmpl = template.Must(template.
	New("--help").
	Parse(`Usage: {{ .Arg0 }} run --output-file=FILE -- [MAKE_ARGS]
   or: {{ .Arg0 }} visualize
   or: {{ .Arg0 }} help
Run GNU Make under a profiler.
`))

func usage() {
	usageTmpl.Execute(os.Stdout, map[string]interface{}{
		"Arg0": os.Args[0],
	})
}

func errusage(err error) {
	fmt.Fprintf(os.Stderr, "profile-make: error: %v\nSee '%s help' for more information.\n", err, os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		errusage(errors.New("expected a sub-command"))
	}
	var err error
	switch os.Args[1] {
	case "help":
		usage()
		return
	case "run":
		err = runmake.Main(os.Args[2:]...)
	case "shell":
		err = runshell.Main(os.Args[2:]...)
	case "visualize":
		panic("TODO: visualize not yet implemented")
	default:
		errusage(errors.Errorf("unrecognized sub-command: %q", os.Args[1]))
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			status := ee.Sys().(syscall.WaitStatus)
			switch {
			case status.Exited():
				os.Exit(status.ExitStatus())
			case status.Signaled():
				// POSIX shells use 128+SIGNAL for the exit
				// code when the process is killed by a
				// signal.
				os.Exit(128 + int(status.Signal()))
			default:
				panic("should not happen")
			}
		}
		fmt.Fprintln(os.Stderr, "profile-make:", "error:", err)
		os.Exit(127)
	}
}
