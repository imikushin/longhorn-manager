package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/longhorn-orc/types"
	"io"
	"time"
)

var (
	MonitoringPeriod     = time.Second * 5
	MonitoringMaxRetries = 3
)

type monitorChan chan<- Event

func (mc monitorChan) Close() error {
	defer func() {
		recover()
	}()
	close(mc)
	return nil
}

func Monitor(getController types.GetController) types.Monitor {
	return func(volume *types.VolumeInfo, man types.VolumeManager) io.Closer {
		ch := make(chan Event)
		go monitor(getController(volume.Controller), volume, man, ch)
		return monitorChan(ch)
	}
}

func monitor(ctrl types.Controller, volume *types.VolumeInfo, man types.VolumeManager, ch chan Event) {
	ticker := NewTicker(MonitoringPeriod, ch)
	defer ticker.Start().Stop()
	failedAttempts := 0
	for range ch {
		if err := func() error {
			defer ticker.Stop().Start()
			if err := man.CheckController(ctrl, volume); err != nil {
				if failedAttempts++; failedAttempts > MonitoringMaxRetries {
					return errors.Wrapf(err, "failed checking controller '%s'", volume.Name)
				}
				logrus.Warnf("%+v", errors.Wrapf(err, "failed checking controller '%s', going to retry", volume.Name))
				return nil
			}
			failedAttempts = 0
			return nil
		}(); err != nil {
			close(ch)
			logrus.Error(errors.Wrapf(err, "detaching volume"))
			if err := man.Detach(volume.Name); err != nil {
				logrus.Errorf("%+v", errors.Wrapf(err, "error detaching failed volume '%s'", volume.Name))
			}
		}
	}
}
