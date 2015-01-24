package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gophergala/matching-snuggies/slicerjob"
)

const (
	slicerBackend = "slic3r"
)

type statusType int

const (
	accepted statusType = iota
	processing
	complete
	failed
)

var statuses = []string{
	accepted:   "accepted",
	processing: "processing",
	complete:   "complete",
	failed:     "failed",
}

var config = map[string]string{
	"version": "0.0.0",
	"URL":     "http://localhost:8888/slicer/jobs",
}

func processJob(meshfile multipart.File, slicerBackend string, preset string) *slicerjob.Job {
	job := slicerjob.New()

	//do stuff to the job.
	job.Status = "processing"
	job.Progress = 0.0
	job.URL = fmt.Sprint(config["URL"]+"/", job.ID)

	return job
}

func snuggiedHandler(config map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case (r.Method == "GET"):
			fmt.Fprint(w, "a GET")
		case (r.Method == "POST"):
			//TODO make sure meshfile is at least .stl
			meshfile, _, err := r.FormFile("meshfile")
			if err != nil {
				http.Error(w, "bad meshfile, or 'meshfile' field not present", http.StatusBadRequest)
				return
			}

			slicerBackend := r.FormValue("slicer")
			if slicerBackend != slicerBackend {
				http.Error(w, "slicer not supported", http.StatusBadRequest)
				return
			}

			preset := r.FormValue("preset")
			if preset == "" {
				http.Error(w, "invalid quality config.", http.StatusBadRequest)
				return
			}

			job := processJob(meshfile, slicerBackend, preset)

			jsonJob, err := json.Marshal(job)
			if err != nil {
				http.Error(w, "json didn't encode properly...Derp?\n"+err.Error(), http.StatusBadRequest)
				return
			}

			w.Write(jsonJob)

		default:
			http.Error(w, "not supported", http.StatusMethodNotAllowed)
			return
		}
	}
}

func main() {
	http.HandleFunc("/slicer/jobs/", snuggiedHandler(config))

	log.Fatal(http.ListenAndServe(":8888", nil))
}
