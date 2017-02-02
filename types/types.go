package types

import (
	"io"
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

	CheckVolume(volume *VolumeInfo) error
	Cleanup(volume *VolumeInfo) error
}

type MonitorVolume func(name string, man VolumeManager) io.Closer

type WaitForDevice func(name string) error

type Orchestrator interface {
	CreateVolumeRecord(volume *VolumeInfo) (*VolumeInfo, error)
	DeleteVolumeRecord(volumeName string) error
	GetVolumeRecord(volumeName string) (*VolumeInfo, error)
	MarkBadReplica(containerID string) error

	CreateController(volumeName, hostID string, replicas []*ContainerInfo) (*ContainerInfo, error)
	CreateReplica(volume *VolumeInfo) (*ContainerInfo, error)

	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RemoveContainer(containerID string) error

	GetThisHostID() (string, error)
}

type VolumeInfo struct {
	Name                string
	Size                int64
	NumberOfReplicas    int
	StaleReplicaTimeout time.Duration
	Controller          *ContainerInfo
	Replicas            []*ContainerInfo
	State               *VolumeState
}

type ContainerInfo struct {
	ID           string
	HostID       string
	Running      bool
	BadTimestamp *time.Time
}
