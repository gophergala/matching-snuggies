package main

import (
	"fmt"
	"strings"
	"sync"
)

// Scheduler is the write-end of a job queue.  It takes a mesh file url, the
// name of a slicer and a preset for that slicer.  Scheduler is responsible for
// routing the job to a machine capable of servicing the request.
type Scheduler interface {
	ScheduleSliceJob(id, meshurl, slicer, preset string) (cancel func(), err error)
}

// Consumer is the read-end of a job queue.  It reserves a from the queue and
// ensures any remote mesh file locations are downloaded to local paths.
type Consumer interface {
	NextSliceJob() (*Job, error)
}

type Job struct {
	ID       string
	MeshPath string
	Slicer   string
	Preset   string

	// Cancel receives a value if the job has been cancelled by the scheduling
	// process.
	Cancel <-chan error

	// Done is called when the slicing process has terminated.  Done is passed
	// a URL at which the output G-code can be retreived.  If the G-code could
	// not be generated due to failure a non-nil error must be passed to Done.
	Done func(location string, err error)
}

// MemQueue is an in memory database and job queue that implements the
// Scheduler and Consumer interfaces.  MemQueue is safe for many producers and
// consumers to be calling interface methods simultaneously.
type MemQueue struct {
	Done func(id, location string, err error)
	cond sync.Cond
	jobs []*memJob
	db   map[string]*memJob
}

var _ Scheduler = new(MemQueue)
var _ Consumer = new(MemQueue)

// MemeoryQueue allocates and initializes a new MemQueue.
func MemoryQueue(done func(id, location string, err error)) *MemQueue {
	return &MemQueue{
		Done: done,
		cond: sync.Cond{L: new(sync.Mutex)},
	}
}

// ScheduleSliceJob enqueues a job in q.
func (q *MemQueue) ScheduleSliceJob(id, meshurl, slicer, preset string) (cancel func(), err error) {
	j := &memJob{
		ID:       id,
		Location: meshurl,
		Slicer:   slicer,
		Preset:   preset,
		Cancel:   make(chan error, 1),
		Done:     make(chan struct{}),
		Fin:      q.Done,
	}

	// append the job to the queue and signal a waiting consumer goroutine to
	// wake up and process the job.
	q.cond.L.Lock()
	q.jobs = append(q.jobs, j)
	q.cond.Signal()
	q.cond.L.Unlock()

	cancel = func() { j.Cancel <- fmt.Errorf("the job was cancelled") }
	return cancel, nil
}

// NextSliceJob dequeues a job from q or blocks until one is available.
func (q *MemQueue) NextSliceJob() (*Job, error) {
	q.cond.L.Lock()
	for len(q.jobs) == 0 {
		q.cond.Wait()
	}
	j := q.jobs[0]
	q.jobs = q.jobs[1:]
	q.cond.L.Unlock()
	return j.Job(), nil
}

type memJob struct {
	ID       string
	Location string
	Slicer   string
	Preset   string
	Cancel   chan error
	Done     chan struct{}
	Fin      func(string, string, error)
}

func (m *memJob) Job() *Job {
	return &Job{
		ID:       m.ID,
		MeshPath: strings.TrimPrefix(m.Location, "file://"),
		Slicer:   m.Slicer,
		Preset:   m.Preset,
		Cancel:   m.Cancel,
		Done: func(location string, err error) {
			close(m.Done)
			m.Fin(m.ID, location, err)
		},
	}
}
