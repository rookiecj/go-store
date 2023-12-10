package store

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

func (c *doTask) Result() any {
	return nil
}
