# scheduler
A scheduler for periodic tasks. Thread safe and efficient.

# How to use

```
import "github.com/xavier268/scheduler"

// ....

s := scheduler.New()                // create empty scheduler
defer s.Stop()                     // defer ressource release

// Add tasks
s.Add(10, task1, task2, task3)      // These 3 tasks will run every 10 ticks

s.Start(time.Second / 10)           // Start scheduler, a tick every tenth of second

// tasks are now all running in a single separate goroutine. 
// You do your other stuff here....

s.Remove(task2)                     // add or remove task while scheduler is running
fmt.Println("Scheduler load : %.2f %%",s.Load()*100)    // display scheduler load as a percentage

s2 := s.New()                       // prepare a second scheduler with the same tasks
s2.Start(time.Minute)               // run it at a différent seed in parallele
defer s2.Stop()

```

See test file for examples.


## Features

Task must follow the *Task* interface. They should execute rapidly compared with the tick duration. If a task freeze it will lock the scheduler.
If they return an error, they are removed from the scheduler and will not be called again.

When Tasks are added, a period is specified as a number of ticks, between two successive calls.
Tasks can be added and removed when the scheduler is running.

A specific single goroutine is attached to each started scheduler, and all tasks execute sequentially within this goroutine.

Scheduling attempts to balance the tasks evenly over time between ticks, but not within a specific interval between two ticks.

* At each tick, the same approximative number of task will be run.
* At each tick, the task that should run are called in a fixed order, sequentially in the same thread. If the last one finishes after the next tick was send, the new tasks are run immediately. If not, nothing happens until the new tick arrives.

When the scheduler is stopped, it cannot be restarted. Create a New one reusing the existing tasked from the stopped one.

If the scheduler load exceeds 100% (overrun) some tasks will not run, but the scheduler should not freeze.
