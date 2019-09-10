package visualize

func convertProfile(rawProfile RawCommandList) *SVGProfile {
	ret := &SVGProfile{
		StartTime:  rawProfile.StartTime(),
		FinishTime: rawProfile.FinishTime(),
	}
	for _, rawCommand := range rawProfile {
		ret.Commands = append(ret.Commands, convertCommand(rawCommand, ret))
	}
	return ret
}

func convertCommand(rawCommand RawCommand, profile *SVGProfile) SVGCommand {
	ret := SVGCommand{
		X:    XTime{Profile: profile, X: rawCommand.StartTime},
		W:    XDuration{Profile: profile, W: rawCommand.FinishTime.Sub(rawCommand.StartTime)},
		Args: rawCommand.Args,
	}
	for _, rawSubCommand := range rawCommand.SubCommands {
		ret.SubCommands = append(ret.SubCommands, convertCommand(rawSubCommand, profile))
	}
	return ret
}
