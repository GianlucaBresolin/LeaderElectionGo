package state

func (state *State) handleFollowerRequest() {
	state.value = "follower"
}
