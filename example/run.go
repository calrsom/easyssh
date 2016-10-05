package main

import (
	"github.com/calrsom/easyssh"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:     "root",
		Server:   "10.1.1.10",
		Password: "TjSDBkAu",
		Port:     "22",
	}

	// Call Run method with command you want to run on remote server.
	response, _, err := ssh.Run("ls")
	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		print(response)
	}
}
