package processorcommand

func StripMetadata(filename string) error {
	args := []string{
		"-all=",
		"--icc_profile:all",
		"-overwrite_original",
		filename,
	}

	err := runProcessorCommand("exiftool", args)
	if err != nil {
		return err
	}

	return nil
}
