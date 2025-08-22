package term

type SetTermSignal struct {
	Value      int
	ResponseCh chan<- bool
}
type GetTermSignal struct {
	ResponseCh chan<- int
}

type Term struct {
	currentTerm int
	SetTermReq  chan SetTermSignal
	GetTermReq  chan GetTermSignal
}

func NewTerm() *Term {
	term := &Term{
		currentTerm: 0,
		SetTermReq:  make(chan SetTermSignal),
		GetTermReq:  make(chan GetTermSignal),
	}

	go func() {
		for {
			select {
			case signal := <-term.SetTermReq:
				term.setTerm(signal)
			case signal := <-term.GetTermReq:
				signal.ResponseCh <- term.currentTerm
			}
		}
	}()

	return term
}

func (t *Term) setTerm(signal SetTermSignal) {
	if signal.Value >= t.currentTerm {
		t.currentTerm = signal.Value
		if signal.ResponseCh != nil {
			signal.ResponseCh <- true
		}
	} else {
		// stale term
		if signal.ResponseCh != nil {
			signal.ResponseCh <- false
		}
	}
}
