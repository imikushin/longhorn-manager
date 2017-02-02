package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/types"
	"io"
)

type monitorChan chan<- struct{}

func (mc monitorChan) Close() error {
	close(mc)
	return nil
}

func MonitorVolume(name string, man types.VolumeManager) io.Closer {
	ch := make(chan struct{})
	go doMonitorVolume(name, man, ch)
	return monitorChan(ch)
}

func doMonitorVolume(name string, man types.VolumeManager, ch chan struct{}) {
	// TODO perform checks in regular intervals
	for range ch {
		vol, err := man.Get(name)
		if err != nil {
			logrus.Errorf("%+v", errors.Wrapf(err, "monitoring: error getting volume '%s'", name))
			continue
		}
		if err := man.CheckVolume(vol); err != nil {
			logrus.Errorf("%+v", errors.Wrapf(err, "monitoring: error checking volume '%s'", name))
		}
	}
}
