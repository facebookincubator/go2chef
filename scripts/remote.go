package main

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"

	"github.com/spf13/pflag"
)

func binName() string {
	if runtime.GOOS == "windows" {
		return "go2chef.exe"
	}
	return "go2chef"
}

var (
	targets  = pflag.StringArrayP("target", "T", []string{}, "target to remotely execute on")
	username = pflag.StringP("username", "u", "root", "username to connect as")
	password = pflag.StringP("password", "p", "root", "password to connect with")
	binary   = pflag.StringP("binary", "b", fmt.Sprintf("build/%s/%s/%s", runtime.GOOS, runtime.GOARCH, binName()), "binary to copy")
	port     = pflag.StringP("port", "P", "22", "port")
	windows  = pflag.BoolP("windows", "W", false, "enable windows mode")
	cfg      = pflag.StringP("config", "c", "", "config to copy")
	bundles  = pflag.StringArrayP("bundle", "B", []string{}, "bundles to copy")
)

func main() {
	pflag.Parse()

	config := &ssh.ClientConfig{
		User: *username,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if *windows && !strings.HasSuffix(*binary, ".exe") {
		*binary = *binary + ".exe"
	}

	var wg sync.WaitGroup
	for _, target := range *targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			doTarget(t, config)
		}(target)
	}
	wg.Wait()
}

func doTarget(target string, config *ssh.ClientConfig) {
	client, err := ssh.Dial("tcp", net.JoinHostPort(target, *port), config)
	if err != nil {
		log.Printf("ERROR: failed to connect to %s: %s", target, err)
		return
	}
	defer client.Close()

	c, err := sftp.NewClient(client)
	if err != nil {
		log.Printf("ERROR: failed to establish sftp to %s: %s", target, err)
		return
	}
	defer c.Close()

	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		log.Printf("ERROR: failed to create temp dir: %s", err)
		return
	}

	binPath, err := copySFTP(c, tmp, *binary, 0755)
	if err != nil {
		log.Printf("ERROR: failed to copy sftp to %s: %s", target, err)
		return
	}
	log.Printf("copied %s to %s", *binary, binPath)

	cfgPath, err := copySFTP(c, tmp, *cfg, 0644)
	if err != nil {
		log.Printf("ERROR: failed to copy sftp to %s: %s", target, err)
		return
	}
	log.Printf("copied %s to %s", *cfg, cfgPath)

	for _, b := range *bundles {
		bPath, err := copyBundle(c, tmp, b)
		if err != nil {
			log.Printf("ERROR: failed to copy sftp to %s: %s", b, err)
			return
		}
		log.Printf("copied %s to %s", b, bPath)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Printf("ERROR: failed to create session on %s: %s", target, err)
		return
	}
	defer session.Close()

	session.Stderr = os.Stderr
	session.Stdout = os.Stdout

	if err := session.Run(
		fmt.Sprintf(
			"cd %s && %s --log-level DEBUG --local-config %s",
			filepath.Dir(binPath), binPath, cfgPath,
		),
	); err != nil {
		log.Printf("ERROR: failed to run go2chef remotely: %s", err)
		return
	}
}

func copySFTP(c *sftp.Client, tmp string, path string, mode os.FileMode) (string, error) {

	if *windows {
		tmp = filepath.Join("/", filepath.Base(tmp))
	}

	realPath := filepath.Join(tmp, filepath.Base(path))
	if err := c.MkdirAll(tmp); err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	f, err := c.OpenFile(realPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	_ = f.Chmod(mode)

	if _, err := f.Write(data); err != nil {
		return "", err
	}

	return realPath, nil
}

func copyBundle(c *sftp.Client, ctmp string, path string) (string, error) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	rName := filepath.Join(tmp, filepath.Base(path)) + ".tar.gz"
	cmd := exec.Command("tar", "czvf", rName, "./")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return copySFTP(c, ctmp, rName, 0644)
}
