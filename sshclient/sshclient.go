package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var (
	privateKeyPath   string
	knownHostsPath   string
	hostKeyAlgorithm string
	user             string
	network          string
	address          string
	command          string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get user home directory: %v\n", err)
		os.Exit(1)
	}

	flag.StringVar(&privateKeyPath, "key", homeDir+"/.ssh/id_ed25519", "Path to the private key file")
	flag.StringVar(&knownHostsPath, "known-hosts", homeDir+"/.ssh/known_hosts", "Path to the known hosts file")
	flag.StringVar(&hostKeyAlgorithm, "host-key-algorithm", ssh.KeyAlgoED25519, "Host key algorithm to use")
	flag.StringVar(&user, "user", "admin", "Username for SSH authentication")
	flag.StringVar(&network, "network", "tcp", "Network type (e.g., tcp)")
	flag.StringVar(&address, "address", "", "Address of the SSH server (e.g., host:port)")
	flag.StringVar(&command, "command", "uname -a\nnvram get wan0_ipaddr\nnvram get wan1_ipaddr\n", "Command to run on the SSH server")
}

func main() {
	flag.Parse()

	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read private key: %v\n", err)
		os.Exit(1)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse private key: %v\n", err)
		os.Exit(1)
	}

	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create host key callback: %v\n", err)
		os.Exit(1)
	}

	clientConfig := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoED25519,
		},
	}
	client, err := ssh.Dial(network, address, &clientConfig)
	if err != nil {
		keyErr, ok := errors.AsType[*knownhosts.KeyError](err)
		if ok {
			fmt.Fprintf(os.Stderr, "Host key mismatch: %v\n", keyErr.Want)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to dial SSH: %v\n", err)
		}
		os.Exit(1)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create SSH session: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	session.Stdin = strings.NewReader(command)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start shell: %v\n", err)
		os.Exit(1)
	}

	if err := session.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Shell exited with error: %v\n", err)
		os.Exit(1)
	}
}
