package main


import (
	ec2util "./ec2"
	"./ssh"

	"log"
	"fmt"
	"os"
	"os/user"
	"github.com/urfave/cli"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
	"strings"
	"io/ioutil"
	"errors"
)


func readYaml(path string)(map[interface{}]interface{}, error ) {
	bytes, err := ioutil.ReadFile(path)
	if( err != nil ){
		return nil, err
	}

	val := make(map[interface{}]interface{} )
	err = yaml.Unmarshal(bytes, &val)
	if err != nil {
		return nil, err
	}
	fmt.Printf("--- t:\n%v\n\n", val)

	return val, nil
}

// recursively iterates down a "dot-path" of a map – returning the matching element
func maps(m map[interface{}]interface{}, key string) (interface{}, error){
	var ok bool

	keys := strings.Split(key, ".")

	log.Printf("Looking at %v", keys)

	for i, key := range keys {
		log.Printf("Looking at %v", key)

		val := m[key]

		if i == len(keys) - 1 {
			return val, nil
		} else if m, ok = val.(map[interface{}]interface{}); !ok {
			log.Printf("Non-Map at %v of type %v", key, val)
			return nil, errors.New("Received a non-map value along the path at " + key)
		}else {
			m = val.(map[interface{}]interface{})
		}
	}

	return nil, errors.New("Should have never gotten here!")
}


func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		addCmd(),
		nestedCmd(),
		listInstances(),
	}

	user, err := user.Current()
	if err != nil {
		log.Fatal( "Failed to determine current user: ", err )
	}

	base := user.HomeDir + "/.lanes/"
	config, err := readYaml( base + "lanes.yml")
	if( err != nil ){
		log.Fatal("Failed to load base lanes configuration file: ", err)
	}

	var profileStr string
	var ok bool

	profile := config["profile"]
	if profileStr, ok = profile.(string); !ok {
		log.Fatal("Problem getting profile ")
	}
	if(  len(strings.TrimSpace(profileStr)) < 1 ){
		log.Fatal("Failed to load base lanes configuration's profile: ", err)
	}

	log.Printf("Connecting " + base + profileStr + ".yml")
	config, err = readYaml( base + profileStr + ".yml")
	if( err != nil ){
		log.Fatal("Failed to load base lanes configuration file: ", err)
	}

	spew.Dump(config)

	raw, e := maps(config, "ssh.mods.dev")
	log.Printf("path %v %v \n", e, raw)


	//app.Run(os.Args)
	log.Printf("test %v \n", os.Args)

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
