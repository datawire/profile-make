package visualize

import (
	"github.com/pkg/errors"
)

func convertProfile(rawProfile RawProfile) (*SVGProfile, error) {
	makes := convertMakes(RawCommandList(rawProfile.Commands))
	var make *SVGMake
	switch len(makes) {
	case 0:
		// do nothing
	case 1:
		for _, m := range makes {
			make = m
		}
	default:
		return nil, errors.New("CURDIR is inconsistent between top-level commands")
	}

	return &SVGProfile{
		StartTime:  rawProfile.StartTime,
		FinishTime: rawProfile.FinishTime,
		Make:       make,
	}, nil
}

func convertMake(rawCommands RawCommandList) *SVGMake {
	if len(rawCommands) == 0 {
		return nil
	}
	svgMake := &SVGMake{
		Restarts: nil,
	}
	for restartNum, numRestarts := uint(0), rawCommands.CountRestarts(); restartNum <= numRestarts; restartNum++ {
		svgRestart := &SVGRestart{
			Parent: svgMake,
		}
		recipes := make(map[string]*SVGRecipe)
		for _, rawCommand := range rawCommands {
			if rawCommand.MakeRestarts != restartNum {
				continue
			}
			name := rawCommand.RecipeTarget
			if _, exists := recipes[name]; !exists {
				recipes[name] = &SVGRecipe{
					Parent: svgRestart,
					Name:   name,
				}
			}
			cmd := &SVGCommand{
				Parent:   recipes[name],
				Raw:      rawCommand,
				SubMakes: convertMakes(RawCommandList(rawCommand.SubCommands)),
			}
			for _, submake := range cmd.SubMakes {
				submake.Parent = cmd
			}
			recipes[name].Commands = append(recipes[name].Commands, cmd)
		}
		for _, recipe := range recipes {
			svgRestart.Recipes = append(svgRestart.Recipes, recipe)
		}
		svgMake.Restarts = append(svgMake.Restarts, svgRestart)
	}
	return svgMake
}

func convertMakes(rawCommands RawCommandList) map[string]*SVGMake {
	if len(rawCommands) == 0 {
		return nil
	}
	// TODO: maybe look for non-monotinic MakeRestarts?
	sets := make(map[string]RawCommandList)
	for _, cmd := range rawCommands {
		if _, exists := sets[cmd.MakeDir]; !exists {
			sets[cmd.MakeDir] = nil
		}
		sets[cmd.MakeDir] = append(sets[cmd.MakeDir], cmd)
	}
	makes := make(map[string]*SVGMake, len(sets))
	for dir, cmds := range sets {
		makes[dir] = convertMake(cmds)
	}
	return makes
}
