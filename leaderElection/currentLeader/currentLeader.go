package currentLeader

type SetCurrentLeaderSignal struct {
	Leader     string
	Term       int
	ResponseCh chan<- string
}
type ResetSignal struct{}

type CurrentLeader struct {
	leader              string
	term                int
	SetCurrentLeaderReq chan SetCurrentLeaderSignal
	ResetReq            chan ResetSignal
}

func NewCurrentLeader() *CurrentLeader {
	currentLeader := &CurrentLeader{
		leader:              "",
		term:                0,
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
	if signal.Term > currentLeader.term {
		if currentLeader.leader == "" {
			// we don't have a currentLeader yet, set it
			currentLeader.leader = signal.Leader
		}
		currentLeader.term = signal.Term
	}
	// else signal.Term <= currentLeader.term
	// if the term is not greater, we do not change the current leader, stale request

	// provide the current leader as response
	signal.ResponseCh <- currentLeader.leader
}
