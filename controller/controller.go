package controller

import (
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/longhorn-orc/util"
	"os/exec"
	"strings"
	"sync"
)

type controller struct {
	name string
	url  string
}

func New(volume *types.VolumeInfo) types.Controller {
	url := volume.Controller.Address
	return &controller{
		name: volume.Name,
		url:  url,
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
	replicas := []*types.ReplicaInfo{}
	cancel := make(chan interface{})
	defer close(cancel)
	lineCh, errCh := util.CmdOutLines(exec.Command("longhorn", "--url", c.url, "ls"), cancel)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for s := range lineCh {
			if strings.HasPrefix(s, "ADDRESS") {
				continue
			}
			fields := strings.Fields(s)
			replicas = append(replicas, &types.ReplicaInfo{
				InstanceInfo: types.InstanceInfo{
					Address: fields[0],
				},
				Mode: modes[fields[1]],
			})
		}
	}()
	for err := range errCh {
		return nil, err
	}

	wg.Wait()
	return replicas, nil
}

func (c *controller) AddReplica(replica *types.ReplicaInfo) error {
	err := exec.Command("longhorn", "--url", c.url, "add", replica.Address).Run()
	return errors.Wrapf(err, "failed to add replica address='%s' to controller '%s'", replica.Address, c.name)
}

func (c *controller) RemoveReplica(replica *types.ReplicaInfo) error {
	err := exec.Command("longhorn", "--url", c.url, "rm", replica.Address).Run()
	return errors.Wrapf(err, "failed to rm replica address='%s' from controller '%s'", replica.Address, c.name)
}
