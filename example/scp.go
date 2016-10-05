package main

import (
	"github.com/calrsom/easyssh"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		Server:   "10.1.1.10",
		Port:     "22",
		User:     "root",
		Password: "TjSDBkAu",
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

		response, _, _ = ssh.Run("rm run.go")

		print(response)

		response, _, _ = ssh.Run("ls -al run.go")

		print(response)
	}
}
