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

func MonitorVolume(name string, man types.VolumeManager) io.Closer {
	ch := make(chan Event)
	go doMonitorVolume(name, man, ch)
	return monitorChan(ch)
}

func doMonitorVolume(name string, man types.VolumeManager, ch chan Event) {
	ticker := NewTicker(MonitoringPeriod, ch)
	go Send(ch, ticker.NewTick())
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
