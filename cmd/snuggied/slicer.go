package main

import (
	"fmt"
	"io"
	"os"
)

type SlicerCmd struct {
	Bin  string
	Args []string
	Log  io.Writer
}

type Slicer interface {
	SlicerCmd() SlicerCmd
}

func Run(s Slicer, kill <-chan struct{}) error {
	return fmt.Errorf("x")
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
		Bin:  bin,
		Args: args,
		Log:  os.Stderr,
	}
}
