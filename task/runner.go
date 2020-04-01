package task

import (
	"fmt"

	"github.com/segmentio/go-prompt"
)

// DefaultChannelSize is not really a default, but the hard-coded size of the internal queue for
// running tasks
const DefaultChannelSize = 100

type Runner struct {
	Finished chan bool

	tasks chan Task
}

type Task interface {
	ConfirmationMessage() string
	Run() error
}

func NewRunner() *Runner {
	c := make(chan Task, DefaultChannelSize)
	f := make(chan bool)
	return &Runner{
		tasks:    c,
		Finished: f,
	}
}

func (r *Runner) SubmitTask(t Task) error {
	r.tasks <- t

	return nil
}

func (r *Runner) RunTasks() error {
	for t := range r.tasks {
		if prompt.Confirm(t.ConfirmationMessage()) {
			fmt.Println("running")
			t.Run()
		} else {
			fmt.Println("skipping")
		}
	}
	r.Finished <- true
	return nil
}

func (r *Runner) Finish() {
	close(r.tasks)
}
