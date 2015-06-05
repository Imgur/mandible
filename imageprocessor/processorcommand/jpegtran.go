package processorcommand

import (
	"fmt"
)

func Jpegtran(filename string) (string, error) {
	outfile := fmt.Sprintf("%s_opti", filename)

	args := []string{
		"-copy",
		"all",
		"-optimize",
		"-outfile",
		outfile,
		filename,
	}

	err := runProcessorCommand("jpegtran", args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}
