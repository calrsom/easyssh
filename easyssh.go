// Package easyssh provides a simple implementation of some SSH protocol
// features in Go. You can simply run a command on a remote server or get a file
// even simpler than native console SSH client. You don't need to think about
// Dials, sessions, defers, or public keys... Let easyssh think about it!
package easyssh

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var keyMap map[string][]byte = make(map[string][]byte)

// Contains main authority information.
// User field should be a name of user on remote server (ex. john in ssh john@example.com).
// Server field should be a remote machine address (ex. example.com in ssh john@example.com)
// Key is a path to private key on your local machine.
// Port is SSH server port on remote machine.
// Note: easyssh looking for private key in user's home directory (ex. /home/john + Key).
// Then ensure your Key begins from '/' (ex. /.ssh/id_rsa)
type MakeConfig struct {
	User      string
	Server    string
	Key       string
	Port      string
	Password  string
	EnablePTY bool
	Update    bool
	client    *ssh.Client
}

// returns ssh.Signer from user you running app home path + cutted key path.
// (ex. pubkey,err := getKeyFile("/.ssh/id_rsa") )
func getKeyFile(keypath string) (pubkey ssh.Signer, err error) {
	var buf []byte
	var ok bool
	if buf, ok = keyMap[keypath]; !ok {
		file := keypath
		buf, err = ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		keyMap[keypath] = buf
	}
	pubkey, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return pubkey, nil

}

// connects to remote server using MakeConfig struct and returns *ssh.Session
func (ssh_conf *MakeConfig) connect() (*ssh.Session, error) {
	// auths holds the detected ssh auth methods
	auths := []ssh.AuthMethod{}

	// figure out what auths are requested, what is supported
	if ssh_conf.Password != "" {
		auths = append(auths, ssh.Password(ssh_conf.Password))

		auths = append(auths, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			// Just send the password back for all questions
			answers := make([]string, len(questions))
			for i, _ := range answers {
				answers[i] = ssh_conf.Password // replace this
			}
			return answers, nil
		}))
	}

	// Default port 22
	if ssh_conf.Port == "" {
		ssh_conf.Port = "22"
	}

	// Default current user
	if ssh_conf.User == "" {
		ssh_conf.User = os.Getenv("USER")
	}

	//	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
	//		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	//		defer sshAgent.Close()
	//	}

	if pubkey, err := getKeyFile(ssh_conf.Key); err == nil {
		auths = append(auths, ssh.PublicKeys(pubkey))
	}

	config := &ssh.ClientConfig{
		User: ssh_conf.User,
		Auth: auths,
	}

	if ssh_conf.Update {
		if ssh_conf.client != nil {
			ssh_conf.client.Close()
		}
		var err error
		ssh_conf.client, err = ssh.Dial("tcp", ssh_conf.Server+":"+ssh_conf.Port, config)
		if err != nil {
			return nil, err
		}
		ssh_conf.Update = false
	}

	session, err := ssh_conf.client.NewSession()
	if err != nil {
		return nil, err
	}

	if ssh_conf.EnablePTY {
		modes := ssh.TerminalModes{
			ssh.ECHO:          0,
			ssh.TTY_OP_ISPEED: 144000,
			ssh.TTY_OP_OSPEED: 144000,
		}

		if err := session.RequestPty("xterm", 1024, 1024, modes); err != nil {
			session.Close()
			return nil, err
		}
	}

	return session, nil
}

// Stream returns one channel that combines the stdout and stderr of the command
// as it is run on the remote machine, and another that sends true when the
// command is done. The sessions and channels will then be closed.
func (ssh_conf *MakeConfig) Stream(command string) (output chan string, done chan bool, err, sessionErr error) {
	// connect to remote host
	session, err := ssh_conf.connect()
	if err != nil {
		return output, done, err, sessionErr
	}
	outReader, err := session.StdoutPipe()
	if err != nil {
		return output, done, err, sessionErr
	}
	errReader, err := session.StderrPipe()
	if err != nil {
		return output, done, err, sessionErr
	}
	// combine outputs, create a line-by-line scanner
	outputReader := io.MultiReader(outReader, errReader)
	sessionErr = session.Run(command)
	scanner := bufio.NewScanner(outputReader)
	// continuously send the command's output over the channel
	outputChan := make(chan string)
	done = make(chan bool)
	go func(scanner *bufio.Scanner, out chan string, done chan bool) {
		defer close(outputChan)
		defer close(done)
		for scanner.Scan() {
			outputChan <- scanner.Text()
		}
		// close all of our open resources
		done <- true
		session.Close()
	}(scanner, outputChan, done)
	return outputChan, done, err, sessionErr
}

// Runs command on remote machine and returns its stdout as a string
func (ssh_conf *MakeConfig) Run(command string) (outStr, errStr string, err error) {
	outChan, doneChan, err, _ := ssh_conf.Stream(command)
	if err != nil {
		ssh_conf.Update = true
		outChan, doneChan, err, _ = ssh_conf.Stream(command)
		if err != nil {
			return outStr, errStr, err
		}
	}
	// read from the output channel until the done signal is passed
	stillGoing := true
	for stillGoing {
		select {
		case <-doneChan:
			stillGoing = false
		case line := <-outChan:
			outStr += line + "\n"
		}
	}
	/*
		if sessionErr != nil && outStr != "" {
			errStr = outStr
			outStr = ""
		}
	*/
	// return the concatenation of all signals from the output channel
	return outStr, errStr, err
}

// Scp uploads sourceFile to remote machine like native scp console app.
func (ssh_conf *MakeConfig) Scp(sourceFile string, destDir string) error {
	session, err := ssh_conf.connect()

	if err != nil {
		return err
	}
	defer session.Close()

	targetFile := filepath.Base(sourceFile)

	src, srcErr := os.Open(sourceFile)

	if srcErr != nil {
		return srcErr
	}

	srcStat, statErr := src.Stat()

	if statErr != nil {
		return statErr
	}

	go func() {
		w, _ := session.StdinPipe()

		fmt.Fprintln(w, "C0644", srcStat.Size(), targetFile)

		if srcStat.Size() > 0 {
			io.Copy(w, src)
			fmt.Fprint(w, "\x00")
			w.Close()
		} else {
			fmt.Fprint(w, "\x00")
			w.Close()
		}
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s/%s", destDir, targetFile)); err != nil {
		return err
	}

	return nil
}
