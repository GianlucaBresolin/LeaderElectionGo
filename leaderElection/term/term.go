package term

import (
	"log"
)

type IncrementSignal struct {
	ResponseCh chan int
}

type Term struct {
	currentTerm int
	IncReq      chan IncrementSignal
	SetValueReq chan int
	GetTermReq  chan chan int
}

func NewTerm() *Term {
	incReq := make(chan IncrementSignal)

	term := &Term{
		currentTerm: 0,
		IncReq:      incReq,
	}

	go func() {
		for {
			select {
			case signal := <-incReq:
				term.inc(signal)
			case value := <-term.SetValueReq:
				term.setValue(value)
			case responseCh := <-term.GetTermReq:
				responseCh <- term.currentTerm
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

func (t *Term) setValue(value int) {
	if value > t.currentTerm {
		t.currentTerm = value
	} else {
		log.Printf("Received a term value (%d) that is not greater than the current term (%d)", value, t.currentTerm)
	}
}
