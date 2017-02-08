package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/rancher/longhorn-orc/controller"
	"github.com/rancher/longhorn-orc/driver"
	"github.com/rancher/longhorn-orc/manager"
	"github.com/rancher/longhorn-orc/orch"
	"github.com/rancher/longhorn-orc/orch/cattle"
	"github.com/rancher/longhorn-orc/storagepool"
	"github.com/rancher/longhorn-orc/util/daemon"
)

var VERSION = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Usage = "Rancher Longhorn storage driver/orchestration"
	app.Action = RunManager

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug logging level",
		},
		cli.StringFlag{
			Name:   "cattle-url",
			Usage:  "The URL endpoint to communicate with cattle server",
			EnvVar: "CATTLE_URL",
		},
		cli.StringFlag{
			Name:   "cattle-access-key",
			Usage:  "The access key required to authenticate with cattle server",
			EnvVar: "CATTLE_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "cattle-secret-key",
			Usage:  "The secret key required to authenticate with cattle server",
			EnvVar: "CATTLE_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   orch.LonghornImageParam,
			EnvVar: "LONGHORN_IMAGE",
		},
		cli.StringFlag{
			Name:   orch.OrcImageParam,
			EnvVar: "ORC_IMAGE",
		},
		cli.IntFlag{
			Name:  "healthcheck-interval",
			Value: 5000,
			Usage: "set the frequency of performing healthchecks",
		},
		cli.StringFlag{
			Name:  "metadata-url",
			Usage: "set the metadata url",
			Value: driver.RancherMetadataURL,
		},
	}

	app.Commands = []cli.Command{storagepool.Command, driver.Command}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatalf("Error running longhorn driver: %v", err)
	}

}

func RunManager(c *cli.Context) error {
	man := manager.New(cattle.New(c), manager.WaitForDevice, manager.Monitor(controller.New))

	go manager.ServeAPI(man)

	return daemon.WaitForExit()
}
