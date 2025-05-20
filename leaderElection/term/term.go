package term

import (
	"log"
)

type IncrementSignal struct{}

type Term struct {
	currentTerm int
	IncReq      chan IncrementSignal
	SetValueReq chan int
}

func NewTerm() *Term {
	incReq := make(chan IncrementSignal)

	term := Term{
		currentTerm: 0,
		IncReq:      incReq,
	}

	go func() {
		for {
			select {
			case <-incReq:
				term.inc()
			case value := <-term.SetValueReq:
				term.setValue(value)
			}
		}
	}()

	return &term
}

func (t *Term) inc() {
	t.currentTerm++
}

func (t *Term) setValue(value int) {
	if value > t.currentTerm {
		t.currentTerm = value
	} else {
		log.Printf("Received a term value (%d) that is not greater than the current term (%d)", value, t.currentTerm)
	}
}
