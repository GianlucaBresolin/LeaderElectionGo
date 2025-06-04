package currentLeader

type SetCurrentLeaderSignal struct {
	Leader     string
	ResponseCh chan<- string
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
	if currentLeader.leader == "" {
		// we don't have a currentLeader yet, set it
		currentLeader.leader = signal.Leader
	}
	// provide the current leader as response
	signal.ResponseCh <- currentLeader.leader
}
