/* switch ssh
 */
package main

import (
	"flag"
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
	//go run ./testConfig.go --username="aaa" --passwd='aaa' --ip_port="192.168.6.87" --cmd='display version'
	username := flag.String("username", "aaa", "username")
	passwd := flag.String("passwd", "aaa", "password")
	ip_port := flag.String("ip_port", "1.1.1.1:22", "ip and port")
	cmdstring := flag.String("cmd", "display arp statistics all", "cmdstring")
	flag.Parse()
	fmt.Println("username:", *username)
    fmt.Println("passwd:", *passwd)
    fmt.Println("ip_port:", *ip_port)
    fmt.Println("cmdstring:", *cmdstring)
	config := &ssh.ClientConfig{
		User: *username,
		Auth: []ssh.AuthMethod{
			ssh.Password(*passwd),
		},
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc"},
		},
	}
	// config.Config.Ciphers = append(config.Config.Ciphers, "aes128-cbc")
	clinet, err := ssh.Dial("tcp", *ip_port, config)
	checkError(err, "connet " + *ip_port)

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
	in <- *cmdstring

	fmt.Printf("%s\n", <-out)

	in <- "quit"
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
			wg.Wait()
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
			if strings.Contains(string(buf[t-n:t]), "More") {
				w.Write([]byte("\n"))
			}
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
