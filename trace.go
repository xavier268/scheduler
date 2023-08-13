package scheduler

import (
	"math"
	"sync"
	"time"
)

// TaskTracer is a wrapper around a Task that allows the Task stats to be traced.
// TaskTracer is itself a Task.
type TaskTracer struct {
	task  Task         // underlying Task
	count int64        // nb of calls to Run
	d     int64        // cumulative duration in nanoseconds
	d2    int64        // cumulative  duration squared
	max   int64        // max duration
	min   int64        // min duration
	lock  sync.RWMutex // lock for the stats
}

var _ Task = &TaskTracer{} // TaskTracer implements Task

// Return a TaskTracer to be registered in the scheduler as a normal Task.
func Trace(t Task) *TaskTracer {
	return &TaskTracer{
		task:  t,
		count: 0,
		d:     0,
		d2:    0,
		max:   0,
		min:   math.MaxInt64,
		lock:  sync.RWMutex{},
	}
}

func (t *TaskTracer) Run() error {

	t.lock.Lock()
	defer t.lock.Unlock()

	start := time.Now()
	err := t.task.Run()
	dur := int64(time.Now().Sub(start))

	t.count += 1
	t.d += dur
	t.d2 += dur * dur
	t.max = max(t.max, dur)
	t.min = min(t.min, dur)

	return err
}

// Count is the nb of calls to Run
func (t *TaskTracer) Count() int64 {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.count
}

// CumulativeDuration is the cumulative duration of the task
func (t *TaskTracer) CumulativeDuration() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return time.Duration(t.d)
}

// AverageDuration is the average duration of the task
func (t *TaskTracer) AverageDuration() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.count == 0 {
		return 0
	}
	return time.Duration(float64(t.d) / float64(t.count))
}

// MaxDuration is the maximum duration of the task
func (t *TaskTracer) MaxDuration() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return time.Duration(t.max)
}

// MinDuration is the minimum duration of the task
func (t *TaskTracer) MinDuration() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return time.Duration(t.min)
}

// StandardDeviationDuration is the standard deviation of the duration of the task
func (t *TaskTracer) StandardDeviationDuration() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.count <= 1 {
		return 0
	}
	return time.Duration(math.Sqrt(float64(t.d2)/float64(t.count) - float64(t.d)/float64(t.count)*float64(t.d)/float64(t.count)))
}

// Reset stats
func (t *TaskTracer) Reset() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.count = 0
	t.d = 0
	t.d2 = 0
	t.max = 0
	t.min = math.MaxInt64
}
