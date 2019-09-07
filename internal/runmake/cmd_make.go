package runmake

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

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
	makeArgs := argparser.Args()

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

	var cmdErr error
	cmds, err := protocol.WithServer(listenerName, stderrLogger{}, func() {

		shell := runshell.GetProfilingShell(exe, listenerName)

		cmdline := append(append([]string{"make"}, makeArgs...), "SHELL="+shell)
		cmd := exec.Command(cmdline[0], cmdline[1:]...)
		cmd.Env = append(os.Environ(),
			"MAKEFILES="+filepath.Join(tmpdir, "stub.mk"),
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmdErr = cmd.Run()
	})
	if err != nil {
		return err
	}

	file, err := os.OpenFile(*argOutputFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(cmds); err != nil {
		return err
	}

	return cmdErr
}
