package controller

import "github.com/rancher/longhorn-orc/types"

type controller struct {
}

func New(volume *types.VolumeInfo) types.Controller {
	return nil
}

func (c *controller) Name() string {
	return "DUMMY"
}

func (c *controller) GetReplicaStates() ([]*types.ReplicaInfo, error) {
	return nil, nil
}

func (c *controller) AddReplica(replica *types.ReplicaInfo) error {
	return nil
}

func (c *controller) RemoveReplica(replica *types.ReplicaInfo) error {
	return nil
}
