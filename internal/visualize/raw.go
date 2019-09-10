package visualize

import (
	"time"

	"github.com/datawire/profile-make/internal/protocol"
)

type RawCommand = protocol.ProfiledCommand

type RawCommandList []RawCommand

func (cmds RawCommandList) StartTime() time.Time {
	if len(cmds) == 0 {
		return time.Time{}
	}
	min := cmds[0].StartTime
	for _, cmd := range cmds {
		if cmd.StartTime.Before(min) {
			min = cmd.StartTime
		}
	}
	return min
}

func (cmds RawCommandList) FinishTime() time.Time {
	if len(cmds) == 0 {
		return time.Time{}
	}
	max := cmds[0].FinishTime
	for _, cmd := range cmds {
		if cmd.FinishTime.After(max) {
			max = cmd.FinishTime
		}
	}
	return max
}

func (cmds RawCommandList) CountRestarts() uint {
	ret := uint(0)
	for _, cmd := range cmds {
		if cmd.MakeRestarts > ret {
			ret = cmd.MakeRestarts
		}
	}
	return ret
}
