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

	monitors       map[string]io.Closer
	addingReplicas map[string]int

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
		monitors:       map[string]io.Closer{},
		addingReplicas: map[string]int{},

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

func (man *volumeManager) startMonitoring(volume *types.VolumeInfo) {
	man.Lock()
	defer man.Unlock()
	if man.monitors[volume.Name] == nil {
		man.monitors[volume.Name] = man.monitor(volume, man)
	}
}

func (man *volumeManager) stopMonitoring(volume *types.VolumeInfo) {
	man.Lock()
	defer man.Unlock()
	if mon := man.monitors[volume.Name]; mon != nil {
		mon.Close()
		delete(man.monitors, volume.Name)
	}
}

func (man *volumeManager) Attach(name string) error {
	volume, err := man.Get(name)
	if err != nil {
		return err
	}
	if volume.Controller != nil && volume.Controller.Running {
		if volume.Controller.HostID == man.hostID {
			man.startMonitoring(volume)
			return nil
		}
		return errors.Errorf("volume already attached to host '%s'", volume.Controller.HostID)
	}
	replicas := []*types.ReplicaInfo{}
	var mostRecentBadReplica *types.ReplicaInfo
	for _, replica := range volume.Replicas {
		if replica.Running {
			if err := man.orc.StopContainer(replica.ID); err != nil {
				return errors.Wrapf(err, "failed to stop replica '%s' on host '%s' for volume '%s'", replica.ID, replica.HostID, volume.Name)
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
		return errors.Errorf("no replicas to start the controller for volume '%s'", volume.Name)
	}
	for _, replica := range replicas {
		if err := man.orc.StartContainer(replica.ID); err != nil {
			return errors.Wrapf(err, "failed to start replica '%s' on host '%s' for volume '%s'", replica.ID, replica.HostID, volume.Name)
		}
	}
	controllerInfo, err := man.orc.CreateController(volume.Name, man.hostID, replicas)
	if err != nil {
		return errors.Wrapf(err, "failed to start the controller for volume '%s'", volume.Name)
	}
	if err := man.waitForDevice(volume.Name); err != nil {
		return errors.Wrapf(err, "error waiting for device for volume '%s'", volume.Name)
	}

	volume.Controller = controllerInfo
	man.startMonitoring(volume)
	return nil
}

func (man *volumeManager) Detach(name string) error {
	return nil
}

func (man *volumeManager) createAndAddReplicaToController(volumeName string, ctrl types.Controller) error {
	replica, err := man.orc.CreateReplica(volumeName)
	if err != nil {
		return errors.Wrapf(err, "failed to create a replica for volume '%s'", volumeName)
	}
	go func() {
		man.addingReplicasCount(volumeName, 1)
		defer man.addingReplicasCount(volumeName, -1)
		if err := ctrl.AddReplica(replica); err != nil {
			logrus.Errorf("%+v", errors.Wrapf(err, "failed to add replica '%s' to volume '%s'", replica.ID, volumeName))
		}
	}()
	return nil
}

func (man *volumeManager) addingReplicasCount(name string, add int) int {
	man.Lock()
	defer man.Unlock()
	count := man.addingReplicas[name] + add
	man.addingReplicas[name] = count
	return count
}

func (man *volumeManager) CheckController(volume *types.VolumeInfo) error {
	ctrl := man.getController(volume.Controller)
	replicas, err := ctrl.GetReplicaStates()
	if err != nil {
		return errors.Wrapf(err, "error getting replica states for volume '%s'", volume.Name)
	}
	goodReplicas := []*types.ReplicaInfo{}
	woReplicas := []*types.ReplicaInfo{}
	removeErrs := make(chan error, len(replicas))
	removeOps := &sync.WaitGroup{}
	for _, replica := range replicas {
		switch replica.State {
		case types.RW:
			goodReplicas = append(goodReplicas, replica)
		case types.WO:
			woReplicas = append(woReplicas, replica)
		case types.ERR:
			removeOps.Add(1)
			go func(replica *types.ReplicaInfo) {
				defer removeOps.Done()
				if err := ctrl.RemoveReplica(replica); err != nil {
					removeErrs <- errors.Wrapf(err, "failed to remove ERR replica '%s' from volume '%s'", replica.Address, volume.Name)
					return
				}
				if err := man.orc.MarkBadReplica(replica); err != nil {
					removeErrs <- errors.Wrapf(err, "failed to mark replica '%s' bad for volume '%s'", replica.Address, volume.Name)
				}
			}(replica)
		}
	}
	go func() {
		removeOps.Wait()
		close(removeErrs)
	}()
	for err := range removeErrs {
		return err
	}
	if len(goodReplicas) == 0 {
		logrus.Errorf("volume '%s' has no more good replicas, shutting it down", volume.Name)
		return man.Detach(volume.Name)
	}

	addingReplicas := man.addingReplicasCount(volume.Name, 0)
	if len(goodReplicas)+len(woReplicas)+addingReplicas < volume.NumberOfReplicas {
		if err := man.createAndAddReplicaToController(volume.Name, ctrl); err != nil {
			return err
		}
	}
	if len(goodReplicas)+len(woReplicas)+addingReplicas > volume.NumberOfReplicas {
		logrus.Errorf("volume '%s' has more replicas than needed: has %v, needs %v", volume.Name, len(goodReplicas), volume.NumberOfReplicas)
	}

	return nil
}

func (man *volumeManager) Cleanup(volume *types.VolumeInfo) error {
	return nil
}
