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

func parseReplica(s string) (*types.ReplicaInfo, error) {
	fields := strings.Fields(s)
	if len(fields) < 2 {
		return nil, errors.Errorf("cannot parse line `%s`", s)
	}
	mode, ok := modes[fields[1]]
	if !ok {
		mode = types.ERR
	}
	if mode == types.RW && len(fields) == 2 {
		mode = types.ERR // TODO it's a workaround, remove when fixed in longhorn
	}
	return &types.ReplicaInfo{
		InstanceInfo: types.InstanceInfo{
			Address: fields[0],
		},
		Mode: mode,
	}, nil
}

func (c *controller) GetReplicaStates() ([]*types.ReplicaInfo, error) {
	replicas := []*types.ReplicaInfo{}
	cancel := make(chan interface{})
	defer close(cancel)
	lineCh, cliErrCh := util.CmdOutLines(exec.Command("longhorn", "--url", c.url, "ls"), cancel)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	parsingErrCh := make(chan error)
	go func() {
		defer wg.Done()
		defer close(parsingErrCh)
		for s := range lineCh {
			if strings.HasPrefix(s, "ADDRESS") {
				continue
			}
			replica, err := parseReplica(s)
			if err != nil {
				parsingErrCh <- errors.Wrapf(err, "error parsing replica status from `%s`", s)
				break
			}
			replicas = append(replicas, replica)
		}
	}()
	for err := range parsingErrCh {
		return nil, err
	}
	for err := range cliErrCh {
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
