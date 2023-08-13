package scheduler

import (
	"fmt"
	"testing"
	"time"
)

type testTask float32

func (t testTask) Run() error {
	time.Sleep(time.Duration(t)) // slow  down the execution
	fmt.Printf("%1.1f\t", t)
	return nil
}

func TestAddRemove(t *testing.T) {

	s := New()
	if s.Tasks() != 0 {
		t.Fatalf("Expected 0 tasks, got %d", s.Tasks())
	}

	t1, t2 := testTask(1), testTask(2)

	s.Add(3, t1)
	s.Add(3, t2)
	if s.Tasks() != 2 {
		t.Fatalf("Expected 2 tasks, got %d", s.Tasks())
	}

	s.Remove(t2)
	if s.Tasks() != 1 {
		t.Fatalf("Expected 1 tasks, got %d", s.Tasks())
	}

	s.Remove(t1)
	if s.Tasks() != 0 {
		t.Fatalf("Expected 0 tasks, got %d", s.Tasks())
	}
}

func TestTicksVisualManual(_ *testing.T) {
	t1, t2, t3, t11, t21, t22, t33 := testTask(1.0), testTask(2.0), testTask(3.0), testTask(1.1), testTask(2.1), testTask(2.2), testTask(3.3)
	t5, t51, t52, t53 := testTask(5.0), testTask(5.1), testTask(5.2), testTask(5.3)
	s := New()

	s.Add(3, t3, t33)
	s.Add(2, t2, t21, t22)
	s.Add(1, t1, t11)
	s.Add(5, t5, t51, t52, t53)

	for i := 0; i < 25; i++ {
		fmt.Printf("\nStep %d : ", i)
		s.(*scheduler).tick()
	}
}

func TestTicksVisualAuto(t *testing.T) {
	t1, t2, t3, t11, t21, t22, t33 := testTask(1.0), testTask(2.0), testTask(3.0), testTask(1.1), testTask(2.1), testTask(2.2), testTask(3.3)
	t5, t51, t52, t53 := testTask(5.0), testTask(5.1), testTask(5.2), testTask(5.3)
	trace := Trace(testTask(100.0))

	s := New()

	s.Add(3, t3, t33)
	s.Add(2, t2, t21, t22)
	s.Add(1, t1, t11)
	s.Add(5, t5, t51, t52, t53)

	s.Add(2, trace) // This task is traced

	s.SetBefore(func(_ Scheduler) {
		fmt.Print(time.Now(), "\t")
	})

	s.SetAfter(func(_ Scheduler) {
		fmt.Println(time.Now())
	})

	t.Log("Starting with 1/3 second delay")
	s.Start(time.Second / 3)
	s.Add(2, t1) // add tasks while running
	time.Sleep(time.Second)
	s.Add(2, t1) // add tasks while running
	s.Remove(t2) // remove tasks while running
	time.Sleep(time.Second)
	ps(s, trace)
	time.Sleep(time.Second * 2)
	t.Log("Stop request ...")
	s.Stop()
	t.Log("Stopped !")
	ps(s, trace)

}

func TestOverrun(t *testing.T) {
	s := New()
	t1, t2 := testTask(100.0), testTask(200.0)
	trace := Trace(testTask(10.0))

	s.Add(1, t1, t1, t1, t1, t1, t1, t1, t1, t1, t1, t1, t1, t1, t2)
	s.Add(2, t2)
	s.Add(3, trace)

	s.Start(time.Second / 10)
	ps(s, trace)
	time.Sleep(time.Second)
	ps(s, trace)
	time.Sleep(time.Second)
	ps(s, trace)
	time.Sleep(time.Second)
	ps(s, trace)
	time.Sleep(time.Second)
	ps(s, trace)
	time.Sleep(time.Second)
	ps(s, trace)

	s.Add(1, t2) // add while running and saturated

	time.Sleep(time.Second)
	ps(s, trace)
	time.Sleep(time.Second)
	ps(s, trace)

	s.Stop()
	ps(s, trace)

}

func ps(s Scheduler, trace *TaskTracer) {

	fmt.Printf("\n=================================\nLoad : %0.2f %% Elapsed : %v (%d ticks) \n", 100*s.Load(), s.Elapsed(), s.Ticks())
	fmt.Printf("Trace : count : %d, avg : %v, min :%v, max : %v, stdev : %v\n===========================================\n",
		trace.Count(),
		trace.AverageDuration(),
		trace.MinDuration(),
		trace.MaxDuration(),
		trace.StandardDeviationDuration())
}
