package scheduler

import (
	"log"
	"sync"
	"time"
)

// Tasks are run at regular number of ticks.
// If Task generates an error, it is removed from scheduler.
type Task interface {
	Run() error
}

// Scheduler is responsible for holding tasks and running them at regular intervals.
type Scheduler struct {
	done       chan struct{}  // channel for signalling scheduler closing
	wg         sync.WaitGroup // wait group for scheduler closing
	ticker     *time.Ticker   // ticker for scheduling
	elapsed    int            // total number of ticks since start
	tasks      map[int][]Task // database of active tasks
	beforeTick Hook           // Hook called before all tasks are run at every tick
	afterTick  Hook           // Hook called after all tasks are run at every tick
}

// Hook is executed before and after all tasks are run at every tick
type Hook func(s *Scheduler)

// New creates a new empty scheduler
func New() *Scheduler {
	return &Scheduler{
		ticker:  nil,
		elapsed: 0,
		tasks:   map[int][]Task{},
	}
}

// Set a Hook that will be executed before all tasks are run at every tick.
func (s *Scheduler) SetBefore(h Hook) {
	s.beforeTick = h
}

// Set a Hook that will be executed after all tasks are run at every tick.
func (s *Scheduler) SetAfter(h Hook) {
	s.afterTick = h
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

// Force the next tick from scheduler, calling the active tasks scheduled to run at that time.
// Task that return an error are removed from scheduler.
func (s *Scheduler) tick() {
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

// Start the scheduler asynchoneously, generating ticks every duration.
// If scheduler is already started, resets the tick duration.
func (s *Scheduler) Start(duration time.Duration) {
	if s.ticker != nil {
		s.ticker.Reset(duration) // ticker was running, reset it
	} else {
		s.done = make(chan struct{})        // create a channel to signal scheduler closing
		s.ticker = time.NewTicker(duration) // start ticker
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for range s.ticker.C {
				select {
				case <-s.done:
					// log.Println("DEBUG : scheduler closed")
					return // scheduler close
				default: // tick
					if s.beforeTick != nil {
						s.beforeTick(s)
					}
					s.tick()
					if s.afterTick != nil {
						s.afterTick(s)
					}
				}
			}
			log.Println("Unexpected : no more ticks to process")

		}()
	}
}

// Stop the scheduler, preventing further ticks.
// Tasks are untouched. Can be restarted with Start.
func (s *Scheduler) Stop() {
	s.ticker.Stop()
}

// Close the scheduler, releasing the goroutine and preventing further ticks.
func (s *Scheduler) Close() {
	s.done <- struct{}{}
	s.Stop()
	s.wg.Wait() // wait for scheduler closing
	s.ticker = nil
	return
}

func (s *Scheduler) IsRunning() bool {
	return s.ticker.C != nil
}

func (s *Scheduler) NbTasks() int {
	nb := 0
	for _, v := range s.tasks {
		nb += len(v)
	}
	return nb
}
