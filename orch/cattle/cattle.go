package cattle

import (
	"github.com/rancher/longhorn-orc/types"
	"github.com/urfave/cli"
)

type cattleOrc struct {
}

func New(c *cli.Context) types.Orchestrator {
	return &cattleOrc{}
}

func (orc *cattleOrc) CreateVolumeRecord(volume *types.VolumeInfo) (*types.VolumeInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) DeleteVolumeRecord(volumeName string) error {
	return nil
}

func (orc *cattleOrc) GetVolumeRecord(volumeName string) (*types.VolumeInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) MarkBadReplica(containerID string) error {
	return nil
}

func (orc *cattleOrc) CreateController(volumeName, hostID string, replicas []*types.ContainerInfo) (*types.ContainerInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) CreateReplica(volume *types.VolumeInfo) (*types.ContainerInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) StartContainer(containerID string) error {
	return nil
}

func (orc *cattleOrc) StopContainer(containerID string) error {
	return nil
}

func (orc *cattleOrc) RemoveContainer(containerID string) error {
	return nil
}