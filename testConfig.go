/* switch ssh
 */
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

import (
	"golang.org/x/crypto/ssh"
)

func main() {
	config := &ssh.ClientConfig{
		User: "aaa",
		Auth: []ssh.AuthMethod{
			ssh.Password("aaa"),
		},
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc"},
		},
	}
	// config.Config.Ciphers = append(config.Config.Ciphers, "aes128-cbc")
	clinet, err := ssh.Dial("tcp", "1.1.1.1:22", config)
	checkError(err, "connet 1.1.1.1")

	session, err := clinet.NewSession()
	defer session.Close()
	checkError(err, "creae shell")

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		log.Fatal(err)
	}

	w, err := session.StdinPipe()
	if err != nil {
		panic(err)
	}
	r, err := session.StdoutPipe()
	if err != nil {
		panic(err)
	}
	e, err := session.StderrPipe()
	if err != nil {
		panic(err)
	}

	in, out := MuxShell(w, r, e)
	if err := session.Shell(); err != nil {
		log.Fatal(err)
	}
	<-out //ignore the shell output
	in <- "display version"
	in <- "display arp statistics all"

	fmt.Printf("%s\n", <-out)
	fmt.Printf("%s\n", <-out)

	in <- "quit"
	in <- "quit"
	_ = <-out
	_ = <-out
	session.Wait()
}

func checkError(err error, info string) {
	if err != nil {
		fmt.Printf("%s. error: %s\n", info, err)
		os.Exit(1)
	}
}

func MuxShell(w io.Writer, r, e io.Reader) (chan<- string, <-chan string) {
	in := make(chan string, 5)
	out := make(chan string, 5)
	var wg sync.WaitGroup
	wg.Add(1) //for the shell itself
	go func() {
		for cmd := range in {
			wg.Add(1)
			w.Write([]byte(cmd + "\n"))
			wg.Wait() //控制每次goroutine执行一条命令
		}
	}()

	go func() {
		var (
			buf [1024 * 1024]byte
			t   int
		)
		for {
			n, err := r.Read(buf[t:])
			if err != nil {
				fmt.Println(err.Error())
				close(in)
				close(out)
				return
			}
			t += n
			result := string(buf[:t])
			//处理设备分页符
			if strings.Contains(string(buf[t-n:t]), "More") {
				w.Write([]byte("\n"))
			}
			//完成一条操作清空goroutine
			if strings.Contains(result, "username:") ||
				strings.Contains(result, "password:") ||
				strings.Contains(result, ">") {
				out <- string(buf[:t])
				t = 0
				wg.Done()
			}
		}
	}()
	return in, out
}
