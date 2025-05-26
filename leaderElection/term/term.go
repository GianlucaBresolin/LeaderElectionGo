package term

import (
	"log"
)

type IncrementSignal struct {
	ResponseCh chan int
}
type SetValueSignal struct {
	Value int
}
type GetTermSignal struct {
	ResponseCh chan int
}

type Term struct {
	currentTerm int
	IncReq      chan IncrementSignal
	SetValueReq chan SetValueSignal
	GetTermReq  chan GetTermSignal
}

func NewTerm() *Term {
	term := &Term{
		currentTerm: 0,
		IncReq:      make(chan IncrementSignal),
		SetValueReq: make(chan SetValueSignal),
		GetTermReq:  make(chan GetTermSignal),
	}

	go func() {
		for {
			select {
			case signal := <-term.IncReq:
				term.inc(signal)
			case signal := <-term.SetValueReq:
				term.setValue(signal)
			case signal := <-term.GetTermReq:
				signal.ResponseCh <- term.currentTerm
			}
		}
	}()

	return term
}

func (t *Term) inc(signal IncrementSignal) {
	t.currentTerm++

	// if the response channel is present, send the current term back
	if signal.ResponseCh != nil {
		signal.ResponseCh <- t.currentTerm
	}
}

func (t *Term) setValue(signal SetValueSignal) {
	if signal.Value > t.currentTerm {
		t.currentTerm = signal.Value
	} else {
		log.Printf("Received a term value (%d) that is not greater than the current term (%d)", signal.Value, t.currentTerm)
	}
}
