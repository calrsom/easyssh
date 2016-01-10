# easyssh

## Description

Package easyssh provides a simple implementation of some SSH protocol features in Go.
You can simply run command on remote server or upload a file even simple than native console SSH client.
Do not need to think about Dials, sessions, defers and public keys...Let easyssh will be think about it!

## Install


```go
go get github.com/jniltinho/easyssh
```

## So easy to use!

[Run a command on remote server and get STDOUT output](example/run.go)

```go
package main

import (
	"github.com/jniltinho/easyssh"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:   "user1",
		Server: "server.example.com",
		Password:  "<user1password>",
		Port: "22",
	}

	// Call Run method with command you want to run on remote server.
	response, err := ssh.Run("ps ax")
	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		print(response)
	}
}
```

## Upload a file
[Upload a file to remote server](example/scp.go)

```go
package main

import (
	"github.com/jniltinho/easyssh"
)

func main() {
	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		User:   "user1",
		Server: "server.example.com",
		Password:  "<user1password>",
		Port: "22",
	}

	// Call Scp method with file you want to upload to remote server.
	err := ssh.Scp("/home/linuxpro/GO/src/goclientssh.go")

	// Handle errors
	if err != nil {
		panic("Can't run remote command: " + err.Error())
	} else {
		println("success")

		response, _ := ssh.Run("ls -al zipkin.rb")

		print(response)
	}
}
```


## SSH Error

```
panic: Can't run remote command: ssh: handshake failed: 
ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain
```

```bash
## Change file: /etc/ssh/sshd_config
sed -i 's|PasswordAuthentication no|PasswordAuthentication yes|' /etc/ssh/sshd_config


## Or add the end of the file.
## Match address 192.168.1.0/24
##    PasswordAuthentication yes

## And restart sshd service
sudo service sshd restart
```



## Thank 

@hypersleep: Vladislav Spirenkov -> https://github.com/hypersleep/easyssh
