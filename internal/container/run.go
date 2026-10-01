package container

import (
	"fmt"

	"containerruntime/internal/cgroup"
)

func (container *Container) run(cgroup *cgroup.CgroupContainer) error {
	cmd, err := container.create(cgroup)
	if err != nil {
		return fmt.Errorf("there was an erro while creating the container: %w", err)
	}

	err = container.start()
	if err != nil {
		return fmt.Errorf("there was an error while starting the container: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("there was an error while waiting for the container to finish: %w", err)
	}

	return nil
}
