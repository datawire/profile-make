package runmake

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/datawire/profile-make/internal/protocol"
	"github.com/datawire/profile-make/internal/runshell"
)

type stderrLogger struct{}

func (stderrLogger) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintln(os.Stderr, "profile-make:", "make:", fmt.Sprintf(format, a...))
}

func Main(args ...string) error {
	argparser := pflag.NewFlagSet("run", pflag.ContinueOnError)
	var (
		argOutputFile = argparser.String("output-file", "", "Filename to write profiler results to")
	)
	err := argparser.Parse(args)
	if err != nil {
		return err
	}
	cmdline := argparser.Args()

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	tmpdir, err := ioutil.TempDir("", "profile-make.")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	tmpdir, err = filepath.Abs(tmpdir)
	if err != nil {
		return err
	}
	listenerName := filepath.Join(tmpdir, "socket")

	startTime := time.Now()
	var cmdErr error
	cmds, err := protocol.WithServer(listenerName, stderrLogger{}, func() {

		shell := runshell.GetProfilingShell(exe, listenerName)

		cmdline := append(cmdline, "SHELL="+shell)
		cmd := exec.Command(cmdline[0], cmdline[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmdErr = cmd.Run()
	})
	if cmdErr != nil {
		if _, ok := cmdErr.(*exec.Error); ok {
			prefix := os.Args[:len(os.Args)-len(cmdline)]
			suffix := os.Args[len(prefix):]
			return errors.Errorf("%v\n"+
				"  You wrote:\n"+
				"      %s\n"+
				"  Did you mean to write:\n"+
				"      %s make %s\n"+
				"      %*s ^^^^",
				cmdErr,
				strings.Join(os.Args, " "),
				strings.Join(prefix, " "),
				strings.Join(suffix, " "),
				len(strings.Join(prefix, " ")), "",
			)
		}
	}
	if err != nil {
		return err
	}
	finishTime := time.Now()

	file, err := os.OpenFile(*argOutputFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	profile := protocol.Profile{
		StartTime:  startTime,
		FinishTime: finishTime,
		Commands:   cmds,
	}
	if err := json.NewEncoder(file).Encode(profile); err != nil {
		return err
	}

	return cmdErr
}
