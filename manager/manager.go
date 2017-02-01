package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/orch/cattle"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/longhorn-orc/util/daemon"
	"github.com/urfave/cli"
)

func RunManager(c *cli.Context) error {
	orch := cattle.New(c)
	man := New(orch, ThisHost())

	go serveAPI(man)

	return daemon.WaitForExit()
}

type volumeManager struct {
	orc    types.Orchestrator
	host   types.ControllerHost
	hostID string
}

func New(orc types.Orchestrator, host types.ControllerHost) types.VolumeManager {
	hostID, err := orc.GetThisHostID()
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "failed to get this host ID from the orchestrator"))
	}
	return &volumeManager{orc: orc, hostID: hostID}
}

func (man *volumeManager) Create(volume *types.VolumeInfo) (*types.VolumeInfo, error) {
	return nil, nil
}

func (man *volumeManager) Delete(name string) error {
	return nil
}

func (man *volumeManager) Get(name string) (*types.VolumeInfo, error) {
	return nil, nil
}

func (man *volumeManager) Attach(name string) error {
	vol, err := man.Get(name)
	if err != nil {
		return err
	}
	if vol.Controller != nil && vol.Controller.Running {
		if vol.Controller.HostID == man.hostID {
			return nil
		}
		return errors.Errorf("volume already attached to host '%s'", vol.Controller.HostID)
	}
	replicas := []*types.ContainerInfo{}
	var mostRecentBadReplica *types.ContainerInfo
	for _, replica := range vol.Replicas {
		if replica.Running {
			if err := man.orc.StopContainer(replica.ID); err != nil {
				return errors.Wrapf(err, "failed to stop replica '%s' on host '%s' for volume '%s'", replica.ID, replica.HostID, vol.Name)
			}
		}
		if replica.BadTimestamp == nil {
			replicas = append(replicas, replica)
		} else if mostRecentBadReplica == nil || replica.BadTimestamp.After(mostRecentBadReplica.BadTimestamp) {
			mostRecentBadReplica = replica
		}
	}
	if len(replicas) == 0 && mostRecentBadReplica != nil {
		replicas = append(replicas, mostRecentBadReplica)
	}
	if len(replicas) == 0 {
		return errors.Errorf("no replicas to start the controller for volume '%s'", vol.Name)
	}
	for _, replica := range replicas {
		if err := man.orc.StartContainer(replica.ID); err != nil {
			return errors.Wrapf(err, "failed to start replica '%s' on host '%s' for volume '%s'", replica.ID, replica.HostID, vol.Name)
		}
	}
	if _, err := man.orc.CreateController(vol.Name, man.hostID, replicas); err != nil {
		return errors.Wrapf(err, "failed to start the controller for volume '%s'", vol.Name)
	}
	if err := man.host.WaitForDevice(vol.Name); err != nil {
		return errors.Wrapf(err, "error waiting for device for volume '%s'", vol.Name)
	}

	// TODO Start monitoring the controller (see Maintain volume health)

	return nil
}

func (man *volumeManager) Detach(name string) error {
	return nil
}

func (man *volumeManager) MonitorVolume(volume *types.VolumeInfo) {
}

func (man *volumeManager) Cleanup(volume *types.VolumeInfo) {
}
