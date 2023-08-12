# scheduler
A simple scheduler lib for repetitive tasks


# How to use

```
import "github.com/xavier268/scheduler"

// ....

s := scheduler.New()                // create empty scheduler
defer s.Close()                     // defer ressource release

// Add tasks
s.Add(10, task1, task2, task3)      // These 3 tasks will run every 20 ticks

s.Start(time.Second / 10)           // Start scheduler, a tick every tenth of second

// tasks are now all running in a single separate goroutine. 
// You do your other stuff here....

s.Stop()                            // Temporary pause the scheduler
s.Remove(task2)                     // remove task when scheduler is stopped
s.Start(time.Minute)                // Restart at a different speed. Can also be used to reset the speed without stopping.

```

See test file for examples.


## Features

Task must follow the *Task* interface. They should execute rapidly compared with the tick duration.
If they return an error, they are removed from the scheduler and will not be called again.

When Tasks are added, a period is specified as a number of ticks, between two successive calls.
Tasks can be added and removed when the scheduler is not running.

A specific single goroutine is attached to each started scheduler, and all tasks execute sequentially within this goroutine.

Scheduling attempts to balance the tasks evenly over time bteween ticks, but not within a specific interval between two ticks.

* At each tick, the same approximative number of task will be run.
* At each tick, the task that should run are called in a fixed order, sequentially in the same thread. If the last one finishes after the next tick was send, the new tasks are run immediately. If not, nothing happens until the new tick arrives.

When the scheduler is closed, pending tasks are never interrupted.
