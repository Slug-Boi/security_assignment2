package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func shareGenerator(secret, NOS int) []int {
	share := make([]int, NOS)

	leftoverShare := secret

	for i := 0; i < NOS-1; i++ {
		random_big, err := rand.Int(rand.Reader, big.NewInt(int64(leftoverShare)))
		if err != nil {
			log.Println(err)
			return nil
		}
		random := int(random_big.Int64())

		leftoverShare -= random
		share[i] = random
	}
	share[NOS-1] = leftoverShare
	return share
}

var shares []int
var hosCh []chan bool
var computations []int
var returnedComputations []int
var mutex sync.Mutex

func main() {
	args := os.Args[1:]
	cert := args[0] + ".crt"
	key := args[0] + ".key"
	ports := strings.Split(args[3], ",")
	secret, err := strconv.Atoi(args[2])
	if err != nil {
		log.Println(err)
		return
	}
	Entry(cert, key, args[1], secret, ports)

	for {

	}
}

func Entry(certName, keyName, clientPort string, secret int, ports []string) {

	shares = shareGenerator(secret, len(ports)+1)

	for _, share := range shares {
		log.Println(clientPort, share)
	}

	computations = append(computations, shares[0])

	cer, err := tls.LoadX509KeyPair(certName, keyName)
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	go listener(config, clientPort)

	time.Sleep(1 * time.Second)
	for i, port := range ports {
		ch := make(chan bool)
		hosCh = append(hosCh, ch)
		share := shares[i+1]
		//log.Println(clientPort,port,strconv.Itoa(share))
		go dialer(config, port, ch, share)
	}

}

func hospitalHandler(conn net.Conn) {
	defer conn.Close()

	for _, ch := range hosCh {
		ch <- true
	}

	for {
		if len(returnedComputations) == len(hosCh) {
			sum_share := 0
			for _, comp := range computations {
				sum_share += comp
			}
			sum := sum_share
			for _, comp := range returnedComputations {
				sum += comp
			}
			returnedComputations = []int{}

			conn.Write([]byte(strconv.Itoa(sum) + "\n"))
			break
		}
		time.Sleep(2 * time.Second)
	}

}

func handler(conn net.Conn, r *bufio.Reader) {
	defer conn.Close()

	// we know all future messages are going to be computations
	msg, err := r.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}

	msg = strings.Trim(msg, "\n")

	value, err := strconv.Atoi(msg)
	if err != nil {
		log.Println(err)
		return
	}
	mutex.Lock()
	returnedComputations = append(returnedComputations, value)
	mutex.Unlock()

}

func dialer(config *tls.Config, port string, ch chan bool, share int) {
	//log.Println("Dialing on port", port)
	var conn net.Conn
	var err error
	for {
		conn, err = tls.Dial("tcp", "localhost:"+port, config)
		if err != nil {
			log.Println(err)
		} else {
			break
		}
		time.Sleep(2 * time.Second)
	}

	conn.Write([]byte(strconv.Itoa(share) + "\n"))

	for {
		accept := <-ch
		if accept {
			sum_share := 0
			for _, comp := range computations {
				sum_share += comp
			}
			if sum_share != 0 {
				log.Println("sending sum to client", sum_share, port)
				conn.Write([]byte(strconv.Itoa(sum_share) + "\n"))
			}
			break
		}
	}
}

func listener(config *tls.Config, port string) {
	ln, err := tls.Listen("tcp", port, config)
	if err != nil {
		log.Println(err)
		return
	}

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		r := bufio.NewReader(conn)
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		if msg == "go\n" {
			log.Println("Hospital connected" + port)
			go hospitalHandler(conn)
		} else {
			log.Println("Client connected" + port)
			log.Println("Message received:", msg)
			msg = strings.Trim(msg, "\n")

			value, err := strconv.Atoi(msg)
			if err != nil {
				log.Println(err)
				return
			}
			computations = append(computations, value)
			go handler(conn, r)
		}
		msg = ""
	}
}
