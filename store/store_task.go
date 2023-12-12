package store

type Task interface {
	// Do runs a Task
	Do()
}

type doTask struct {
	task func()
}

func NewTask(task func()) Task {
	return &doTask{
		task: task,
	}
}

func (c *doTask) Do() {
	c.task()
}
