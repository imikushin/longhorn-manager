package manager

import "github.com/rancher/longhorn-orc/types"

type controllerHost struct {
}

var thisHost = &controllerHost{}

func ThisHost() types.ControllerHost {
	return thisHost
}

func (h *controllerHost) WaitForDevice(name string) error {
	// TODO impl
	return nil
}