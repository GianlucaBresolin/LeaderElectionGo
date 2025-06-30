package currentLeader

type SetCurrentLeaderSignal struct {
	Leader     string
	Term       int
	ResponseCh chan<- string // channel to send back the current leader
}
type ResetSignal struct {
	Term int
}

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
			case signal := <-currentLeader.ResetReq:
				currentLeader.reset(signal)
			}
		}
	}()

	return currentLeader
}

func (currentLeader *CurrentLeader) setCurrentLeader(signal SetCurrentLeaderSignal) {
	switch {
	case signal.Term > currentLeader.term:
		// if the term is greater, we update the leader and term
		currentLeader.term = signal.Term
		currentLeader.leader = signal.Leader
	case signal.Term == currentLeader.term:
		// if the term is equal, we update the leader only if it is empty
		if currentLeader.leader == "" {
			currentLeader.leader = signal.Leader
		}
	case signal.Term < currentLeader.term:
		// stale request, ignore it
	}

	// provide the current leader as response
	signal.ResponseCh <- currentLeader.leader
}

func (currentLeader *CurrentLeader) reset(signal ResetSignal) {
	// reset only if it provide a higher term
	if signal.Term > currentLeader.term {
		currentLeader.term = signal.Term
		currentLeader.leader = ""
	}
	// otherwise stale reset request, ignore it
}
