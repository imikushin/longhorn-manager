package cattle

import (
	"bytes"
	"fmt"
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
	dockerComposeTemplate  *template.Template
	rancherComposeTemplate *template.Template
)

func init() {
	t, err := template.New("docker-compose").Parse(dockerComposeText)
	if err != nil {
		logrus.Fatalf("Error parsing volume stack template: %v", err)
	}
	dockerComposeTemplate = t

	t, err = template.New("docker-compose").Parse(rancherComposeText)
	if err != nil {
		logrus.Fatalf("Error parsing volume stack template: %v", err)
	}
	rancherComposeTemplate = t
}

type cattleOrc struct {
	rancher  *client.RancherClient
	metadata metadata.Client

	hostUUID, containerUUID string

	LonghornImage, OrcImage string

	Env map[string]interface{}
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
	return initOrc(&cattleOrc{
		rancher:       rancherClient,
		metadata:      md,
		hostUUID:      host.UUID,
		containerUUID: container.UUID,
		LonghornImage: c.GlobalString(orch.LonghornImageParam),
		OrcImage:      c.GlobalString(orch.OrcImageParam),
	})
}

func initOrc(orc *cattleOrc) *cattleOrc {
	orc.Env = map[string]interface{}{
		"LONGHORN_IMAGE": orc.LonghornImage,
		"ORC_IMAGE":      orc.OrcImage,
		"ORC_CONTAINER":  orc.containerUUID,
	}
	return orc
}

func volumeStackName(name string) string {
	return "volume-" + name
}

func replicaName(i int) string {
	return "replica" + strconv.Itoa(i)
}

func stackBytes(t *template.Template, volume *types.VolumeInfo) []byte {
	buffer := new(bytes.Buffer)
	if err := t.Execute(buffer, volume); err != nil {
		logrus.Fatalf("Error applying the stack golang template: %v", err)
	}
	logrus.Debugf("%s", buffer)
	return buffer.Bytes()
}

func (orc *cattleOrc) envLookup() config.EnvironmentLookup {
	return &rLookup.MapEnvLookup{Env: orc.Env}
}

func (orc *cattleOrc) composeProject(volume *types.VolumeInfo, stack *client.Stack) project.APIProject {
	ctx := &rancher.Context{
		Context: project.Context{
			EnvironmentLookup: orc.envLookup(),
			LoggerFactory:     logger.NewColorLoggerFactory(),
			ComposeBytes:      [][]byte{stackBytes(dockerComposeTemplate, volume)},
		},
		RancherComposeBytes: stackBytes(rancherComposeTemplate, volume),
		Client:              orc.rancher,
		Stack:               stack,
	}
	p, err := rancher.NewProject(ctx)
	if err != nil {
		logrus.Fatalf("%+v", errors.Wrap(err, "error creating compose project"))
	}
	p.Name = volumeStackName(volume.Name)
	return p
}

func copyVolumeProperties(volume0 *types.VolumeInfo) *types.VolumeInfo {
	volume := new(types.VolumeInfo)
	*volume = *volume0
	volume.Controller = nil
	volume.Replicas = nil
	volume.State = nil
	return volume
}

func (orc *cattleOrc) CreateVolume(volume *types.VolumeInfo) (*types.VolumeInfo, error) {
	volume = copyVolumeProperties(volume)
	stack0 := &client.Stack{
		Name:        volumeStackName(volume.Name),
		ExternalId:  fmt.Sprintf("system://%s?name=%s", "rancher-longhorn", volume.Name),
		Environment: orc.Env,
		Outputs: map[string]interface{}{ // TODO add and use Metadata
			LastReplicaIndexProp: strconv.Itoa(volume.NumberOfReplicas),
		},
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
	if err := p.Create(context.Background(), options.Create{}, replicaNames...); err != nil {
		return nil, errors.Wrap(err, "failed to create replica services")
	}

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
