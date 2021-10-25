package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// Monkey Patching to create consistant test values
type MockClock struct{}

func (t *MockClock) Now() time.Time {
	return time.Time{}
}

type MockIdent struct{}

func (id *MockIdent) New() uuid.UUID {
	return uuid.UUID{}
}
func TestAppendTask_HappyPath(t *testing.T) {
	// Set Up
	actualClock := clock
	defer func() { clock = actualClock }()
	clock = &MockClock{}
	actualIdent := ident
	defer func() { ident = actualIdent }()
	ident = &MockIdent{}
	var suv TaskList
	expected := make([]*Task, 100)
	for i := 0; i < 100; i++ {
		expected[i] = MakeTask("test title"+fmt.Sprint(i), "test description"+fmt.Sprint(i), false)
	}

	// Execute
	for _, task := range expected {
		suv.AppendTask(*task)
	}

	// Evaluate
	current := suv.Head
	for _, task := range expected {
		if !cmp.Equal(task, current) {
			t.Fatalf("actual %v does not equal expected %v", task.ToString(), current.ToString())
		}
		current = current.Next
	}
}
func TestGetAllTasks_HappyPath(t *testing.T) {
	// Set Up
	expected := []*Task{
		MakeTask("test title1", "test description1", false),
		MakeTask("test title2", "test description2", false),
		MakeTask("test title3", "test description3", false),
	}
	// system under test
	var sut TaskList
	for _, v := range expected {
		sut.AppendTask(*v)
	}

	// Execute
	actual := sut.GetAllTasks()

	// Evaluate
	for i, _ := range actual {
		if expected[i].ToString() != actual[i].ToString() {
			t.Fatalf("actual %v does not equal expected %v", actual[i].ToString(), expected[i].ToString())
		}
	}

}

func TestUpdateTask(t *testing.T) {
	// Set Up
	MakeTask("test title1", "test description1", false)
	MakeTask("test title2", "test description2", false)
	MakeTask("test title3", "test description3", false)

}

func TestMakeTask_HappyPath(t *testing.T) {
	// Setup
	expected := &Task{
		Title:    "test title1",
		Desc:     "test description1",
		Complete: false,
	}
	actualClock := clock
	defer func() { clock = actualClock }()
	clock = &MockClock{}
	actualIdent := ident
	defer func() { ident = actualIdent }()
	ident = &MockIdent{}

	// Execute
	actual := MakeTask("test title1", "test description1", false)

	// Evaluate
	if !cmp.Equal(expected, actual) {
		t.Fatalf("actual %v does not equal expected %v", actual.ToString(), expected.ToString())
	}
}

// | close connection		|
// | bullshit				|
// | clock = actualClock	|
// _________________________

// defer.pop().execute()
