package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/orch/cattle"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/longhorn-orc/util/daemon"
	"github.com/urfave/cli"
	"io"
	"sync"
)

func RunManager(c *cli.Context) error {
	man := New(cattle.New(c), WaitForDevice, MonitorVolume)

	go serveAPI(man)

	return daemon.WaitForExit()
}

type volumeManager struct {
	sync.Mutex
	mons map[string]io.Closer

	hostID  string
	orc     types.Orchestrator
	waitDev types.WaitForDevice
	monVol  types.MonitorVolume
}

func New(orc types.Orchestrator, waitDev types.WaitForDevice, monVol types.MonitorVolume) types.VolumeManager {
	hostID, err := orc.GetThisHostID()
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "failed to get this host ID from the orchestrator"))
	}
	return &volumeManager{
		mons:    map[string]io.Closer{},
		hostID:  hostID,
		orc:     orc,
		waitDev: waitDev,
		monVol:  monVol,
	}
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

func (man *volumeManager) startMonitoring(name string) {
	man.Lock()
	defer man.Unlock()
	if man.mons[name] == nil {
		man.mons[name] = man.monVol(name, man)
	}
}

func (man *volumeManager) stopMonitoring(name string) {
	man.Lock()
	defer man.Unlock()
	if mon := man.mons[name]; mon != nil {
		mon.Close()
		delete(man.mons, name)
	}
}

func (man *volumeManager) Attach(name string) error {
	vol, err := man.Get(name)
	if err != nil {
		return err
	}
	if vol.Controller != nil && vol.Controller.Running {
		if vol.Controller.HostID == man.hostID {
			man.startMonitoring(name)
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
		} else if mostRecentBadReplica == nil || replica.BadTimestamp.After(*mostRecentBadReplica.BadTimestamp) {
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
	if err := man.waitDev(vol.Name); err != nil {
		return errors.Wrapf(err, "error waiting for device for volume '%s'", vol.Name)
	}

	man.startMonitoring(name)
	return nil
}

func (man *volumeManager) Detach(name string) error {
	return nil
}

func (man *volumeManager) CheckVolume(volume *types.VolumeInfo) error {
	return nil
}

func (man *volumeManager) Cleanup(volume *types.VolumeInfo) error {
	return nil
}
