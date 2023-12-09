package store

type dispatchTask[S State] struct {
	age        int64
	subscriber Subscriber[S]
	state      S
	oldState   S
	action     Action
}

type reduceTask[S State] struct {
	age      int64
	reducer  Reducer[S]
	state    S
	action   Action
	newState S
}

func NewDispatchTask[S State](age int64, subscriber Subscriber[S], state S, oldState S, action Action) Task {
	return &dispatchTask[S]{
		age:        age,
		subscriber: subscriber,
		state:      state,
		oldState:   oldState,
		action:     action,
	}
}

func NewReduceTask[S State](age int64, reducer Reducer[S], state S, action Action) Task {
	return &reduceTask[S]{
		age:     age,
		reducer: reducer,
		state:   state,
		action:  action,
	}
}

func (c *dispatchTask[S]) Do() {
	if c == nil {
		return
	}
	c.subscriber(c.state, c.oldState, c.action)
}

func (c *dispatchTask[S]) Result() any {
	return nil
}

func (c *reduceTask[S]) Do() {
	if c == nil {
		return
	}
	c.newState = c.reducer(c.state, c.action)
}

func (c *reduceTask[S]) Result() any {
	return c.newState
}
