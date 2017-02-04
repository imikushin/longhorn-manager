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

type ReplicaMode int

const (
	RW ReplicaMode = iota
	WO
	ERR
)

type VolumeManager interface {
	Create(volume *VolumeInfo) (*VolumeInfo, error)
	Delete(name string) error
	Get(name string) (*VolumeInfo, error)
	Attach(name string) error
	Detach(name string) error

	CheckController(ctrl Controller, volume *VolumeInfo) error
	Cleanup(volume *VolumeInfo) error
}

type Monitor func(volume *VolumeInfo, man VolumeManager) io.Closer

type WaitForDevice func(name string) error

type GetController func(volume *VolumeInfo) Controller

type Controller interface {
	Name() string
	GetReplicaStates() ([]*ReplicaInfo, error)
	AddReplica(replica *ReplicaInfo) error
	RemoveReplica(replica *ReplicaInfo) error
}

type Orchestrator interface {
	CreateVolume(volume *VolumeInfo) (*VolumeInfo, error) // creates (but does not start) replicas as well
	DeleteVolume(volumeName string) error                 // removes all the volume components
	GetVolume(volumeName string) (*VolumeInfo, error)
	MarkBadReplica(replica *ReplicaInfo) error // find replica by Address

	CreateController(volumeName, hostID string, replicas []*ReplicaInfo) (*ControllerInfo, error)
	CreateReplica(volumeName string) (*ReplicaInfo, error)

	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RemoveContainer(containerID string) error

	GetThisHostID() string
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
	Address string
	Running bool
}

type ControllerInfo struct {
	ContainerInfo

	Name string
}

type ReplicaInfo struct {
	ContainerInfo

	Mode         ReplicaMode
	BadTimestamp *time.Time
}
