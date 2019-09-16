package visualize

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

func inArray(needle string, haystack []string) bool {
	for _, straw := range haystack {
		if straw == needle {
			return true
		}
	}
	return false
}

func Main(args ...string) error {
	layouts := []string{
		"wallclock",
		"compact",
	}
	argparser := pflag.NewFlagSet("visualize", pflag.ContinueOnError)
	var (
		argLayout       = argparser.String("layout", "wallclock", fmt.Sprintf("Layout algorithm to use; one of [%v]", layouts))
		argShowOverflow = argparser.Bool("show-overflow", false, "Display overflowed text")
	)
	err := argparser.Parse(args)
	if err != nil {
		return err
	}
	if !inArray(*argLayout, layouts) {
		return errors.Errorf("invalid --layout: %q", *argLayout)
	}
	if argCnt := len(argparser.Args()); argCnt > 0 {
		return errors.Errorf("got %d positional arguments; visualize doesn't take positional arguments", argCnt)
	}

	profileBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var profileStructRaw RawProfile
	err = json.Unmarshal(profileBytes, &profileStructRaw)
	if err != nil {
		return err
	}

	profileStructSVG, err := convertProfile(profileStructRaw)
	if err != nil {
		return err
	}

	if err = profileStructSVG.SVG(os.Stdout, *argLayout, *argShowOverflow); err != nil {
		return err
	}

	return nil
}
