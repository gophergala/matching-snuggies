package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

type SlicerCmd struct {
	Bin    string
	Args   []string
	OutLog io.Writer
	ErrLog io.Writer
}

type Slicer interface {
	SlicerCmd() *SlicerCmd
}

func Run(s Slicer, kill <-chan error) error {
	scmd := s.SlicerCmd()
	cmd := exec.Command(scmd.Bin, scmd.Args...)
	cmd.Stdout = scmd.OutLog
	cmd.Stderr = scmd.ErrLog
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("%s: %v", scmd.Bin, err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	for {
		select {
		case err := <-done:
			return err
		case err := <-kill:
			if errkill := cmd.Process.Kill(); err != nil {
				// we couldn't kill the process. don't exit the loop.
				log.Printf("kill: %v", errkill)
				continue
			}
			return err
		}
	}
}

type Slic3r struct {
	Bin        string
	ConfigPath string
	OutPath    string
	InPath     string
}

func (s *Slic3r) SlicerCmd() *SlicerCmd {
	bin := s.Bin
	if bin == "" {
		bin = "slic3r"
	}
	var args []string
	config := s.ConfigPath
	if config != "" {
		args = append(args, config)
	}
	out := s.OutPath
	if out != "" {
		args = append(args, "-o", out)
	}
	in := s.InPath
	if in != "" {
		args = append(args, in)
	}
	return &SlicerCmd{
		Bin:    bin,
		Args:   args,
		OutLog: os.Stderr,
		ErrLog: os.Stderr,
	}
}
