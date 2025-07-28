package internalUtils

type StopLeadershipSignal struct{}

type BecomeLeaderSignal struct {
	Term int
}
