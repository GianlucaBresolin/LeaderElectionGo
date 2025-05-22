package electionTimer

import (
	"log"
	"math/rand"
	"time"
)

type ElectionTimeoutSignal struct{}
type ResetSignal struct{}
type StopSignal struct{}

type ElectionTimer struct {
	minElectionTimeout int
	maxElectionTimeout int

	timer *time.Timer

	SetMinTimeoutReq chan int
	SetMaxTimeoutReq chan int
	StartReq         chan chan ElectionTimeoutSignal
	StopReq          chan StopSignal
	ResetReq         chan ResetSignal
}

func NewElectionTimer(minTimeout, maxTimeout int) *ElectionTimer {
	electionTimer := ElectionTimer{
		minElectionTimeout: minTimeout,
		maxElectionTimeout: maxTimeout,
		SetMinTimeoutReq:   make(chan int),
		SetMaxTimeoutReq:   make(chan int),
		StartReq:           make(chan chan ElectionTimeoutSignal),
		ResetReq:           make(chan ResetSignal),
		StopReq:            make(chan StopSignal),
	}

	go func() {
		for {
			select {
			case minTimeout := <-electionTimer.SetMinTimeoutReq:
				electionTimer.minElectionTimeout = minTimeout
			case maxTimeout := <-electionTimer.SetMaxTimeoutReq:
				electionTimer.maxElectionTimeout = maxTimeout
			case signalCh := <-electionTimer.StartReq:
				electionTimer.start(signalCh)
			case <-electionTimer.StopReq:
				electionTimer.stop()
			case <-electionTimer.ResetReq:
				electionTimer.reset()
			}
		}
	}()

	return &electionTimer
}

func (electionTimer *ElectionTimer) start(signalCh chan ElectionTimeoutSignal) {
	timeout := rand.Intn(electionTimer.maxElectionTimeout-electionTimer.minElectionTimeout) + electionTimer.minElectionTimeout
	electionTimer.timer = time.NewTimer(time.Duration(timeout) * time.Millisecond)

	go func() {
		<-electionTimer.timer.C
		// timout has occurred
		signalCh <- ElectionTimeoutSignal{}
	}()
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
