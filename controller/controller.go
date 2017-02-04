package controller

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/longhorn/controller/client"
	"github.com/rancher/longhorn/sync"
)

type controller struct {
	name   string
	url    string
	client *client.ControllerClient
}

func New(volume *types.VolumeInfo) types.Controller {
	url := fmt.Sprintf("http://%s:9501", volume.Controller.Address)
	return &controller{
		name:   volume.Name,
		url:    url,
		client: client.NewControllerClient(url),
	}
}

func (c *controller) Name() string {
	return c.name
}

func (c *controller) GetReplicaStates() ([]*types.ReplicaInfo, error) {
	return nil, nil
}

func (c *controller) AddReplica(replica *types.ReplicaInfo) error {
	err := sync.NewTask(c.url).AddReplica(replica.Address)
	return errors.Wrapf(err, "failed to add replica '%s' IP='%s' to controller '%s'", replica.ID, replica.Address, c.name)
}

func (c *controller) RemoveReplica(replica *types.ReplicaInfo) error {
	_, err := c.client.DeleteReplica(replica.Address)
	return errors.Wrapf(err, "failed to rm replica '%s', IP='%s' from controller '%s'", replica.ID, replica.Address, c.name)
}
