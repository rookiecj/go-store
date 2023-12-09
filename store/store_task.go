package store

type baseTask[S State] struct {
	subscriber Subscriber[S]
	state      S
	oldState   S
	action     Action
}

func NewTask[S State](subscriber Subscriber[S], state S, oldState S, action Action) Task {
	return &baseTask[S]{
		subscriber: subscriber,
		state:      state,
		oldState:   oldState,
		action:     action,
	}
}

func (c *baseTask[S]) Do() {
	c.subscriber(c.state, c.oldState, c.action)
}
