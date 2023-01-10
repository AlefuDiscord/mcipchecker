package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	os.Create("connectIP.txt")
	runtime.GOMAXPROCS(runtime.NumCPU())
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func(i int) {
			for {
			loop:
				defer wg.Done()
				var existingIPs []string
				ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
				ipandport := fmt.Sprint(ip + ":25565")
				for _, existingIP := range existingIPs {
					if ip == existingIP {
						fmt.Println("IP already exists")
						goto loop
					}
				}
				parsedIP := net.ParseIP(ip)
				if parsedIP == nil {
					goto loop
				}
				_, err := net.Dial("tcp", ipandport)
				if err != nil {
					fmt.Println("接続できませんでした", ipandport)
				} else {
					fmt.Println("接続できました", ipandport)
					fmt.Println("再接続を試みます", ipandport)
					conn, err := net.Dial("tcp", ipandport)
					if err != nil {
						fmt.Println("エラーなのでgotoする")
						goto loop
					}
					conn.SetDeadline(time.Now().Add(5 * time.Second))
					fmt.Fprintf(conn, "\xFE\x01")

					scanner := bufio.NewScanner(conn)
					for scanner.Scan() {
						line := scanner.Text()
						if len(line) > 0 && line[0] == 0xFF {
							file, err := os.OpenFile("connectIP.txt", os.O_WRONLY|os.O_APPEND, 0644)
							if err != nil {
								fmt.Println(err)
							}
							file.WriteString(ipandport + "\n")
							goto loop
						}
					}
				}
			}
		}(i)
	}
	wg.Wait()
}
