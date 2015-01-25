package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"flag"

	"github.com/gophergala/matching-snuggies/slicerjob"
)

const (
	slicerBackend = "slic3r"
)

var config = map[string]string{
	"nodeID":  "snuggie0",
	"version": "0.0.0",
	"URL":     "http://localhost:8888",
}

type SnuggieServer struct {
	Config map[string]string

	// Prefix should not end in a slash '/'.
	Prefix  string
	DataDir string

	LocalConsumer bool
	S             Scheduler
	C             Consumer
}

func (srv *SnuggieServer) RegisterHandlers(mux *http.ServeMux) http.Handler {
	mux.HandleFunc(srv.route("/jobs"), func(w http.ResponseWriter, r *http.Request) {
		// the request does not have an ID suffix on the url path so we are
		// either creating or listing jobs.
		switch r.Method {
		case "POST":
			srv.CreateJob(w, r)
		// TODO:
		// GET handler (index)
		default:
			http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc(srv.route("/jobs/"), func(w http.ResponseWriter, r *http.Request) {
		// the request has an ID suffix on the url path so we are showing a
		// single job resource.
		switch r.Method {
		case "GET":
			srv.GetJob(w, r)
		// TODO:
		// allow DELETE requests to cancel jobs
		default:
			http.Error(w, "only GET is allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc(srv.route("/gcodes/"), func(w http.ResponseWriter, r *http.Request) {
		// the only operation allowed on a gcode resource is to get the gcode
		// content for a job.
		switch r.Method {
		case "GET":
			srv.GetGCode(w, r)
		default:
			http.Error(w, "only GET is allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc(srv.route("/meshes/"), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			srv.GetMesh(w, r)
		default:
			http.Error(w, "only GET is allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

// path is a simple helper for constructing url paths by appending suffix to
// srv.Prefix.
func (srv *SnuggieServer) route(suffix string) string {
	return srv.Prefix + suffix
}

// trimPath trims the routing prefix from path and returns the suffix and the
// routing prefix.  The route must end in a slash '/'.  If path does not match
// the route an empty prefix is returned.
func (srv *SnuggieServer) trimPath(path, route string) (suffix, prefix string) {
	if !strings.HasSuffix(route, "/") {
		return "", ""
	}
	prefix = srv.route(route)
	suffix = strings.TrimPrefix(path, prefix)
	if len(suffix) == len(path) {
		return "", ""
	}
	return suffix, prefix
}

func (srv *SnuggieServer) GetGCode(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
; generated by Slic3r 1.1.7 on 2015-01-23 at 23:48:20

; perimeters extrusion width = 0.44mm
; infill extrusion width = 0.44mm
; solid infill extrusion width = 0.44mm
; top infill extrusion width = 0.44mm

G21 ; set units to millimeters
M107
M104 S195 ; set temperature
`)
}

func (srv *SnuggieServer) GetMesh(w http.ResponseWriter, r *http.Request) {
	id, _ := srv.trimPath(r.URL.Path, "/meshes/")
	path := queue[id]
	if path == "" {
		http.Error(w, "unknown mesh id", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, path)
}

func (srv *SnuggieServer) GetJob(w http.ResponseWriter, r *http.Request) {
	id, _ := srv.trimPath(r.URL.Path, "/jobs/")
	job, err := srv.lookupJob(id)
	if err != nil {
		http.Error(w, "lookup: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(job)
	if err != nil {
		log.Printf("http response: %v", err)
	}
}

func (srv *SnuggieServer) CreateJob(w http.ResponseWriter, r *http.Request) {
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

	job, err := srv.registerJob(meshfile, slicerBackend, preset)
	if err != nil {
		// TODO: distinguish unknown preset (Bad Request) from backend failure.
		http.Error(w, "registration failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonJob, err := json.Marshal(job)
	if err != nil {
		http.Error(w, "json didn't encode properly...Derp?\n"+err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(jsonJob)
}

func (srv *SnuggieServer) registerJob(meshfile multipart.File, slicerBackend string, preset string) (*slicerjob.Job, error) {
	job := slicerjob.New()

	//do stuff to the job.
	job.Status = slicerjob.Accepted
	job.Progress = 0.0
	job.URL = srv.url("/jobs/" + job.ID)

	//if location flag not set, default temp file location is used
	tmp, err := ioutil.TempFile(srv.DataDir, job.ID+"-")
	if err != nil {
		return nil, fmt.Errorf("meshfile not saved: %v", err)
	}
	defer tmp.Close()

	PutGCodeFile(job.ID, tmp.Name())

	err = PutJob(job.ID, job)
	if err != nil {
		return nil, err
	}
	url := srv.url("/meshes/" + job.ID)
	if srv.LocalConsumer {
		url = "file://" + tmp.Name()
	}
	err = srv.S.ScheduleSliceJob(job.ID, url, slicerBackend, preset)
	if err != nil {
		os.Remove(tmp.Name())
		//TODO:
		//DeleteGCodeFile(job.ID)
		//DeleteJob(job.ID)
		return nil, err
	}

	return job, nil
}

func (srv *SnuggieServer) lookupJob(id string) (*slicerjob.Job, error) {
	job, err := ViewJob(id)
	if err != nil {
		err := fmt.Errorf("Job not found with id: %v", id)
		return nil, err
	}

	if err != nil {
		err := fmt.Errorf("json unmarshal problem: %v", id)
		return nil, err
	} else {
		log.Println("mocking status")
		//mock progress
		job.Progress += 0.1
		if job.Progress >= 1.0 {
			job.Status = slicerjob.Complete
		}
		PutJob(id, job)
		//end mock progress
	}
	return job, nil
}

func (srv *SnuggieServer) url(pathquery string) string {
	return srv.Config["URL"] + srv.Prefix + pathquery
}

// JobDone stores the location of the successful output g-code for job id.  it
// returns the url of the gcode resource.
func (srv *SnuggieServer) JobDone(id, path string, err error) {
	if err != nil {
		log.Printf("FIXME -- failed job:%v err:%v", id, err)
		return
	}

	// TODO:
	// write the gcode path to the database

	log.Printf("completed job:%v gcode:%v", id, path)
}

// RunConsumers pops jobs off the queue, fetches remote mesh files, slices
// them, and makes the resulting gcode accessible over HTTP,
func (srv *SnuggieServer) RunConsumer() {
	for {
		job, err := srv.C.NextSliceJob()
		if err != nil {
			log.Printf("consumer: %v", err)
			return
		}
		type jobResult struct {
			path string
			err  error
		}
		joberr := make(chan jobResult, 1)
		go func() {
			// slice the file at job.MeshPath and save the gcode to a file.
			// send the out gcode's path over joberr so the call to srv.Done
			// can be serialized with any job cancelation.

			// TODO:
			// replace code below with call to external slicer that generates
			// gcode
			f, err := ioutil.TempFile(srv.DataDir, "snuggied-gcode-")
			if err != nil {
				joberr <- jobResult{err: err}
				return
			}
			defer func() {
				select {
				case joberr <- jobResult{path: f.Name()}:
				default:
				}
			}()
			defer func() {
				err := f.Close()
				if err != nil {
					select {
					case joberr <- jobResult{err: err}:
					default:
					}
				}
			}()
			fmt.Fprintf(f, "; this is g-code data\n")
			fmt.Fprintf(f, "; generated %v\n", time.Now)
			fmt.Fprintf(f, "; perimeters extrusion width = 0.44mm\n")
			fmt.Fprintf(f, "; infill extrusion width = 0.44mm\n")
			fmt.Fprintf(f, "; solid infill extrusion width = 0.44mm\n")
			fmt.Fprintf(f, "; top infill extrusion width = 0.44mm\n")
			fmt.Fprintf(f, "\n")
			fmt.Fprintf(f, "G21 ; set units to millimeters\n")
			fmt.Fprintf(f, "M107\n")
			fmt.Fprintf(f, "M104 S195 ; set temperature\n")
		}()
		select {
		case err := <-job.Cancel:
			job.Done("", err)
			// TODO: cleanup process
		case result := <-joberr:
			job.Done(result.path, result.err)
		}
	}
}

func main() {
	dataDir := flag.String("dataDir", "", "set meshfile save location")
	httpAddr := flag.String("http", ":8888", "address to serve traffic")
	flag.Parse()

	srv := &SnuggieServer{
		Config:  config,
		Prefix:  "/slicer",
		DataDir: *dataDir,
	}

	// register http handlers
	srv.RegisterHandlers(http.DefaultServeMux)

	// the scheduler/consumer for the server are implemented using an in-memory
	// queue.
	memq := MemoryQueue(srv.JobDone)
	srv.S, srv.C = memq, memq
	srv.LocalConsumer = true // use file:// locations instead of http://

	// BUG:
	// there is a race condition starting the queue consumer before serving
	// http traffic. slice jobs could be finished before the http server is
	// capable of serving the result. this would be most problematic if binding
	// the address fails.
	go srv.RunConsumer()
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
