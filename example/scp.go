package main

import (
	"github.com/calrsom/easyssh"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:   "root",
		Server: "10.1.1.10",
		Key:    "TjSDBkAu",
		Port:   "22",
	}

	// Call Scp method with file you want to upload to remote server.
	err := ssh.Scp("run.go", "~")

	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		println("success")

		response, _, _ := ssh.Run("ls -al run.go")

		print(response)
	}
}
