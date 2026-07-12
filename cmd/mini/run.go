package main

import "fmt"

func run(pathConfig string, containerID string) error {
	cmd, err := create(pathConfig)
	if err != nil {
		return fmt.Errorf("there was an erro while creating the container: %w", err)
	}

	err = start(containerID)
	if err != nil {
		return fmt.Errorf("there was an error while starting the container: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("there was an error while waiting for the container to finish: %w", err)
	}

	return nil
}
