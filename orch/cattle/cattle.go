package cattle

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
	client "github.com/rancher/go-rancher/v2"
	"github.com/rancher/longhorn-orc/orch"
	"github.com/rancher/longhorn-orc/types"
	"github.com/rancher/rancher-compose/rancher"
	"github.com/urfave/cli"
	"strconv"
)

const (
	LastReplicaIndexField = "lastReplicaIndex"
)

type cattleOrc struct {
	rancher  *client.RancherClient
	metadata metadata.Client

	hostID string

	LonghornImage, OrcImage string
}

func New(c *cli.Context) types.Orchestrator {
	rancherClient, err := client.NewRancherClient(&client.ClientOpts{
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
		rancher:       rancherClient,
		metadata:      md,
		hostID:        host.UUID,
		LonghornImage: c.GlobalString(orch.LonghornImageParam),
		OrcImage:      c.GlobalString(orch.OrcImageParam),
	}
}

func volumeStackName(name string) string {
	return "volume-" + name
}

func replicaName(i int) string {
	return "replica-" + strconv.Itoa(i)
}

func (orc *cattleOrc) CreateVolume(volume *types.VolumeInfo) (*types.VolumeInfo, error) {
	stack0 := &client.Stack{
		Name:          volumeStackName(volume.Name),
		System:        true,
		StartOnCreate: false,
		Outputs:       map[string]interface{}{LastReplicaIndexField: "0"},
	}
	_, err := orc.rancher.Stack.Create(stack0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create stack '%s'", stack0.Name)
	}

	replicas := map[int]*types.ReplicaInfo{}
	for i := 1; i <= volume.NumberOfReplicas; i++ {
		replicas[i] = &types.ReplicaInfo{Name: replicaName(i)}
	}
	//lastReplicaIndex := volume.NumberOfReplicas

	// TODO create replicas

	context := &rancher.Context{}
	project := rancher.NewProject(context)
	project.Parse()

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
