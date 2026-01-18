package main

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strings"
)

type RemoteClient struct {
	Client *ssh.Client
}

func Connect(user, addr, password string) (*RemoteClient, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	return &RemoteClient{Client: client}, nil
}

func (rc *RemoteClient) RunScript(script string) error {
	session, err := rc.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	session.Stdin = strings.NewReader(script)

	return session.Run("sh -s")
}

func (rc *RemoteClient) Download(remotepath, localpath string) error {
	sc, err := sftp.NewClient(rc.Client)
	if err != nil {
		return fmt.Errorf("SFTP istemcisi oluşturulamadı: %v", err)
	}
	defer sc.Close()

	srcfile, err := sc.Open(remotepath)
	if err != nil {
		return fmt.Errorf("uzaktaki dosya açılamadı (%s): %v", remotepath, err)
	}
	defer srcfile.Close()

	dstfile, err := os.Create(localpath)
	if err != nil {
		return fmt.Errorf("yerel dosya oluşturulamadı: %v", err)
	}
	defer dstfile.Close()

	_, err = io.Copy(dstfile, srcfile)
	if err != nil {
		return fmt.Errorf("kopyalama hatası: %v", err)
	}

	return nil
}

func (rc *RemoteClient) RunCommand(cmd string) (string, error) {
	session, err := rc.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var b strings.Builder
	session.Stdout = &b
	session.Stderr = &b

	err = session.Run(cmd)
	return b.String(), err
}

func (rc *RemoteClient) Close() {
	rc.Client.Close()
}
