package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID         uuid.UUID
	Title      string
	Desc       string
	StartTime  time.Time
	FinishTime time.Time
	Complete   bool
	Next       *Task
	Prev       *Task
}

func (t Task) ToString() string {
	str := fmt.Sprintf(
		"Title: %s\t"+
			"Complete: %t\t"+
			"Time Starte/Finished: %s - %s\t"+
			"Description: %s\t"+
			"ID: %s\t",
		t.Title,
		t.Complete,
		t.StartTime.UTC().Format("Jan 2, 2006 at 3:04pm (CST)"),
		t.FinishTime.UTC().Format("Jan 2, 2006 at 3:04pm (CST)"),
		t.Desc,
		t.ID)
	return str
}

type TaskList struct {
	Head *Task
	Tail *Task
	Size int
}

// Clocks allow mokey patching in tests
type Clock interface {
	Now() time.Time
}

type TaskClock struct{}

func (t *TaskClock) Now() time.Time {
	return time.Now()
}

var clock Clock = &TaskClock{}

// Allows for Monkey Patching in tests
type Ident interface {
	New() uuid.UUID
}

type TaskIdent struct{}

func (ident *TaskIdent) New() uuid.UUID {
	return uuid.New()
}

var ident Ident = &TaskIdent{}

func MakeTask(title, desc string, complete bool) *Task {
	id := ident.New()
	task := Task{}
	if complete {
		task = Task{id, title, desc, clock.Now(), clock.Now(), complete, nil, nil}
	} else {
		task = Task{id, title, desc, clock.Now(), time.Time{}, complete, nil, nil}
	}
	return &task
}

type Options struct {
	task     *Task
	ID       uuid.UUID
	Title    string
	Desc     string
	Complete bool
}

func (tl *TaskList) UpdateTask(op *Options) {
	if op.task != nil {
		tl.updateTaskWithTask(op.task)
	} else {
		tl.updateTaskWithVariables(op.ID, op.Title, op.Desc, op.Complete)
	}
}

func (tl *TaskList) updateTaskWithVariables(id uuid.UUID, title, desc string, complete bool) error {
	oldTask, err := tl.GetTaskByID(id)
	if err != nil {
		return err
	}
	oldTask.Title = title
	oldTask.Desc = desc
	if complete && !oldTask.Complete {
		oldTask.Complete = true
		oldTask.FinishTime = time.Now()
	} else if oldTask.Complete && !complete {
		oldTask.Complete = false
		oldTask.FinishTime = time.Time{}
	}
	return nil
}

func (tl *TaskList) updateTaskWithTask(t *Task) error {
	oldtask, err := tl.GetTaskByID(t.ID)
	if err != nil {
		return err
	}

	oldtask.Title = t.Title
	oldtask.Desc = t.Desc
	if t.Complete && !oldtask.Complete {
		oldtask.Complete = true
		oldtask.FinishTime = clock.Now()
	} else if oldtask.Complete && !t.Complete {
		oldtask.Complete = false
		oldtask.FinishTime = time.Time{}
	}
	return nil
}

func (tl *TaskList) GetTaskByID(id uuid.UUID) (*Task, error) {
	current := tl.Head
	//assert head has data
	if current == nil {
		return nil, errors.New("TaskList Head is nil")
	}
	for {
		if current.ID == id {
			return current, nil
		}
		if current.Next == nil {
			break
		}
		current = current.Next
	}
	return nil, errors.New("ID NOT found")
}

func (tl *TaskList) GetTaskByTitle(title string) (*Task, error) {
	current := tl.Head
	//assert head has data
	if current == nil {
		return nil, errors.New("TaskList Head is nil")
	}
	for {
		if current.Title == title {
			return current, nil
		}
		if current.Next == nil {
			break
		}
		current = current.Next
	}
	return nil, errors.New("title not found")
}

func (tl *TaskList) GetAllTasks() []*Task {
	current := tl.Head

	// make sure tl has content
	if tl.Size == 0 {
		return nil
	}
	// make array the exact size needed
	allTasks := make([]*Task, tl.Size)

	for i := range allTasks {
		allTasks[i] = current
		if current.Next != nil {
			current = current.Next
		} else {
			break
		}
	}
	return allTasks
}

// Task is passed by value as to prevent chaos in linked list
func (tl *TaskList) AppendTask(t Task) {
	if tl.Size == 0 {
		tl.Head = &t
		tl.Tail = &t
	} else {
		// link t to tail of tasklist
		t.Prev = tl.Tail
		// link the tail of tasklist to t
		tl.Tail.Next = &t
		// Make t the new tail
		tl.Tail = &t
	}
	tl.Size++
}

func (tl *TaskList) RemoveTaskByTitle(title string) error {
	// check if tasklist has content
	if tl.Size == 0 {
		return errors.New("TaskList is empty: Nothing to remove")
	}
	// get task
	task, err := tl.GetTaskByTitle(title)
	// return error if task not found
	if err != nil {
		return err
	}

	// if
	if tl.Size == 1 {
		// unlink Tasklist from requested task
		tl.Head = nil
		tl.Tail = nil
		// unlink requested task from everything
		task.Prev = nil
		task.Next = nil
		tl.Size = tl.Size - 1
		return nil
	}
	if tl.Head.Title == title {
		// unlink list from requested task
		tl.Head = task.Next
		// unlink nextTask from requested task
		task.Next.Prev = nil
		// unlink requested task from everything
		task.Prev = nil
		task.Next = nil
		tl.Size = tl.Size - 1
		return nil
	}
	if tl.Tail.Title == title {
		// link tail to previos task
		tl.Tail = task.Prev
		// unlink previous task from requested task
		task.Prev.Next = nil

		// unlink requested task from everything
		task.Prev = nil
		task.Next = nil
		tl.Size = tl.Size - 1
		return nil
	}
	// link up prev and next task to each other
	task.Prev.Next = task.Next
	task.Next.Prev = task.Prev
	// unlink requested task from everything
	task.Prev = nil
	task.Next = nil
	tl.Size = tl.Size - 1

	return nil
}

func (tl *TaskList) RemoveTaskByID(ID uuid.UUID) error {
	// check if tasklist has content
	if tl.Size == 0 {
		return errors.New("TaskList is empty: Nothing to remove")
	}
	// get task
	task, err := tl.GetTaskByID(ID)
	// return error if task not found
	if err != nil {
		return err
	}

	// check if requested task is only task
	if tl.Size == 1 {
		// unlink Tasklist from requested task
		tl.Head = nil
		tl.Tail = nil
		// unlink requested task from everything
		task.Prev = nil
		task.Next = nil
		tl.Size = tl.Size - 1
		return nil
	}
	if tl.Head.ID == ID {
		// unlink list from requested task
		tl.Head = task.Next
		// unlink nextTask from requested task
		task.Next.Prev = nil
		// unlink requested task from everything
		task.Prev = nil
		task.Next = nil
		tl.Size = tl.Size - 1
		return nil
	}
	if tl.Tail.ID == ID {
		// link tail to previos task
		tl.Tail = task.Prev
		// unlink previous task from requested task
		task.Prev.Next = nil

		// unlink requested task from everything
		task.Prev = nil
		task.Next = nil
		tl.Size = tl.Size - 1
		return nil
	}
	// link up prev and next task to each other
	task.Prev.Next = task.Next
	task.Next.Prev = task.Prev
	// unlink requested task from everything
	task.Prev = nil
	task.Next = nil
	tl.Size = tl.Size - 1

	return nil
}
