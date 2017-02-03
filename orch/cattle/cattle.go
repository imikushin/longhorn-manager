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

func (orc *cattleOrc) MarkBadReplica(replica *types.ReplicaInfo) error {
	return nil
}

func (orc *cattleOrc) CreateController(volumeName, hostID string, replicas []*types.ReplicaInfo) (*types.ControllerInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) CreateReplica(volumeName string) (*types.ReplicaInfo, error) {
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

func (orc *cattleOrc) GetThisHostID() (string, error) {
	return "DUMMY", nil
}
