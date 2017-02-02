package controller

import "github.com/rancher/longhorn-orc/types"

type controller struct {
}

func New(controllerInfo *types.ControllerInfo) types.Controller {
	return nil
}

func (c *controller) Name() string {
	return "DUMMY"
}

func (c *controller) GetReplicaStates() ([]*types.ReplicaInfo, error) {
	return nil, nil
}
