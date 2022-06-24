/*
 * Copyright (c) 2022.
 *
 * Originally created by F4 Developer (Stanisław Kowański). Released under GNU GPLv3 (see LICENSE)
 */

package sftp

import (
	"fmt"
	"github.com/kovansky/midas"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

type Client struct {
	sshConfig  midas.SSHDeploymentSettings
	connection *ssh.Client
}

// Connect establishes a connection to the remote server using SSH protocol using given configuration.
func (c *Client) Connect() error {
	var hostKey ssh.PublicKey
	config := &ssh.ClientConfig{}

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

	// Set SSH connection address
	port := 22
	if c.sshConfig.Port != nil {
		port = *c.sshConfig.Port
	}

	addr := fmt.Sprintf("%s:%d", c.sshConfig.Host, port)

	// Connect and perform a handshake
	connection, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}

	c.connection = connection
	return nil
}

// Close closes the connection to the remote server.
func (c *Client) Close() {
	_ = c.connection.Close()
}

// authenticationMethod returns the authentication method name and the authentication method slice based on the provided SSH configuration.
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

		if c.sshConfig.KeyPassword != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(c.sshConfig.KeyPassword))
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
