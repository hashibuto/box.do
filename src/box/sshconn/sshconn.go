package sshconn

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

type SSHConn struct {
	Conn *ssh.Client
}

// NewSSHConn returns a new SSH connection
func NewSSHConn(privkeyFilename, username, hostname string) (*SSHConn, error) {
	var err error
	keyData, err := ioutil.ReadFile(privkeyFilename)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	signer, err = ssh.ParsePrivateKey(keyData)
	if _, ok := err.(*ssh.PassphraseMissingError); ok == true {
		for {
			var input string
			fmt.Print("Enter your SSH key passphrase\n>")
			reader := bufio.NewReader(os.Stdin)
			input, err = reader.ReadString('\n')
			if err != nil {
				return nil, err
			}

			// Trim any leading or trailing whitespace, including the delimiter
			input = strings.Trim(input, " \n\t")
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(input))
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Yes, this is a bad practice
	}

	conn, err := ssh.Dial("tcp", hostname, config)
	if err != nil {
		return nil, err
	}

	return &SSHConn{
		Conn: conn,
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

	// stdoutPipe, err := session.StdoutPipe()
	// if err != nil {
	// 	return err
	// }
	// stderrPipe, err := session.StderrPipe()
	// if err != nil {
	// 	return err
	// }

	// // Perform background copy of SSH session output to the local output streams
	// go io.Copy(os.Stdout, stdoutPipe)
	// go io.Copy(os.Stderr, stderrPipe)

	err = session.Run(strings.Join(commands, " && "))
	return err
}

// Close closes the underlying SSH connection
func (conn *SSHConn) Close() {
	conn.Conn.Close()
}
