package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/concourse"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"

	"gopkg.in/urfave/cli.v1"
)

var destroyArgs config.DestroyArgs

var destroyFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &destroyArgs.AWSRegion,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(optional) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Value:       "AWS",
		Hidden:      true,
		Destination: &destroyArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &destroyArgs.Namespace,
	},
}

func destroyAction(c *cli.Context, destroyArgs config.DestroyArgs, iaasFactory iaas.Factory) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `concourse-up destroy <name>`")
	}

	version := c.App.Version

	destroyArgs, err := setRegion(c, destroyArgs)
	if err != nil {
		return err
	}

	client, err := buildDestroyClient(name, version, destroyArgs, iaasFactory)
	if err != nil {
		return err
	}

	if !NonInteractiveModeEnabled() {
		confirm, err := util.CheckConfirmation(os.Stdin, os.Stdout, name)
		if err != nil {
			return err
		}

		if !confirm {
			fmt.Println("Bailing out...")
			return nil
		}
	}

	return client.Destroy()
}

func setRegion(c *cli.Context, destroyArgs config.DestroyArgs) (config.DestroyArgs, error) {
	if !c.IsSet("region") {
		if destroyArgs.IAAS == "AWS" {
			destroyArgs.AWSRegion = "eu-west-1"
		} else if destroyArgs.IAAS == "GCP" {
			destroyArgs.AWSRegion = "europe-west1"
		}
	}

	return destroyArgs, nil
}
func buildDestroyClient(name, version string, destroyArgs config.DestroyArgs, iaasFactory iaas.Factory) (*concourse.Client, error) {
	awsClient, err := iaasFactory(destroyArgs.IAAS, destroyArgs.AWSRegion)
	if err != nil {
		return nil, err
	}
	terraformClient, err := terraform.New(terraform.DownloadTerraform())
	if err != nil {
		return nil, err
	}
	client := concourse.NewClient(
		awsClient,
		terraformClient,
		bosh.New,
		fly.New,
		certs.Generate,
		config.New(awsClient, name, destroyArgs.Namespace),
		nil,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		version,
	)

	return client, nil
}

var destroy = cli.Command{
	Name:      "destroy",
	Aliases:   []string{"d"},
	Usage:     "Destroys a Concourse",
	ArgsUsage: "<name>",
	Flags:     destroyFlags,
	Action: func(c *cli.Context) error {
		return destroyAction(c, destroyArgs, iaas.New)
	},
}
