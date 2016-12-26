package main


import (
	"os"
	"github.com/urfave/cli"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	ec2util "./ec2"
	"ssh"
	"log"
)


func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		addCmd(),
		nestedCmd(),
		listInstances(),
	}


	app.Run(os.Args)

}

func nestedCmd() cli.Command{
	return cli.Command {
		Name:        "template",
		Aliases:     []string{"t"},
		Usage:       "options for task templates",
		Subcommands: []cli.Command{
			{
				Name:  "add",
				Usage: "add a new template",
				Action: func(c *cli.Context) error {
					fmt.Println("new task template: ", c.Args().First())
					return nil
				},
			},
			{
				Name:  "remove",
				Usage: "remove an existing template",
				Action: func(c *cli.Context) error {
					fmt.Println("removed task template: ", c.Args().First())
					return nil
				},
			},
		},
	}
}
func addCmd() cli.Command{
	return cli.Command {
		Name:    "add",
		Aliases: []string{"a"},
		Usage:   "add a task to the list",
		Action:  func(c *cli.Context) error {
			fmt.Println("added task: ", c.Args().First())
			return nil
		},
	};
}

func listInstances() cli.Command{
	return cli.Command {
		Name:    "list",
		Aliases: []string{"l"},
		Usage:   "List EC2 instances",
		Action:  func(c *cli.Context) error {
			sess, err := session.NewSession()
			if err != nil {
				return nil
			}

			var svc ec2util.Ec2ServiceImpl
			svc = ec2util.Ec2ServiceImpl{
				Session: sess,
				Region: "us-west-2",
			}
			instances := svc.InstancesByRegionTagAndValue("us-west-2", "Lane", "inf")

			tErr := sshUtil.Tunnel(*instances["jenkins"].PublicIpAddress, 22, "/Users/dave/.ssh/id_rsa", "9000:davidwelch.co:80")
			if tErr != nil {
				log.Fatal("Failed ", tErr)
			}

			return nil;
		},
	};
}
