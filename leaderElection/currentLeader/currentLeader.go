package currentLeader

type SetCurrentLeaderSignal struct {
	LeaderID   string
	ResponseCh chan<- bool
}
type ResetSignal struct{}

type CurrentLeader struct {
	leader              string
	SetCurrentLeaderReq chan SetCurrentLeaderSignal
	ResetReq            chan ResetSignal
}

func NewCurrentLeader() *CurrentLeader {
	currentLeader := &CurrentLeader{
		leader:              "",
		SetCurrentLeaderReq: make(chan SetCurrentLeaderSignal),
		ResetReq:            make(chan ResetSignal),
	}

	go func() {
		for {
			select {
			case signal := <-currentLeader.SetCurrentLeaderReq:
				currentLeader.setCurrentLeader(signal)
			case <-currentLeader.ResetReq:
				currentLeader.leader = ""
			}
		}
	}()

	return currentLeader
}

func (currentLeader *CurrentLeader) setCurrentLeader(signal SetCurrentLeaderSignal) {
	if currentLeader.leader != "" {
		// we altready recognized a leader, check if it is the same as the requested one
		if currentLeader.leader == signal.LeaderID {
			// the leader is already set to the requested one
			signal.ResponseCh <- true
		} else {
			// we already have a leader, but it is not the same as the requested one
			signal.ResponseCh <- false
		}
	} else {
		// we don't have a currentLeader yet, set it
		currentLeader.leader = signal.LeaderID
		signal.ResponseCh <- true
	}
}
