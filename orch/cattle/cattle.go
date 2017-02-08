package cattle

import (
	"bytes"
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/cli/logger"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
	client "github.com/rancher/go-rancher/v2"
	"github.com/rancher/longhorn-orc/orch"
	"github.com/rancher/longhorn-orc/types"
	rLookup "github.com/rancher/rancher-compose/lookup"
	"github.com/rancher/rancher-compose/rancher"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"strconv"
	"text/template"
)

const (
	LastReplicaIndexProp = "lastReplicaIndex"
)

var (
	stackTemplate *template.Template
)

func init() {
	t, err := template.New("volume-stack").Parse(stackTemplateText)
	if err != nil {
		logrus.Fatalf("Error parsing volume stack template: %v", err)
	}
	stackTemplate = t
}

type cattleOrc struct {
	rancher  *client.RancherClient
	metadata metadata.Client

	hostUUID, containerUUID string

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
	container, err := md.GetSelfContainer()
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "failed to get self container from rancher metadata"))
	}
	return &cattleOrc{
		rancher:       rancherClient,
		metadata:      md,
		hostUUID:      host.UUID,
		containerUUID: container.UUID,
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

func stackBytes(volume *types.VolumeInfo) []byte {
	buffer := new(bytes.Buffer)
	if err := stackTemplate.Execute(buffer, volume); err != nil {
		logrus.Fatalf("Error applying the stack golang template: %v", err)
	}
	return buffer.Bytes()
}

func (orc *cattleOrc) envLookup() config.EnvironmentLookup {
	return &rLookup.MapEnvLookup{Env: map[string]interface{}{
		"LONGHORN_IMAGE": orc.LonghornImage,
		"ORC_IMAGE":      orc.OrcImage,
		"ORC_CONTAINER":  orc.containerUUID,
	}}
}

func (orc *cattleOrc) composeProject(volume *types.VolumeInfo, stack *client.Stack) project.APIProject {
	context := &rancher.Context{
		Context: project.Context{
			EnvironmentLookup: orc.envLookup(),
			LoggerFactory:     logger.NewColorLoggerFactory(),
		},
		RancherComposeBytes: stackBytes(volume),
		Client:              orc.rancher,
		Stack:               stack,
	}
	p, err := rancher.NewProject(context)
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "error creating compose project"))
	}
	if err := p.Parse(); err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "error parsing compose project"))
	}
	p.Name = volumeStackName(volume.Name)
	return p
}

func (orc *cattleOrc) CreateVolume(volume *types.VolumeInfo) (*types.VolumeInfo, error) {
	stack0 := &client.Stack{
		Name:          volumeStackName(volume.Name),
		System:        true,
		StartOnCreate: false,
		Outputs:       map[string]interface{}{LastReplicaIndexProp: strconv.Itoa(volume.NumberOfReplicas)},
	}
	stack, err := orc.rancher.Stack.Create(stack0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create stack '%s'", stack0.Name)
	}

	replicas := map[int]*types.ReplicaInfo{}
	replicaNames := make([]string, volume.NumberOfReplicas)
	for i := 1; i <= volume.NumberOfReplicas; i++ {
		replicas[i] = &types.ReplicaInfo{Name: replicaName(i)}
		replicaNames[i-1] = replicaName(i)
	}
	volume.Replicas = replicas

	p := orc.composeProject(volume, stack)
	p.Create(context.Background(), options.Create{}, replicaNames...)

	// TODO get replica addresses and stuff

	return volume, nil
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
	return orc.hostUUID
}
