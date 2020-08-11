package cmd

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/GwonsooLee/kubenx/pkg/runner"
	"github.com/GwonsooLee/kubenx/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh to specific node",
	Long:  `ssh to specific node`,
	Run: func(cmd *cobra.Command, args []string) {
		argsLen := len(args)

		if argsLen > 2 {
			utils.Red("Too many Arguments")
			os.Exit(1)
		}

		if argsLen == 0 {
			utils.Red("Usage: kubenx ssh [key]")
			os.Exit(1)
		}

		key := args[0]
		tryRemoteSSH(key)
	},
}

// SSH Client Configuration
type SSHClient struct {
	client *ssh.Client
}

// SSH Configuration
type SSHConfig struct {
	addr    string
	port    string
	user    string
	keyfile string
}

// an array of bastion server
type Bastion struct {
	Servers []Server `json:"bastion"`
}

//Server Configuration
type Server struct {
	Key     string `json:"key"`
	Addr    string `json:"addr"`
	Port    string `json:"port"`
	User    string `json:"user"`
	KeyFile string `json:"keyfile"`
}

// Terminal Configuration
type TerminalConfig struct {
	Term   string
	Height int
	Weight int
	Modes  ssh.TerminalModes
}

type remoteShell struct {
	client         *ssh.Client
	requestPty     bool
	terminalConfig *TerminalConfig

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// Terminal create a interactive shell on client.
func (c *SSHClient) Terminal(config *TerminalConfig) *remoteShell {
	return &remoteShell{
		client:         c.client,
		terminalConfig: config,
		requestPty:     true,
	}
}

// Start start a remote shell on client
func (rs *remoteShell) Start(hostname string) error {
	utils.Blue("Getting new session...")
	session, err := rs.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if rs.stdin == nil {
		session.Stdin = os.Stdin
	} else {
		session.Stdin = rs.stdin
	}
	if rs.stdout == nil {
		session.Stdout = os.Stdout
	} else {
		session.Stdout = rs.stdout
	}
	if rs.stderr == nil {
		session.Stderr = os.Stderr
	} else {
		session.Stderr = rs.stderr
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	fileDescriptor := int(os.Stdin.Fd())

	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			utils.Red(err)
			os.Exit(1)
		}
		defer terminal.Restore(fileDescriptor, originalState)
		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)

		if rs.requestPty {
			tc := rs.terminalConfig
			if tc == nil {
				tc = &TerminalConfig{
					Term:   "xterm-256color",
					Height: 40,
					Weight: 80,
				}
			}
			if err := session.RequestPty(tc.Term, termHeight, termWidth, tc.Modes); err != nil {
				return err
			}
		}
	}

	if err := session.Shell(); err != nil {
		return err
	}

	if err := session.Wait(); err != nil {
		return err
	}

	utils.Yellow("Connection to " + hostname + " closed.")

	return nil
}

// Try to access remote SSH
func tryRemoteSSH(key string) {
	//Get SSH Configuration
	sshConfig := getSshConfiguration(key)

	targetHost, targetPort := getTargetInstanceConfiguration()

	client, err := _dial_with_keypair(sshConfig, targetHost, targetPort)
	if err != nil {
		utils.Red(err)
		os.Exit(1)
	}

	// with a terminal config
	config := &TerminalConfig{
		Term:   "xterm-256color",
		Height: 40,
		Weight: 80,
		Modes: ssh.TerminalModes{
			ssh.ECHO:          1,     // echo disabled
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		},
	}
	if err := client.Terminal(config).Start(sshConfig.addr); err != nil {
		utils.Red(err)
		os.Exit(1)
	}
}

// Choose instance that you want to access
func getTargetInstanceConfiguration() (string, string) {
	port := utils.TARGET_DEFAULT_PORT
	nodeList := runner.GetNodeListForOption(nil)

	var server string
	if len(nodeList) > 0 {
		prompt := &survey.Select{
			Message: "Choose a node:",
			Options: nodeList,
		}
		survey.AskOne(prompt, &server)
	}

	return server, port
}

// Get SSH from configuration
func getSshConfiguration(key string) *SSHConfig {
	// Open our jsonFile
	sshFile, err := os.Open(filepath.Join(utils.HomeDir(), utils.KUBENX_HOMEDIR, utils.SSH_DEFAULT_PATH))
	if err != nil {
		utils.Red(utils.NO_FILE_EXCEPTION)
		os.Exit(1)
	}
	defer sshFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(sshFile)

	// we initialize our Users array
	var bastion Bastion

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &bastion)

	// Get Server Information of server[key]
	var addr, port, user, keyfile string
	for _, server := range bastion.Servers {
		if key == server.Key {
			addr = server.Addr
			port = server.Port
			user = server.User
			keyfile = server.KeyFile

			break
		}
	}

	// Make sshConfig
	sshConfig := &SSHConfig{
		addr:    addr,
		port:    port,
		user:    user,
		keyfile: filepath.Join(utils.HomeDir(), ".ssh", keyfile),
	}

	return sshConfig
}

// _dial_with_keypair starts a client connection to the given SSH server with key authmethod.
func _dial_with_keypair(sshConfig *SSHConfig, targetHost, targetPort string) (*SSHClient, error) {
	key, err := ioutil.ReadFile(sshConfig.keyfile)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: "ec2-user",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	return _dial("tcp", sshConfig.addr+":"+sshConfig.port, targetHost+":"+targetPort, config)
}

// Dial starts a client connection to the given SSH server.
// This is wrap the ssh.Dial
func _dial(network, addr, targetAddr string, config *ssh.ClientConfig) (*SSHClient, error) {
	utils.Blue("Get Bastion Client Connection to " + addr)
	client, err := ssh.Dial(network, addr, config)
	if err != nil {
		utils.Red(err)
		return nil, err
	}

	//// Remote Connection Through
	utils.Blue("Get Remote Connection to " + targetAddr)
	remoteConn, err := client.Dial("tcp", targetAddr)
	if err != nil {
		utils.Red(err)
		return nil, err
	}

	//key, err := ioutil.ReadFile(filepath.Join(homeDir(), ".ssh", "k8s_rsa"))
	//if err != nil {
	//	return nil, err
	//}
	//
	//signer, err := ssh.ParsePrivateKey(key)
	//if err != nil {
	//	return nil, err
	//}
	//
	//newConfig := &ssh.ClientConfig{
	//	User: "ec2-user",
	//	Auth: []ssh.AuthMethod{
	//		ssh.PublicKeys(signer),
	//	},
	//	HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	//}

	newConn, _, _, err := ssh.NewClientConn(remoteConn, targetAddr, config)
	if err != nil {
		utils.Red(err)
		return nil, err
	}

	return &SSHClient{
		client: &ssh.Client{Conn: newConn},
		//client: client,
	}, nil
}
