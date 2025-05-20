package state

func (state *State) handleLeaderRequest() {
	state.value = "leader"
}
