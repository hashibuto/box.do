package sshconn

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHConn struct {
	Conn      *ssh.Client
	SSHSigner *SSHSigner
}

type SSHSigner struct {
	Conn   net.Conn
	Signer ssh.Signer
}

// GetSigner returns an SSH signer object for use with SSH connections
func GetSigner(privkeyFilename string) (*SSHSigner, error) {
	var signer ssh.Signer
	var err error
	keyData, err := ioutil.ReadFile(privkeyFilename)
	if err != nil {
		return nil, err
	}

	signer, err = ssh.ParsePrivateKey(keyData)
	if _, ok := err.(*ssh.PassphraseMissingError); ok == true {
		pubkeyFilename := fmt.Sprintf("%v.pub", privkeyFilename)
		pubkeyData, err := ioutil.ReadFile(pubkeyFilename)
		if err != nil {
			fmt.Println("Unable to locate corresponding public key file", pubkeyFilename)
			return nil, err
		}
		pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubkeyData)
		myPubKeyBlob := pubKey.Marshal()
		if err != nil {
			return nil, err
		}

		socket := os.Getenv("SSH_AUTH_SOCK")
		if socket == "" {
			return nil, errors.New("No SSH agent available to process encrypted private key")
		}
		conn, err := net.Dial("unix", socket)
		if err != nil {
			return nil, err
		}
		agentClient := agent.NewClient(conn)
		signers, err := agentClient.Signers()
		if err != nil {
			return nil, err
		}
		for _, signer = range signers {
			signerPubKey := signer.PublicKey()
			blob := signerPubKey.Marshal()
			if bytes.Compare(blob, myPubKeyBlob) == 0 {
				return &SSHSigner{
					Signer: signer,
					Conn:   conn,
				}, nil
			}
		}

		return nil, errors.New("This key is not being managed by the SSH agent")
	}
	if err != nil {
		return nil, err
	}

	return &SSHSigner{
		Signer: signer,
		Conn:   nil,
	}, nil
}

// NewSSHConn returns a new SSH connection
func NewSSHConn(signer *SSHSigner, username, hostname string) (*SSHConn, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer.Signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Yes, this is a bad practice
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%v:22", hostname), config)
	if err != nil {
		return nil, err
	}

	return &SSHConn{
		Conn:      conn,
		SSHSigner: signer,
	}, nil
}

// Run runs commands in succession on the remote server.  If any command fails, the subsequent
// in the list will not be executed.
func (conn *SSHConn) Run(commands []string) error {
	session, err := conn.Conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run(strings.Join(commands, " && "))
	return err
}

// Close closes the underlying SSH connection
func (conn *SSHConn) Close() {
	conn.Conn.Close()
	if conn.SSHSigner.Conn != nil {
		conn.SSHSigner.Conn.Close()
	}
}
