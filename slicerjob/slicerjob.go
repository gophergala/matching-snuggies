package slicerjob

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"
)

type Job struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	URL      string  `json:"url"`
	GCodeURL string  `json:"gcode_url"`
}

// New creates a new Job with a random UUID for an ID.  If urlformat is
// non-empty the URL of the returned job is computed as
// fmt.Sprintf(urlformat,job.ID).
func New(urlformat string) *Job {
	job := new(Job)
	job.ID = uuid.New()
	if urlformat != "" {
		job.URL = fmt.Sprintf(urlformat, job.ID)
	}
	return job
}
