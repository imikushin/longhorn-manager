package controller

import (
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
	url := volume.Controller.Address
	return &controller{
		name:   volume.Name,
		url:    url,
		client: client.NewControllerClient(url),
	}
}

func (c *controller) Name() string {
	return c.name
}

var modes = map[string]types.ReplicaMode{
	"RW":  types.RW,
	"WO":  types.WO,
	"ERR": types.ERR,
}

func (c *controller) GetReplicaStates() ([]*types.ReplicaInfo, error) {
	rs, err := c.client.ListReplicas()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get replicas from controller '%s'", c.name)
	}
	replicas := make([]*types.ReplicaInfo, len(rs))
	for i, r := range rs {
		replicas[i] = &types.ReplicaInfo{
			InstanceInfo: types.InstanceInfo{
				Address: r.Address,
			},
			Mode: modes[r.Mode],
		}
	}
	return replicas, nil
}

func (c *controller) AddReplica(replica *types.ReplicaInfo) error {
	err := sync.NewTask(c.url).AddReplica(replica.Address)
	return errors.Wrapf(err, "failed to add replica address='%s' to controller '%s'", replica.Address, c.name)
}

func (c *controller) RemoveReplica(replica *types.ReplicaInfo) error {
	_, err := c.client.DeleteReplica(replica.Address)
	return errors.Wrapf(err, "failed to rm replica address='%s' from controller '%s'", replica.Address, c.name)
}
