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
	ret := &SVGMake{
		Restarts: nil,
	}
	for restartNum, numRestarts := uint(0), rawCommands.CountRestarts(); restartNum <= numRestarts; restartNum++ {
		recipes := make(map[string]*SVGRecipe)
		for _, rawCommand := range rawCommands {
			if rawCommand.MakeRestarts != restartNum {
				continue
			}
			name := rawCommand.RecipeTarget
			if _, exists := recipes[name]; !exists {
				recipes[name] = &SVGRecipe{
					Name: name,
				}
			}
			recipes[name].Commands = append(recipes[name].Commands, &SVGCommand{
				Raw:      rawCommand,
				SubMakes: convertMakes(RawCommandList(rawCommand.SubCommands)),
			})
		}
		restart := &SVGRestart{
			Recipes: make([]*SVGRecipe, 0, len(recipes)),
		}
		for _, recipe := range recipes {
			restart.Recipes = append(restart.Recipes, recipe)
		}
		ret.Restarts = append(ret.Restarts, restart)
	}
	return ret
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
