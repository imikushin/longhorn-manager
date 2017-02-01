package types

import (
	"time"
)

const (
	DefaultNumberOfReplicas    = 2
	DefaultStaleReplicaTimeout = time.Hour * 16 * 24
)

type VolumeState int

const (
	Detached VolumeState = iota
	Faulted
	Healthy
	Degraded
)

type VolumeManager interface {
	Create(volume *VolumeInfo) (*VolumeInfo, error)
	Delete(name string) error
	Get(name string) (*VolumeInfo, error)
	Attach(name string) error
	Detach(name string) error

	MonitorVolume(volume *VolumeInfo)
	Cleanup(volume *VolumeInfo)
}

type Orchestrator interface {
	CreateVolumeRecord(volume *VolumeInfo) (*VolumeInfo, error)
	DeleteVolumeRecord(volumeName string) error
	GetVolumeRecord(volumeName string) (*VolumeInfo, error)
	MarkBadReplica(containerID string) error

	CreateController(volumeName, hostID string) (*ContainerInfo, error)
	CreateReplica(volume *VolumeInfo) (*ContainerInfo, error)

	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RemoveContainer(containerID string) error
}

type VolumeInfo struct {
	Name                string
	Size                int64
	NumberOfReplicas    int64
	StaleReplicaTimeout time.Duration
	Controller          *ContainerInfo
	Replicas            []*ContainerInfo
	State               *VolumeState
}

type ContainerInfo struct {
	ID           string
	HostID       string
	Running      *bool
	BadTimestamp *time.Time
}
