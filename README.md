# simple ssh toolkit

A simple ssh tool base on `golang.org/x/crypto/ssh` for copy file or run command on remote server

## Get start

### Get Source Code

```console
$ go get github.com/simonyangcj/simple-ssh-toolkit
```

### Basic Usage

```go
package main

import (
	"bytes"
	sshPipe "github.com/simonyangcj/simple-ssh-toolkit"
	"golang.org/x/crypto/ssh"
	"log"
)

func main() {
    //config, errConfig := sshPipe.CreateUserPasswordConfig("user","password", ssh.InsecureIgnoreHostKey())
	config, errConfig := sshPipe.CreatePrivateKeyConfig("root","/root/.ssh/id_rsa", ssh.InsecureIgnoreHostKey())
	if errConfig != nil {
		log.Fatal(errConfig)
	}
	clientConnection, errCon := sshPipe.CreateConnection("10.0.0.28", "22", config)
	if errCon != nil {
		log.Fatal(errCon)
	}
	defer clientConnection.Close()
	var stdout, stderr bytes.Buffer
	session,errSession := sshPipe.CreateSession(clientConnection, &stdout, &stderr)
	if errSession != nil {
		log.Fatal(errSession)
	}
	defer session.Close()
	commands := []string {
		"whoami",
		"exit", // has to set exit signal to remote indicate it`s all
    }
    sshPipe.RunSshCommand(session, commands)
}

```

### Run multiple command on remote server

```go
package main

import (
	"bytes"
	sshPipe "github.com/simonyangcj/simple-ssh-toolkit"
	"golang.org/x/crypto/ssh"
	"log"
)

func main()  {
	commands := []string{
		"pwd",
		"whoami",
		"echo 'bye'",
		"exit",
	}
	config := sshPipe.CreateUserPasswordConfig("root", "trystack", ssh.InsecureIgnoreHostKey())
	clientConnection, errCon := sshPipe.CreateConnection("10.0.0.28", "22", config)
	if errCon != nil {
		log.Fatal(errCon)
	}
	defer clientConnection.Close()
	var stdout, stderr bytes.Buffer
	session,errSession := sshPipe.CreateSession(clientConnection, &stdout, &stderr)
	if errSession != nil {
		log.Fatal(errSession)
	}
	defer session.Close()
	sshPipe.RunSshCommand(session, commands)
	log.Println(stdout.String(), stderr.String())
}

```

### copy local file to remote server

```go
package main

import (
	"bytes"
	sshPipe "github.com/simonyangcj/simple-ssh-toolkit"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

func main()  {
	config := sshPipe.CreateUserPasswordConfig("root", "trystack", ssh.InsecureIgnoreHostKey())
	clientConnection, errCon := sshPipe.CreateConnection("10.0.0.28", "22", config)
	if errCon != nil {
		log.Fatal(errCon)
	}
	defer clientConnection.Close()
	var stdout, stderr bytes.Buffer
	session,errSession := sshPipe.CreateSession(clientConnection, &stdout, &stderr)
	if errSession != nil {
		log.Fatal(errSession)
	}
	defer session.Close()
	file, errFile := os.Open("/root/.ssh/id_rsa")
	if errFile != nil {
		log.Fatal(errFile)
	}
	fileInfo, errInfo := file.Stat()
	if errInfo != nil {
		log.Fatal(errInfo)
	}
	sshPipe.ScpFile(session, fileInfo.Size(), file, "test1", "/root", "0644")
}
```