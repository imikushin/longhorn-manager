package manager

import (
	"github.com/rancher/longhorn-orc/orch/cattle"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/longhorn-orc/util/daemon"
	"github.com/urfave/cli"
)

func RunManager(c *cli.Context) error {
	orch := cattle.New(c)
	man := New(orch)

	go serveAPI(man)

	return daemon.WaitForExit()
}

type volumeManager struct {
	orch types.Orchestrator
}

func New(orch types.Orchestrator) types.VolumeManager {
	return &volumeManager{orch: orch}
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
	return nil
}

func (man *volumeManager) Detach(name string) error {
	return nil
}

func (man *volumeManager) MonitorVolume(volume *types.VolumeInfo) {
}

func (man *volumeManager) Cleanup(volume *types.VolumeInfo) {
}
