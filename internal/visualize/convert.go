package visualize

func convertProfile(rawProfile RawCommandList) *SVGProfile {
	ret := &SVGProfile{
		StartTime:  rawProfile.StartTime(),
		FinishTime: rawProfile.FinishTime(),
	}
	ret.Make = convertMake(rawProfile, ret)
	return ret
}

func convertMake(rawProfile RawCommandList, profile *SVGProfile) *SVGMake {
	if len(rawProfile) == 0 {
		return nil
	}
	var ret SVGMake
	for restartNum, numRestarts := uint(0), rawProfile.CountRestarts(); restartNum <= numRestarts; restartNum++ {
		restart := SVGRestart{}
		for _, rawCommand := range rawProfile {
			if rawCommand.MakeRestarts != restartNum {
				continue
			}
			name := rawCommand.RecipeTarget
			if _, exists := restart[name]; !exists {
				restart[name] = &SVGRecipe{
					Name: name,
				}
			}
			restart[name].Commands = append(restart[name].Commands, &SVGCommand{
				X:       XTime{Profile: profile, X: rawCommand.StartTime},
				W:       XDuration{Profile: profile, W: rawCommand.FinishTime.Sub(rawCommand.StartTime)},
				Args:    rawCommand.Args,
				SubMake: convertMake(rawCommand.SubCommands, profile),
			})
		}
		ret = append(ret, &restart)
	}
	return &ret
}
