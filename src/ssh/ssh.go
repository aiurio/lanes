package sshUtil

// adapted from https://gist.github.com/svett/5d695dcc4cc6ad5dd275
import (
	"os"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io/ioutil"
	"log"
	"regexp"
	"errors"
	"strconv"
)

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHtunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint

	Config *ssh.ClientConfig

	localConn net.Conn
	remoteConn net.Conn
	serverConn *ssh.Client
}


func Tunnel(host string, port int, privateKey string, lFlag string) error {
	fmt.Printf("%v %v %v %v", host, port, privateKey, lFlag)

	var lPort , rPort int
	var parseErr error

	exp := regexp.MustCompile(`(?P<lport>\d*):(?P<rhost>.*):(?P<rport>\d*)`)
	if( !exp.Match( []byte(lFlag) )){
		return errors.New("Unable to parse tunnel string: " + lFlag)
	}

	m := exp.FindStringSubmatch( lFlag )

	// parse ports
	if lPort, parseErr = strconv.Atoi(m[1]); parseErr != nil {
		return errors.New("Failed to parse Local Port from " + m[1])
	}
	if rPort, parseErr = strconv.Atoi(m[3]); parseErr != nil {
		return errors.New("Failed to parse Remote Port from " + m[3])
	}

	// setup endpoints
	localEndpoint := & Endpoint{
		Host: "localhost",
		Port: lPort,
	}
	serverEndpoint := & Endpoint{
		Host: host,
		Port: port,
	}
	remoteEndpoint := & Endpoint{
		Host: m[2],
		Port: rPort,
	}

	// read private key & create signer
	key, err := ioutil.ReadFile(privateKey)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	// configure ClientConfig with user and Private Key
	sshConfig := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
	}

	tunnel := & SSHtunnel{
		Config: sshConfig,
		Local:  localEndpoint,
		Server: serverEndpoint,
		Remote: remoteEndpoint,
	}

	log.Printf("Preparing tunnel at %v \n", tunnel)

	tunnel.Start()
	log.Print("Tunneled! \n")
	
	return nil
}



func (tunnel *SSHtunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go tunnel.forward(conn)
	}
}

func (tunnel *SSHtunnel) forward(localConn net.Conn) {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		fmt.Printf("Server dial error: %s\n", err)
		return
	}

	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		fmt.Printf("Remote dial error: %s\n", err)
		return
	}

	copyConn:=func(writer, reader net.Conn) {
		_, err:= io.Copy(writer, reader)
		if err != nil {
			fmt.Printf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)

	tunnel.localConn = localConn
	tunnel.remoteConn = remoteConn
	tunnel.serverConn = serverConn
}

func (tunnel *SSHtunnel) Shutdown() {
	if tunnel.localConn != nil {
		tunnel.localConn.Close()
		tunnel.localConn = nil
	}
	if tunnel.remoteConn != nil {
		tunnel.remoteConn.Close()
		tunnel.remoteConn = nil
	}
	if tunnel.serverConn != nil {
		tunnel.serverConn.Close()
		tunnel.serverConn = nil
	}
}

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}
