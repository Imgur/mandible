package processorcommand

import (
	"errors"
	"fmt"
)

func Jpegtran(filename string) (string, error) {
	outfile := fmt.Sprintf("%s_opi", filename)

	args := []string{
		"-copy",
		"-all",
		"-optimize",
		filename,
		">",
		outfile,
	}

	err := runProcessCommand("jpegtran", args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}