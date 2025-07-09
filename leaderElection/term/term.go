package term

import (
	"log"
)

type IncrementSignal struct {
	ResponseCh chan int
}
type SetTermSignal struct {
	Value int
}
type GetTermSignal struct {
	ResponseCh chan<- int
}

type Term struct {
	currentTerm int
	IncReq      chan IncrementSignal
	SetTermReq  chan SetTermSignal
	GetTermReq  chan GetTermSignal
}

func NewTerm() *Term {
	term := &Term{
		currentTerm: 0,
		IncReq:      make(chan IncrementSignal),
		SetTermReq:  make(chan SetTermSignal),
		GetTermReq:  make(chan GetTermSignal),
	}

	go func() {
		for {
			select {
			case signal := <-term.IncReq:
				term.inc(signal)
			case signal := <-term.SetTermReq:
				term.setTerm(signal)
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

func (t *Term) setTerm(signal SetTermSignal) {
	if signal.Value > t.currentTerm {
		t.currentTerm = signal.Value
	} else {
		log.Printf("Received a term value (%d) that is not greater than the current term (%d)", signal.Value, t.currentTerm)
	}
}
