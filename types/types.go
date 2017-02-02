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

type ReplicaState int

const (
	RW ReplicaState = iota
	WO
	ERR
)

type VolumeManager interface {
	Create(volume *VolumeInfo) (*VolumeInfo, error)
	Delete(name string) error
	Get(name string) (*VolumeInfo, error)
	Attach(name string) error
	Detach(name string) error

	CheckController(volume *VolumeInfo) error
	Cleanup(volume *VolumeInfo) error
}

type Monitor func(volume *VolumeInfo, man VolumeManager) io.Closer

type WaitForDevice func(name string) error

type GetController func(controllerInfo *ControllerInfo) Controller

type Controller interface {
	Name() string
	GetReplicaStates() ([]*ReplicaInfo, error)
}

type Orchestrator interface {
	CreateVolumeRecord(volume *VolumeInfo) (*VolumeInfo, error)
	DeleteVolumeRecord(volumeName string) error
	GetVolumeRecord(volumeName string) (*VolumeInfo, error)
	MarkBadReplica(containerID string) error

	CreateController(volumeName, hostID string, replicas []*ReplicaInfo) (*ControllerInfo, error)
	CreateReplica(volume *VolumeInfo) (*ReplicaInfo, error)

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
	Controller          *ControllerInfo
	Replicas            []*ReplicaInfo
	State               *VolumeState
}

type ContainerInfo struct {
	ID      string
	HostID  string
	Running bool
}

type ControllerInfo struct {
	ContainerInfo

	Name string
}

type ReplicaInfo struct {
	ContainerInfo

	Address      string
	State        *ReplicaState
	BadTimestamp *time.Time
}
