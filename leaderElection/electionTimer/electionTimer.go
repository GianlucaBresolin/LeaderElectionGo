package electionTimer

import (
	"log"
	"math/rand"
	"time"
)

type SetMinTimeoutSignal struct {
	MinTimeout int
}
type SetMaxTimeoutSignal struct {
	MaxTimeout int
}

type ElectionTimeoutSignal struct {
	Term int
}
type StartSignal struct {
	ResponseCh chan ElectionTimeoutSignal
}
type ResetSignal struct {
	Term int
}
type StopSignal struct {
	Term       int
	ResponseCh chan<- bool
}

type ElectionTimer struct {
	minElectionTimeout int
	maxElectionTimeout int

	timer *time.Timer
	term  int

	SetMinTimeoutReq chan SetMinTimeoutSignal
	SetMaxTimeoutReq chan SetMaxTimeoutSignal

	StartReq chan StartSignal
	ResetReq chan ResetSignal
	StopReq  chan StopSignal
}

func NewElectionTimer(minTimeout, maxTimeout int) *ElectionTimer {
	if minTimeout <= 0 || maxTimeout <= 0 {
		// set default values if valid values are not provided
		minTimeout = 150
		maxTimeout = 300
	}

	electionTimer := &ElectionTimer{
		minElectionTimeout: minTimeout,
		maxElectionTimeout: maxTimeout,
		term:               0,
		SetMinTimeoutReq:   make(chan SetMinTimeoutSignal),
		SetMaxTimeoutReq:   make(chan SetMaxTimeoutSignal),
		StartReq:           make(chan StartSignal),
		ResetReq:           make(chan ResetSignal),
		StopReq:            make(chan StopSignal),
	}

	go func() {
		for {
			select {
			case signal := <-electionTimer.SetMinTimeoutReq:
				electionTimer.minElectionTimeout = signal.MinTimeout
			case signal := <-electionTimer.SetMaxTimeoutReq:
				electionTimer.maxElectionTimeout = signal.MaxTimeout
			case signal := <-electionTimer.StartReq:
				electionTimer.start(signal)
			case signal := <-electionTimer.ResetReq:
				electionTimer.reset(signal)
			case signal := <-electionTimer.StopReq:
				electionTimer.stop(signal)
			}
		}
	}()

	return electionTimer
}

func (electionTimer *ElectionTimer) start(signal StartSignal) {
	timeout := rand.Intn(electionTimer.maxElectionTimeout-electionTimer.minElectionTimeout) + electionTimer.minElectionTimeout
	electionTimer.timer = time.NewTimer(time.Duration(timeout) * time.Millisecond)

	go func() {
		for {
			<-electionTimer.timer.C
			// timer has fired, send the timeoutsignal
			signal.ResponseCh <- ElectionTimeoutSignal{
				Term: electionTimer.term + 1,
			}
		}
	}()
}

func (electionTimer *ElectionTimer) reset(signal ResetSignal) {
	if signal.Term >= electionTimer.term {
		if electionTimer.timer != nil {
			if !electionTimer.timer.Stop() {
				// drain the channel if the timer has already fired
				select {
				case <-electionTimer.timer.C:
				default:
				}
			}
			// reset the timer with a new random timeout
			newTimeout := rand.Intn(electionTimer.maxElectionTimeout-electionTimer.minElectionTimeout) + electionTimer.minElectionTimeout
			electionTimer.timer.Reset(time.Duration(newTimeout) * time.Millisecond)
		} else {
			log.Fatal("Timer is not running, cannot reset it")
		}
		// update the term
		electionTimer.term = signal.Term
	}
	// else, stale request, ignore it
}

func (electionTimer *ElectionTimer) stop(signal StopSignal) {
	if signal.Term >= electionTimer.term {
		if electionTimer.timer != nil {
			if !electionTimer.timer.Stop() {
				// drain the channel if the timer has already fired
				select {
				case <-electionTimer.timer.C:
				default:
				}
			}
			signal.ResponseCh <- true
		} else {
			signal.ResponseCh <- false
		}
		// update the term
		electionTimer.term = signal.Term
	} else {
		// stale request, ignore it
		signal.ResponseCh <- false
	}
}
