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

type ElectionTimeoutSignal struct{}
type StartSignal struct {
	ResponseCh chan ElectionTimeoutSignal
}
type ResetSignal struct{}
type StopSignal struct{}

type ElectionTimer struct {
	minElectionTimeout int
	maxElectionTimeout int

	timer *time.Timer

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
				electionTimer.start(signal.ResponseCh)
			case <-electionTimer.ResetReq:
				electionTimer.reset()
			case <-electionTimer.StopReq:
				electionTimer.stop()
			}
		}
	}()

	return electionTimer
}

func (electionTimer *ElectionTimer) start(responseCh chan ElectionTimeoutSignal) {
	timeout := rand.Intn(electionTimer.maxElectionTimeout-electionTimer.minElectionTimeout) + electionTimer.minElectionTimeout
	electionTimer.timer = time.NewTimer(time.Duration(timeout) * time.Millisecond)

	go func() {
		for {
			<-electionTimer.timer.C
			// timer has fired, send the timeoutsignal
			responseCh <- ElectionTimeoutSignal{}
		}
	}()
}

func (electionTimer *ElectionTimer) reset() {
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
}

func (electionTimer *ElectionTimer) stop() {
	if electionTimer.timer != nil {
		if !electionTimer.timer.Stop() {
			// drain the channel if the timer has already fired
			select {
			case <-electionTimer.timer.C:
			default:
			}
		}
	} else {
		log.Fatal("Timer is not running, cannot stop it")
	}
}
