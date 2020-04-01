package simple_ssh_toolkit

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

func CreateSession(clientConnection *ssh.Client, stdout, stderr *bytes.Buffer) (*ssh.Session, error) {
	sesscion, err := clientConnection.NewSession()
	if err != nil {
		return nil, err
	}
	sesscion.Stdout = stdout
	sesscion.Stderr = stderr
	return sesscion, nil
}

func CreateConnection(serverAddress, port string, config *ssh.ClientConfig) (*ssh.Client, error) {
	server := serverAddress + ":" + port
	return  ssh.Dial("tcp", server, config)
}

func CreatePrivateKeyConfig(userName, privateKeyPath string,
	hostKeyCallBack func(hostname string, remote net.Addr, key ssh.PublicKey) error) (*ssh.ClientConfig, error) {

	publicKey, err := publicKeyFile(privateKeyPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			publicKey,
		},
		// golang has change the ssh client behavior
		// to improve security
		// here is the code you verify remote call back
		// return nil means ok not error
		HostKeyCallback: hostKeyCallBack,
	}

	return config,nil
}

func CreatePrivateKeyStringConfig(userName, privateKeyStringPath string,
	hostKeyCallBack func(hostname string, remote net.Addr, key ssh.PublicKey) error) (*ssh.ClientConfig, error)  {
	publicKey, err := publicKeyString([]byte(privateKeyStringPath))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			publicKey,
		},
		// golang has change the ssh client behavior
		// to improve security
		// here is the code you verify remote call back
		// return nil means ok not error
		HostKeyCallback: hostKeyCallBack,
	}

	return config,nil
}

func CreateUserPasswordConfig(userName , password string,
	hostKeyCallBack func(hostname string, remote net.Addr, key ssh.PublicKey) error) *ssh.ClientConfig {

	return &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// golang has change the ssh client behavior
		// to improve security
		// here is the code you verify remote call back
		// return nil means ok not error
		HostKeyCallback: hostKeyCallBack,
	}
}

func publicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return publicKeyString(buffer)
}

func publicKeyString(content []byte) (ssh.AuthMethod, error) {
	key, errPrivate := ssh.ParsePrivateKey(content)
	if errPrivate != nil {
		return nil, errPrivate
	}
	return ssh.PublicKeys(key), nil
}

func RunSshCommand(session *ssh.Session, commands []string,) error {
	stdin, errPipe := session.StdinPipe()
	if errPipe != nil {
		return errPipe
	}

	errShell := session.Shell()
	if errShell != nil {
		return errShell
	}

	for _, cmd := range commands {
		_, errExec := fmt.Fprintf(stdin, "%s\n", cmd)
		if errExec != nil {
			return errExec
		}
	}
	errWait := session.Wait()
	if errWait != nil {
		return errWait
	}
	return nil
}

func ScpFileWithString(session *ssh.Session, contentToWrite, remoteFileName, remoteFilePath, remoteFilePermission string) (int, error) {
	file := strings.NewReader(contentToWrite)
	return ScpFile(session, file.Size(), file, remoteFileName, remoteFilePath, remoteFilePermission)
}

func ScpFile(session *ssh.Session, fileSize int64, file io.Reader, remoteFileName, remoteFilePath, remoteFilePermission string) (int, error) {
	exeReturnCode := 0
	var copyScpError error
	go func() {
		stdin, err := session.StdinPipe()
		if err != nil {
			log.Println(err)
			return
		}
		defer stdin.Close()
		exeReturnCode, copyScpError = fmt.Fprintln(stdin, fmt.Sprintf("C%s", remoteFilePermission), fileSize, remoteFileName)
		if copyScpError != nil {
			log.Println(copyScpError)
			return
		}
		result, errCopy := io.CopyN(stdin, file, fileSize)
		exeReturnCode = int(result)
		if errCopy != nil {
			copyScpError = errCopy
			log.Println(errCopy)
			return
		}
		exeReturnCode, copyScpError = fmt.Fprint(stdin, "\x00")
		if copyScpError != nil {
			log.Println(copyScpError)
			return
		}
	}()
	if err := session.Run(fmt.Sprintf("/usr/bin/scp -qtr %s", remoteFilePath)); err != nil {
		return 1, err
	}
	return exeReturnCode, nil
}
