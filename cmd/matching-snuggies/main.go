package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gophergala/matching-snuggies/slicerjob"
)

func main() {
	server := flag.String("server", "localhost:8888", "snuggied server address")
	slicerBackend := flag.String("backend", "slic3r", "backend slicer")
	slicerPreset := flag.String("preset", "hiQ", "specify a configuration preset for the backend")
	gcodeDest := flag.String("o", "", "specify an output gcode filename")
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("missing argument: mesh file")
	}
	meshpath := flag.Arg(0)

	client := &Client{
		ServerAddr: *server,
	}

	// send files to the slicer to be printed and poll the slicer until the job
	// has completed.
	log.Printf("sending file(s) to snuggied server at %v", *server)
	job, err := client.SliceFile(*slicerBackend, *slicerPreset, meshpath)
	if err != nil {
		log.Fatalf("sending files: %v", err)
	}
	tick := time.After(100 * time.Millisecond)
	for job.Status != "complete" {
		// TODO: retry with exponential backoff on network failure
		select {
		case <-tick:
			job, err = client.SlicerStatus(job)
			if err != nil {
				log.Fatalf("waiting: %v", err)
			}
		}
	}

	// download gcode from the slicer and write to the specified file.
	log.Printf("retreiving gcode file")
	r, err := client.GCode(job)
	if err != nil {
		log.Fatalf("gcode: %v", err)
	}
	defer r.Close()
	var f *os.File
	if *gcodeDest == "" {
		f = os.Stdout
	} else {
		f, err = os.Create(*gcodeDest)
		if err != nil {
			log.Panic(err)
		}
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}()
	_, err = io.Copy(f, r)
	if err != nil {
		log.Panic(err)
	}
}

type Client struct {
	Client     *http.Client
	ServerAddr string
	HTTPS      bool
}

// SliceFiles tells the server to slice the specified paths.
func (c *Client) SliceFile(backend, preset string, path string) (*slicerjob.Job, error) {
	// check that a mesh file is given as the first argument and open it
	// so it may to encode in the form.
	if !IsMeshFile(path) {
		log.Fatalf("path is not a mesh file: %v", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// write the multipart form out to a temporary file.  the temporary
	// file is closed and unlinked when the function terminates.
	tmp, err := ioutil.TempFile("", "matching-snuggies-post-")
	if err != nil {
		return nil, fmt.Errorf("tempfile: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	bodyw := multipart.NewWriter(tmp)
	err = c.writeJobForm(bodyw, backend, preset, path, f)
	if err != nil {
		return nil, fmt.Errorf("tempfile: %v", err)
	}

	// seek back to the beginning of the form and POST it to the slicer
	// server.  decode a slicerjob.Job from successful responses.
	var job *slicerjob.Job
	_, err = tmp.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("tempfile: %v", err)
	}
	url := c.url("/slicer/jobs/")
	log.Printf("POST %v", url)
	resp, err := c.client().Post(url, bodyw.FormDataContentType(), tmp)
	if err != nil {
		return nil, fmt.Errorf("POST /slicer/jobs/: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		p, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("response: %q (%v)", p, http.StatusText(resp.StatusCode))
	}
	err = json.NewDecoder(resp.Body).Decode(&job)
	if err != nil {
		return nil, fmt.Errorf("response: %v", err)
	}
	return job, nil
}

func (c *Client) writeJobForm(w *multipart.Writer, backend, preset, filename string, r io.Reader) error {
	err := w.WriteField("slicer", backend)
	if err != nil {
		return err
	}
	err = w.WriteField("preset", preset)
	if err != nil {
		return err
	}
	file, err := w.CreateFormFile("meshfile", filepath.Base(filename))
	if err != nil {
		return err
	}
	_, err = io.Copy(file, r)
	if err != nil {
		return err
	}
	return w.Close()
}

// SlicerStatus returns a current copy of the provided job.
func (c *Client) SlicerStatus(job *slicerjob.Job) (*slicerjob.Job, error) {
	if job.ID == "" {
		return nil, fmt.Errorf("job missing id")
	}
	var jobcurr *slicerjob.Job
	url := c.url("/slicer/jobs/" + job.ID)
	log.Printf("POST %v", url)
	resp, err := c.client().Get(url)
	if err != nil {
		return nil, fmt.Errorf("POST /slicer/jobs/: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		p, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("response: %q (%v)", p, http.StatusText(resp.StatusCode))
	}
	err = json.NewDecoder(resp.Body).Decode(&jobcurr)
	if err != nil {
		return nil, fmt.Errorf("response: %v", err)
	}
	return jobcurr, nil
}

// GCode requests the gcode for job.
func (c *Client) GCode(job *slicerjob.Job) (io.ReadCloser, error) {
	var jobcurr *slicerjob.Job
	url := c.url("/slicer/gcodes/" + jobcurr.ID)
	log.Printf("POST %v", url)
	resp, err := c.client().Get(url)
	if err != nil {
		return nil, fmt.Errorf("POST /slicer/jobs/: %v", err)
	}
	if resp.StatusCode != 200 {
		p, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("response: %q (%v)", p, http.StatusText(resp.StatusCode))
	}
	return resp.Body, nil
}

func (c *Client) client() *http.Client {
	if c.Client == nil {
		return http.DefaultClient
	}
	return c.Client
}

func (c *Client) url(pathquery string) string {
	pathquery = strings.TrimPrefix(pathquery, "/")
	scheme := "http"
	if c.HTTPS {
		scheme = "https"
	}
	return scheme + "://" + c.ServerAddr + "/" + pathquery
}

var meshExts = map[string]bool{
	".stl": true,
	".amf": true,
}

func IsMeshFile(path string) bool {
	return meshExts[filepath.Ext(path)]
}
