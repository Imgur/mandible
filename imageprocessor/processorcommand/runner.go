package processorcommand

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"time"
)

func runProcessorCommand(command string, args []string) error {
	cmd := exec.Command(command, args...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	cmd.Start()

	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Duration(60) * time.Second):
		killCmd(cmd)
		<-cmdDone
		return errors.New("Command timed out")
	case err := <-cmdDone:
		if err != nil {
			log.Println(stderr.String())
		}

		return err
	}
}

func killCmd(cmd *exec.Cmd) {
	if err := cmd.Process.Kill(); err != nil {
		log.Printf("Failed to kill command: %v", err)
	}
}
