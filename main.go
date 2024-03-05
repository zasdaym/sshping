package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	defaultIdentityFile := path.Join(homeDir, ".ssh", "id_ed25519")

	var (
		addr         = flag.String("addr", "localhost", "SSH server address")
		count        = flag.Int("count", 1000, "Number of characters to echo")
		identityFile = flag.String("identity_file", defaultIdentityFile, "SSH public key")
		timeout      = flag.Duration("timeout", 10*time.Second, "Timeout duration per character")
		user         = flag.String("user", "root", "SSH user")
		password     = flag.String("password", "", "SSH password")
	)
	flag.Parse()

	identityFileBytes, err := os.ReadFile(*identityFile)
	if err != nil {
		return err
	}

	privateKey, err := ssh.ParsePrivateKey(identityFileBytes)
	if err != nil {
		return err
	}

	client, err := ssh.Dial("tcp", *addr, &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            *user,
	})
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	var (
		maxLatency time.Duration
		minLatency time.Duration = *timeout
		sumLatency time.Duration
	)

	for i := 0; i < *count; i++ {
		start := time.Now()
		session.Run("echo -n a")
		elapsed := time.Since(start)
		minLatency = min(elapsed, minLatency)
		maxLatency = max(elapsed, maxLatency)
		sumLatency += elapsed
	}

	fmt.Printf("Avg latency: %s\nMax latency: %s\nMin latency: %s\n", sumLatency/time.Duration(*count), maxLatency, minLatency)
	return nil
}
