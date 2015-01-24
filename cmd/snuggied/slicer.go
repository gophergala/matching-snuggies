package main

type Slicer struct {
	Bin  string
	Args []string
}

func (s *Slicer) Start() error {
}

func Slic3r(bin, outfile, infile string) *Slicer {
	if bin == "" {
		bin = "slic3r"
	}
	var args []string
	if outfile != "" {
		args = append(args, "-o", outfile)
	}
	if infile != "" {
		args = append(args, infile)
	}
	return &Slicer{
		Bin:  bin,
		Args: []string{"-o", outfile, infile},
	}
}
