package leaderElection

type term struct {
	currentTerm int
	incReq      chan struct{}
}

func newTerm() *term {
	incReq := make(chan struct{})

	term := term{
		currentTerm: 0,
		incReq:      make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-incReq:
				term.inc()
			}
		}
	}()

	return &term
}

func (t *term) inc() {
	t.currentTerm++
}
