package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/types"
	"io"
	"time"
)

var (
	MonitoringPeriod = time.Second * 5
)

type monitorChan chan<- Event

func (mc monitorChan) Close() error {
	defer func() {
		recover()
	}()
	close(mc)
	return nil
}

func Monitor(volume *types.VolumeInfo, man types.VolumeManager) io.Closer {
	ch := make(chan Event)
	go monitor(volume, man, ch)
	return monitorChan(ch)
}

func monitor(volume *types.VolumeInfo, man types.VolumeManager, ch chan Event) {
	defer NewTicker(MonitoringPeriod, ch).Start().Stop()
	for range ch {
		if err := man.CheckController(volume); err != nil {
			logrus.Errorf("%+v", errors.Wrapf(err, "monitoring: failed checking controller '%s', detaching volume", volume.Name))
			if err := man.Detach(volume.Name); err != nil {
				logrus.Errorf("%+v", errors.Wrapf(err, "monitoring: error detaching failed volume '%s'", volume.Name))
			}
			break
		}
	}
}
