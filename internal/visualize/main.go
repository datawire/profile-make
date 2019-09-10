package visualize

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

func Main(args ...string) error {
	if len(args) > 0 {
		return errors.Errorf("got %d arguments; visualize doesn't take arguments, it reads from stdin and writes to stdout",
			len(args))
	}

	profileBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var profileStructRaw RawCommandList
	err = json.Unmarshal(profileBytes, &profileStructRaw)
	if err != nil {
		return err
	}

	profileStructSVG := convertProfile(profileStructRaw)

	err = mainTemplate.Execute(os.Stdout, map[string]interface{}{
		"ProfileData": profileStructSVG,
	})
	if err != nil {
		return err
	}

	return nil
}
