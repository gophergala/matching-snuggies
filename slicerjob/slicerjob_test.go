package slicerjob

import "testing"

func TestSlicerJob(t *testing.T) {
	job := New("")
	if job.ID == "" {
		t.Fatalf("new job missing ID")
	}

	job = New("http://example.com/slicer/jobs/%d")
	if job.ID == "" {
		t.Fatalf("new job missing ID")
	}
	if job.URL == "" {
		t.Fatalf("new job supplied with urlformat is missing URL")
	}
}
