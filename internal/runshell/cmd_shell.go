package runshell

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/datawire/profile-make/internal/protocol"
)

func GetProfilingShell(exe, socketName string) string {
	args := []string{
		fmt.Sprintf(`%s %s --profile.socket=%s`, exe, "shell", socketName),
		`--make.level=$(MAKELEVEL)`,
		`--make.restarts=$(MAKE_RESTARTS)`,
		`--make.dir=$(CURDIR)`,
		`--recipe.target=$(abspath $@)`,
		`$(addprefix --recipe.dependency=,$(abspath $^))`,
		`--`,
		`$(or $(profile-make.SHELL),/bin/sh)`,
	}
	return strings.Join(args, " ")
}

type stderrLogger struct{}

func (stderrLogger) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintln(os.Stderr, "profile-make:", "shell:", fmt.Sprintf(format, a...))
}

func Main(args ...string) error {
	startTime := time.Now() // do this as early as possible

	// 0: parse arguments //////////////////////////////////////////////////
	argparser := pflag.NewFlagSet("shell", pflag.ContinueOnError)
	var (
		argProfileSocket      = argparser.String("profile.socket", "", "Socket of parent profile-make server")
		argMakeLevel          = argparser.Uint("make.level", 0, "$(MAKELEVEL)")
		argMakeRestarts       = argparser.Uint("make.restarts", 0, "$(MAKE_RESTARTS)")
		argMakeDir            = argparser.String("make.dir", "", "$(CURDIR)")
		argRecipeTarget       = argparser.String("recipe.target", "", "$@")
		argRecipeDependencies = argparser.StringArray("recipe.dependency", nil, "$^")
	)
	err := argparser.Parse(args)
	if err != nil {
		return err
	}
	cmdline := argparser.Args()

	// 1: connect to parent ////////////////////////////////////////////////
	// do this as early as possible
	conn, connErr := net.Dial("unix", *argProfileSocket)
	if connErr != nil {
		return connErr
	}
	defer conn.Close()

	// 2: run the command //////////////////////////////////////////////////

	socketDir, socketNotdir := filepath.Split(*argProfileSocket)
	tmpdir, err := ioutil.TempDir(socketDir, socketNotdir+".")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	listenerName := filepath.Join(tmpdir, "socket")

	var cmdErr error
	subCmds, err := protocol.WithServer(listenerName, stderrLogger{}, func() {
		cmd := exec.Command(cmdline[0], cmdline[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		for i := range cmd.Args {
			cmd.Args[i] = strings.ReplaceAll(cmd.Args[i], *argProfileSocket, listenerName)
		}
		for i := range cmd.Env {
			cmd.Env[i] = strings.ReplaceAll(cmd.Env[i], *argProfileSocket, listenerName)
		}

		cmdErr = cmd.Run()
	})
	if err != nil {
		return err
	}

	// 3: report to the parent /////////////////////////////////////////////

	finishTime := time.Now() // do this as late as possible

	err = json.NewEncoder(conn).Encode(protocol.ProfiledCommand{
		StartTime:  startTime,
		FinishTime: finishTime,

		MakeLevel:    *argMakeLevel,
		MakeRestarts: *argMakeRestarts,
		MakeDir:      *argMakeDir,

		RecipeTarget:       *argRecipeTarget,
		RecipeDependencies: *argRecipeDependencies,

		Args: cmdline,

		SubCommands: subCmds,
	})
	if err != nil {
		return err
	}

	// 4: exit /////////////////////////////////////////////////////////////
	return cmdErr
}
