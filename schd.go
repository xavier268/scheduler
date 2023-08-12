package scheduler

// Tasks are run at regular number of ticks.
// If Task generates an error, it is removed from scheduler.
type Task interface {
	Run() error
}

type Scheduler struct {
	elapsed int            // total number of ticks since start
	tasks   map[int][]Task // database of active tasks
}

// create a new empty scheduler
func New() *Scheduler {
	return &Scheduler{
		elapsed: 0,
		tasks:   make(map[int][]Task),
	}
}

// number of active tasks scheduled
func (s *Scheduler) NbTasks() int {
	r := 0
	for _, v := range s.tasks {
		r += len(v)
	}
	return r
}

// Add tasks sheduled to run every 'period' ticks.
// Negative or 0 period tasks are not scheduled.
// If same atsk is added multiple times, it will be called treated as separate tasks.
func (s *Scheduler) Add(period int, t ...Task) {
	if period <= 0 {
		return
	}
	s.tasks[period] = append(s.tasks[period], t...)
}

// Remove a given task from the scheduler, preserving order of other tasks.
func (s *Scheduler) Remove(t Task) {
	for p, v := range s.tasks {
		for i, tt := range v {
			if tt == t {
				s.tasks[p] = append(v[:i], v[i+1:]...) // order is preserved
				break
			}
		}
	}
}

// Process the next tick from scheduler, calling the active tasks scheduled to run at that time.
// Task that return an error are removed from scheduler.
func (s *Scheduler) Tick() {
	s.elapsed += 1
	for p, v := range s.tasks {
		k := s.elapsed % p
		for i := k; i < len(v); i += p {
			err := v[i].Run()
			if err != nil { // If tasks returns an error, it is removed from scheduler
				s.Remove(v[i])
			}
		}
	}
}
