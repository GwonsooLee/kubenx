/*
Copyright © 2020 NAME HERE <gwonsoo.lee@gmail>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net"
	"os"
	"encoding/json"
	"path/filepath"
	"github.com/AlecAivazis/survey/v2"
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh to specific node",
	Long: `ssh to specific node`,
	Run: func(cmd *cobra.Command, args []string) {
		argsLen := len(args)

		if argsLen > 2 {
			Red("Too many Arguments")
			os.Exit(1)
		}

		if argsLen == 0 {
			Red("Usage: kubenx ssh [key]")
			os.Exit(1)
		}

		key := args[0]
		_try_remote_ssh(key)
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)
}

// SSH Client Configuration
type SSHClient struct {
	client *ssh.Client
}

// SSH Configuration
type SSHConfig struct {
	addr string
	port string
	user string
	keyfile string
}

// an array of bastion server
type Bastion struct {
	Servers []Server `json:"bastion"`
}

//Server Configuration
type Server struct {
	Key         string `json:"key"`
	Addr   		string `json:"addr"`
	Port   		string `json:"port"`
	User   		string `json:"user"`
	KeyFile   	string `json:"keyfile"`
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
	Blue("Getting new session...")
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
			Red(err)
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

	Yellow("Connection to " + hostname + " closed.")

	return nil
}

// Try to access remote SSH
func _try_remote_ssh(key string)  {
	//Get SSH Configuration
	sshConfig := _get_ssh_configuration(key)

	targetHost, targetPort := _get_target_instance_configuration()

	client, err := _dial_with_keypair(sshConfig, targetHost, targetPort)
	if err != nil {
		Red(err)
		os.Exit(1)
	}

	// with a terminal config
	config := &TerminalConfig {
		Term: "xterm-256color",
		Height: 40,
		Weight: 80,
		Modes: ssh.TerminalModes {
			ssh.ECHO: 1, // echo disabled
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		},
	}
	if err := client.Terminal(config).Start(sshConfig.addr); err != nil {
		Red(err)
		os.Exit(1)
	}
}

// Choose instance that you want to access
func _get_target_instance_configuration() (string, string) {
	port := TARGET_DEFAULT_PORT

	//svc := _get_eks_session()
	//
	//// Check the cluster First
	//cluster := _choose_cluster()
	//
	////Choose Nodegroup
	//nodegroup := _choose_nodegroup(cluster)
	//
	//// Get NodeGroup Information with svc
	//info := _get_nodegroup_info_with_session(svc, cluster, nodegroup)
	//AutoScalingGroupsArray := info.Nodegroup.Resources.AutoScalingGroups
	//for _, group := range AutoScalingGroupsArray {
	//	autoscalingInfo := _get_autoscaling_group_info(nil, *group.Name).AutoScalingGroups[0]
	//	for _, instance := range autoscalingInfo.Instances {
	//		instanceList = append(instanceList, instance.InstanceId)
	//	}
	//}
	//
	//// Get Instance Information
	//instanceInfo := _get_ec2_instance_info(instanceList)
	//serverList := instanceInfo.Reservations[0].Instances
	//
	//
	//server := ""
	//for _, network := range serverList[0].NetworkInterfaces[0].PrivateIpAddresses {
	//	if *network.Primary == true {
	//
	//		Blue(network)
	//		server = *network.PrivateDnsName
	//		break
	//	}
	//}
	nodeList := _get_node_list_for_option(nil)

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
func _get_ssh_configuration(key string) *SSHConfig {
	// Open our jsonFile
	sshFile, err := os.Open(filepath.Join(homeDir(), KUBENX_HOMEDIR, SSH_DEFAULT_PATH))
	if err != nil {
		Red(NO_FILE_EXCEPTION)
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
		addr: addr,
		port: port,
		user: user,
		keyfile: filepath.Join(homeDir(), ".ssh", keyfile),
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
	Blue("Get Bastion Client Connection to " + addr)
	client, err := ssh.Dial(network, addr, config)
	if err != nil {
		Red(err)
		return nil, err
	}

	//// Remote Connection Through
	Blue("Get Remote Connection to " + targetAddr)
	remoteConn, err := client.Dial("tcp", targetAddr)
	if err != nil {
		Red(err)
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
		Red(err)
		return nil, err
	}

	return &SSHClient{
		client: &ssh.Client{Conn: newConn},
		//client: client,
	}, nil
}
