package cattle

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
	client "github.com/rancher/go-rancher/v2"
	"github.com/rancher/longhorn-orc/types"
	"github.com/urfave/cli"
)

type cattleOrc struct {
	rancher  *client.RancherClient
	metadata metadata.Client

	hostID string
}

func New(c *cli.Context) types.Orchestrator {
	rancher, err := client.NewRancherClient(&client.ClientOpts{
		Url:       c.GlobalString("cattle-url"),
		AccessKey: c.GlobalString("cattle-access-key"),
		SecretKey: c.GlobalString("cattle-secret-key"),
	})
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "failed to get rancher client"))
	}
	md := metadata.NewClient(c.GlobalString("metadata-url"))
	host, err := md.GetSelfHost()
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "failed to get self host from rancher metadata"))
	}
	return &cattleOrc{
		rancher:  rancher,
		metadata: md,
		hostID:   host.UUID,
	}
}

func (orc *cattleOrc) CreateVolume(volume *types.VolumeInfo) (*types.VolumeInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) DeleteVolume(volumeName string) error {
	return nil
}

func (orc *cattleOrc) GetVolume(volumeName string) (*types.VolumeInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) MarkBadReplica(replica *types.ReplicaInfo) error {
	return nil
}

func (orc *cattleOrc) CreateController(volumeName string, replicas []*types.ReplicaInfo) (*types.ControllerInfo, error) {
	// create the controller on this host
	return nil, nil
}

func (orc *cattleOrc) CreateReplica(volumeName string) (*types.ReplicaInfo, error) {
	return nil, nil
}

func (orc *cattleOrc) StartReplica(containerID string) error {
	return nil
}

func (orc *cattleOrc) StopReplica(containerID string) error {
	return nil
}

func (orc *cattleOrc) RemoveContainer(containerID string) error {
	return nil
}

func (orc *cattleOrc) GetThisHostID() string {
	return orc.hostID
}
