/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package sftp

import (
	"fmt"
	"github.com/kovansky/midas"
	"github.com/kovansky/midas/walk"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"path/filepath"
)

type Client struct {
	sshConfig midas.SSHDeploymentSettings
	rootDir   string

	closed bool

	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// NewClient creates a new SFTP client.
//
// The Client.Connect method must be called before using the client.
func NewClient(sshConfig midas.SSHDeploymentSettings) *Client {
	path := sshConfig.Path
	if path == "" {
		path = "./"
	}

	return &Client{
		sshConfig: sshConfig,
		rootDir:   path,
	}
}

// Connect establishes a connection to the remote server using SFTP protocol using given configuration.
func (c *Client) Connect() error {
	var hostKey ssh.PublicKey
	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Set authentication data
	if c.sshConfig.User != "" {
		config.User = c.sshConfig.User
	}
	if method, authMethod, err := c.authenticationMethod(); err != nil {
		panic(err)
	} else {
		if authMethod != nil {
			config.Auth = *authMethod

			if *method == "key" {
				config.HostKeyCallback = ssh.FixedHostKey(hostKey)
			}
		}
	}

	// Set SFTP sshClient address
	port := 22
	if c.sshConfig.Port != nil {
		port = *c.sshConfig.Port
	}

	addr := fmt.Sprintf("%s:%d", c.sshConfig.Host, port)

	fmt.Printf("Trying to handshake with %s...\n", addr)

	// Connect and perform a handshake
	connection, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}

	sftpConnection, err := sftp.NewClient(connection)
	if err != nil {
		return err
	}

	c.sshClient = connection
	c.sftpClient = sftpConnection
	return nil
}

// Close closes the connection to the remote server.
func (c *Client) Close() error {
	if !c.closed {
		err := c.sftpClient.Close()
		err2 := c.sshClient.Close()
		if err == nil && err2 == nil {
			c.closed = true
		}

		return err
	}

	return nil
}

// RemoteFiles returns a list of files in the remote directory.
func (c *Client) RemoteFiles() (walk.FileMap, []error) {
	var (
		errors []error
		files  = make(walk.FileMap)
		walker = c.sftpClient.Walk(c.rootDir)
	)

	for walker.Step() {
		if err := walker.Err(); err != nil {
			errors = append(errors, err)
		}

		stat := walker.Stat()
		relPath, _ := filepath.Rel(c.rootDir, walker.Path())

		if relPath != "" && !stat.IsDir() {
			files[relPath] = stat
		}
	}

	return files, errors
}

// RemoveEmptyDirs removes empty directories from the remote server.
func (c *Client) RemoveEmptyDirs() error {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return err
	}

	err = session.Run(fmt.Sprintf("cd %s && find . -type d -empty -delete", c.rootDir))

	return err
}

// authenticationMethod returns the authentication method name and the authentication method slice based on the provided SFTP configuration.
//
// If method field was not configured, it checks if password field was set and uses it; otherwise it uses "none" method.
func (c *Client) authenticationMethod() (*string, *[]ssh.AuthMethod, error) {
	if c.sshConfig.Method == "" {
		if c.sshConfig.Password != "" {
			c.sshConfig.Method = "password"
		} else {
			c.sshConfig.Method = "none"
		}
	}

	switch c.sshConfig.Method {
	case "password":
		if c.sshConfig.Password == "" {
			return nil, nil, fmt.Errorf("password authentication method requires password")
		}
		return &c.sshConfig.Method, &[]ssh.AuthMethod{ssh.Password(c.sshConfig.Password)}, nil
	case "key":
		if c.sshConfig.Key == "" {
			return nil, nil, fmt.Errorf("key authentication method requires key file")
		}

		key, err := ioutil.ReadFile(c.sshConfig.Key)
		if err != nil {
			return nil, nil, err
		}

		var signer ssh.Signer

		if c.sshConfig.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(c.sshConfig.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}

		if err != nil {
			return nil, nil, err
		}

		return &c.sshConfig.Method, &[]ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	default:
		noneName := "none"
		return &noneName, nil, nil
	}
}
