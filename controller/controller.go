package controller

import (
	"fmt"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/longhorn/sync"
)

type controller struct {
	url string
}

func New(volume *types.VolumeInfo) types.Controller {
	return &controller{
		url: fmt.Sprintf("http://%s:9501", volume.Controller.Address),
	}
}

func (c *controller) Name() string {
	return "DUMMY"
}

func (c *controller) GetReplicaStates() ([]*types.ReplicaInfo, error) {
	return nil, nil
}

func (c *controller) AddReplica(replica *types.ReplicaInfo) error {
	return sync.NewTask(c.url).AddReplica(replica.Address)
}

func (c *controller) RemoveReplica(replica *types.ReplicaInfo) error {
	return nil
}
