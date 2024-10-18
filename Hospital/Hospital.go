package main

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

var messages []string
var mutex sync.Mutex

func main() {
	args := os.Args[1:]
	cert := args[0] + ".crt"
	key := args[0] + ".key"
	ports := strings.Split(args[1], ",")
	Entry(cert, key, ports)
	for {

	}
}

func Entry(certName, keyName string, ports []string) {
	cer, err := tls.LoadX509KeyPair(certName, keyName)
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	for _, port := range ports {
		go dialer(config, port)
	}
}

func handler(conn net.Conn) {
	defer conn.Close()

	n, err := conn.Write([]byte("go\n"))
	if err != nil {
		log.Println(n, err)
		return
	}
	log.Println("sending go")

	r := bufio.NewReader(conn)
	msg, err := r.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(msg)

	mutex.Lock()
	messages = append(messages, msg)
	mutex.Unlock()

}

func dialer(config *tls.Config, port string) {
	log.Println("dialing", port)
	conn, err := tls.Dial("tcp", "localhost:"+port, config)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	handler(conn)
}
