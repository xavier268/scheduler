package scheduler

import (
	"fmt"
	"testing"
)

type testTask float32

func (t testTask) Run() error {
	fmt.Printf("%1.1f\t", t)
	return nil
}

func TestAddRemove(t *testing.T) {

	s := New()
	if s.NbTasks() != 0 {
		t.Fatalf("Expected 0 tasks, got %d", s.NbTasks())
	}

	t1, t2 := testTask(1), testTask(2)

	s.Add(3, t1)
	s.Add(3, t2)
	if s.NbTasks() != 2 {
		t.Fatalf("Expected 2 tasks, got %d", s.NbTasks())
	}

	s.Remove(t2)
	if s.NbTasks() != 1 {
		t.Fatalf("Expected 1 tasks, got %d", s.NbTasks())
	}

	s.Remove(t1)
	if s.NbTasks() != 0 {
		t.Fatalf("Expected 0 tasks, got %d", s.NbTasks())
	}
}

func TestTicksVisual(_ *testing.T) {
	t1, t2, t3, t11, t21, t22, t33 := testTask(1.0), testTask(2.0), testTask(3.0), testTask(1.1), testTask(2.1), testTask(2.2), testTask(3.3)
	t5, t51, t52, t53 := testTask(5.0), testTask(5.1), testTask(5.2), testTask(5.3)
	s := New()

	s.Add(3, t3, t33)
	s.Add(2, t2, t21, t22)
	s.Add(1, t1, t11)
	s.Add(5, t5, t51, t52, t53)

	for i := 0; i < 25; i++ {
		fmt.Printf("\nStep %d : ", i)
		s.Tick()
	}
}