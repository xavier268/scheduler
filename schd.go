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

type Scheduler interface {
	// Add tasks to the scheduler.
	Add(period int, t ...Task)
	// Remove a task from the scheduler.
	Remove(t Task)
	// Start the scheduler with the specified clock period.
	Start(duration time.Duration)
	// Stop the scheduler. A stopped scheduler cannot be restarted.
	Stop()
	// Get the elapsed ticks since last scheduler (re)start.
	Ticks() int
	// Get the elapsed duration since last start
	Elapsed() time.Duration
	// Get the number of tasks currently scheduled.
	Tasks() int
	// Set a Hook that will be executed before all tasks are run at every tick.
	SetBefore(h Hook)
	// Set a Hook that will be executed after all tasks are run at every tick.
	SetAfter(h Hook)
	// Get the avearge load of the last run
	Load() float64
}

// scheduler is responsible for holding tasks and running them at regular intervals.
type scheduler struct {
	done   chan struct{}  // channel for signalling scheduler closing
	wg     sync.WaitGroup // wait group for scheduler closing
	ticker *time.Ticker   // ticker for scheduling

	// lockstats sync.RWMutex  // lock for scheduler stats
	duration time.Duration // duration of each tick
	ticks    int           // total number of ticks since start
	load     time.Duration // total running duration since last scheduler start

	// locktasks sync.RWMutex   // lock for scheduler tasks
	tasks map[int][]Task // database of active tasks

	beforeTick Hook // Hook called before all tasks are run at every tick
	afterTick  Hook // Hook called after all tasks are run at every tick

}

// Hook is executed before and after all tasks are run at every tick
type Hook func(s Scheduler)

// New creates a new empty scheduler with the specific duration clock
func New() Scheduler {
	return &scheduler{
		done:     make(chan struct{}),
		wg:       sync.WaitGroup{},
		ticker:   nil,
		duration: 0,
		ticks:    0,
		load:     0,
		tasks:    map[int][]Task{},
		beforeTick: func(s Scheduler) {
		},
		afterTick: func(s Scheduler) {
		},
	}
}

// Total number of ticks since scheduler creation
func (s *scheduler) Ticks() int {

	return s.ticks
}

// Set a Hook that will be executed before all tasks are run at every tick.
func (s *scheduler) SetBefore(h Hook) {
	s.beforeTick = h
}

// Set a Hook that will be executed after all tasks are run at every tick.
func (s *scheduler) SetAfter(h Hook) {
	s.afterTick = h
}

// Add tasks sheduled to run every 'period' ticks.
// Negative or 0 period tasks are not scheduled.
// If same atsk is added multiple times, it will be called treated as separate tasks.
func (s *scheduler) Add(period int, t ...Task) {

	if period <= 0 {
		return
	}
	s.tasks[period] = append(s.tasks[period], t...)
}

// Remove a given task from the scheduler, preserving order of other tasks.
func (s *scheduler) Remove(t Task) {
	s.remove(t)
}

// unsafe remove.
func (s *scheduler) remove(t Task) {
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
func (s *scheduler) tick() {
	start := time.Now()

	if s.beforeTick != nil {
		s.beforeTick(s)
	}

	for p, v := range s.tasks {
		k := s.ticks % p
		for i := k; i < len(v); i += p {
			err := v[i].Run()
			if err != nil { // If tasks returns an error, it is removed from scheduler
				s.remove(v[i])
			}
		}
	}

	if s.afterTick != nil {
		s.afterTick(s)
	}

	s.load = s.load + time.Since(start)
	s.ticks += 1
}

// Start the scheduler asynchoneously, generating ticks every duration.
// If scheduler was already started, even if stopped, panic.
func (s *scheduler) Start(duration time.Duration) {

	if s.ticker != nil {
		panic("trying to start a scheduler already started")
	}

	s.duration = duration
	s.ticker = time.NewTicker(duration) // start ticker
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for range s.ticker.C {
			select {
			case <-s.done:
				log.Println("DEBUG : goroutine terminated")
				return // scheduler close - normal goroutine exit
			default: // tick
				s.tick()
			}
		}
		log.Println("Unexpected : no more ticks to process")
	}()
}

// Stop the scheduler. Stopping a non running scheduler will panic.
// A stopped scheduler should not be started again.
func (s *scheduler) Stop() {

	s.done <- struct{}{} // signal close request
	s.wg.Wait()          // wait for scheduler to finish tasks in current tick.
	s.ticker.Stop()      // stop ticker
	return
}

// Number of active tasks.
func (s *scheduler) Tasks() int {

	nb := 0
	for _, v := range s.tasks {
		nb += len(v)
	}
	return nb
}

// Return load as a percentage of the time spent running tasks versus duration between ticks.
// Calculation will be wrong if duration was changed.
func (s *scheduler) Load() float64 {
	return float64(s.load) / float64(s.Elapsed())
}

// Return the elapsed duration since last start.
// Calculation will be wrong if duration was changed.
func (s *scheduler) Elapsed() time.Duration {

	return s.duration * (time.Duration)(s.ticks)
}
