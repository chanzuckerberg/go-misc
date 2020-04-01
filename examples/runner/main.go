package main

import (
	"fmt"

	"github.com/chanzuckerberg/go-misc/task"
)

type ExampleTask struct {
	name string
}

func (r ExampleTask) Run() error {
	fmt.Printf("running ExampleRunner %s\n", r.name)
	return nil
}

func (r ExampleTask) ConfirmationMessage() string {
	return fmt.Sprintf("do you want to run '%s'?", r.name)
}

func main() {
	t := ExampleTask{name: "foo"}

	r := task.NewRunner()

	go func() {
		r.SubmitTask(t)
		r.Finish()
	}()

	go func() {
		r.RunTasks()
	}()

	<-r.Finished
}
