package heartbeatTimer

import (
	"log"
	"time"
)

type SetHeartbeatTimeoutSignal struct {
	Timeout int
}

type HeartbeatTimeoutSignal struct{}
type StartSignal struct {
	ResponseCh chan HeartbeatTimeoutSignal
}
type ResetSignal struct{}
type StopSignal struct{}

type HeartbeatTimer struct {
	timeout time.Duration

	timer *time.Timer

	SetTimeoutReq chan SetHeartbeatTimeoutSignal

	StartReq chan StartSignal
	ResetReq chan ResetSignal
	StopReq  chan StopSignal
}

func NewHeartbeatTimer(timeout int) *HeartbeatTimer {
	if timeout <= 0 {
		// set default value if valid value is not provided
		timeout = 50 * int(time.Millisecond)
	}

	heartbeatTimer := &HeartbeatTimer{
		timeout:       time.Duration(timeout) * time.Millisecond,
		SetTimeoutReq: make(chan SetHeartbeatTimeoutSignal),
		StartReq:      make(chan StartSignal),
		ResetReq:      make(chan ResetSignal),
		StopReq:       make(chan StopSignal),
	}

	go func() {
		for {
			select {
			case signal := <-heartbeatTimer.SetTimeoutReq:
				heartbeatTimer.timeout = time.Millisecond * time.Duration(signal.Timeout)
			case signal := <-heartbeatTimer.StartReq:
				heartbeatTimer.start(signal.ResponseCh)
			case <-heartbeatTimer.ResetReq:
				heartbeatTimer.reset()
			case <-heartbeatTimer.StopReq:
				heartbeatTimer.stop()
			}
		}
	}()

	return heartbeatTimer
}

func (heartbeatTimer *HeartbeatTimer) start(responseCh chan HeartbeatTimeoutSignal) {
	heartbeatTimer.timer = time.NewTimer(heartbeatTimer.timeout)

	go func() {
		<-heartbeatTimer.timer.C
		// timeout has occurred
		responseCh <- HeartbeatTimeoutSignal{}
	}()
}

func (heartbeatTimer *HeartbeatTimer) reset() {
	if heartbeatTimer.timer != nil {
		if !heartbeatTimer.timer.Stop() {
			// drain the timer if it was already stopped
			select {
			case <-heartbeatTimer.timer.C:
			default:
			}
		}
		heartbeatTimer.timer.Reset(heartbeatTimer.timeout)
	} else {
		log.Fatal("Heartbeat timer is not running, cannot reset it")
	}
}

func (heartbeatTimer *HeartbeatTimer) stop() {
	if heartbeatTimer.timer != nil {
		if !heartbeatTimer.timer.Stop() {
			// drain the timer if it was already stopped
			select {
			case <-heartbeatTimer.timer.C:
			default:
			}
		}
	} else {
		log.Fatal("Heartbeat timer is not running, cannot stop it")
	}
}
