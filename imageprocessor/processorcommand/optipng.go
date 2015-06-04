package processorcommand

import (
	"fmt"
)

func Optipng(filename string) (string, error) {
	outfile := fmt.Sprintf("%s_opi", filename)

	args := []string{
		"-fix",
		"-out",
		outfile,
		filename,
	}

	err := runProcessorCommand("optipng", args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}