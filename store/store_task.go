package store

type baseTask[S State] struct {
	age        int64
	subscriber Subscriber[S]
	state      S
	oldState   S
	action     Action
}

func NewTask[S State](age int64, subscriber Subscriber[S], state S, oldState S, action Action) Task {
	return &baseTask[S]{
		age:        age,
		subscriber: subscriber,
		state:      state,
		oldState:   oldState,
		action:     action,
	}
}

func (c *baseTask[S]) Do() {
	if c == nil {
		return
	}
	c.subscriber(c.state, c.oldState, c.action)
}
