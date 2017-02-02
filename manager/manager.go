package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/controller"
	"github.com/rancher/longhorn-orc/orch/cattle"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/longhorn-orc/util/daemon"
	"github.com/urfave/cli"
	"io"
	"sync"
)

func RunManager(c *cli.Context) error {
	man := New(cattle.New(c), WaitForDevice, Monitor, controller.New)

	go serveAPI(man)

	return daemon.WaitForExit()
}

type volumeManager struct {
	sync.Mutex

	monitors map[string]io.Closer

	hostID        string
	orc           types.Orchestrator
	waitForDevice types.WaitForDevice
	monitor       types.Monitor
	getController types.GetController
}

func New(orc types.Orchestrator, waitDev types.WaitForDevice, monitor types.Monitor, getController types.GetController) types.VolumeManager {
	hostID, err := orc.GetThisHostID()
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "failed to get this host ID from the orchestrator"))
	}
	return &volumeManager{
		monitors:      map[string]io.Closer{},
		hostID:        hostID,
		orc:           orc,
		waitForDevice: waitDev,
		monitor:       monitor,
		getController: getController,
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

func (man *volumeManager) startMonitoring(controllerInfo *types.ControllerInfo) {
	man.Lock()
	defer man.Unlock()
	if man.monitors[controllerInfo.Name] == nil {
		man.monitors[controllerInfo.Name] = man.monitor(man.getController(controllerInfo), man)
	}
}

func (man *volumeManager) stopMonitoring(controllerInfo *types.ControllerInfo) {
	man.Lock()
	defer man.Unlock()
	if mon := man.monitors[controllerInfo.Name]; mon != nil {
		mon.Close()
		delete(man.monitors, controllerInfo.Name)
	}
}

func (man *volumeManager) Attach(name string) error {
	vol, err := man.Get(name)
	if err != nil {
		return err
	}
	if vol.Controller != nil && vol.Controller.Running {
		if vol.Controller.HostID == man.hostID {
			man.startMonitoring(vol.Controller)
			return nil
		}
		return errors.Errorf("volume already attached to host '%s'", vol.Controller.HostID)
	}
	replicas := []*types.ReplicaInfo{}
	var mostRecentBadReplica *types.ReplicaInfo
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
	controllerInfo, err := man.orc.CreateController(vol.Name, man.hostID, replicas)
	if err != nil {
		return errors.Wrapf(err, "failed to start the controller for volume '%s'", vol.Name)
	}
	if err := man.waitForDevice(vol.Name); err != nil {
		return errors.Wrapf(err, "error waiting for device for volume '%s'", vol.Name)
	}

	man.startMonitoring(controllerInfo)
	return nil
}

func (man *volumeManager) Detach(name string) error {
	return nil
}

func (man *volumeManager) CheckController(controller types.Controller) error {
	// TODO impl monitoring logic here
	return nil
}

func (man *volumeManager) Cleanup(volume *types.VolumeInfo) error {
	return nil
}
