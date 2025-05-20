package electionTimer

import (
	"log"
	"math/rand"
	"time"
)

type ElectionTimer struct {
	minElectionTimeout int
	maxElectionTimeout int

	timer *time.Timer

	SetMinTimeoutReq chan int
	SetMaxTimeoutReq chan int
	StartReq         chan struct{}
	StopReq          chan struct{}
	ResetReq         chan struct{}
}

func NewElectionTimer(minTimeout, maxTimeout int) *ElectionTimer {
	electionTimer := ElectionTimer{
		minElectionTimeout: minTimeout,
		maxElectionTimeout: maxTimeout,
		SetMinTimeoutReq:   make(chan int),
		SetMaxTimeoutReq:   make(chan int),
		ResetReq:           make(chan struct{}),
		StopReq:            make(chan struct{}),
	}

	go func() {
		for {
			select {
			case minTimeout := <-electionTimer.SetMinTimeoutReq:
				electionTimer.minElectionTimeout = minTimeout
			case maxTimeout := <-electionTimer.SetMaxTimeoutReq:
				electionTimer.maxElectionTimeout = maxTimeout
			case <-electionTimer.StartReq:
				electionTimer.start()
			case <-electionTimer.StopReq:
				electionTimer.stop()
			case <-electionTimer.ResetReq:
				electionTimer.reset()
			}
		}
	}()

	return &electionTimer
}

func (electionTimer *ElectionTimer) start() {
	timeout := rand.Intn(electionTimer.maxElectionTimeout-electionTimer.minElectionTimeout) + electionTimer.minElectionTimeout
	electionTimer.timer = time.NewTimer(time.Duration(timeout) * time.Millisecond)
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

func (electionTimer *ElectionTimer) reset() {
	if electionTimer.timer != nil {
		electionTimer.stop()
		newTimeout := rand.Intn(electionTimer.maxElectionTimeout-electionTimer.minElectionTimeout) + electionTimer.minElectionTimeout
		electionTimer.timer.Reset(time.Duration(newTimeout) * time.Millisecond)
	} else {
		log.Fatal("Timer is not running, cannot reset it")
	}
}
