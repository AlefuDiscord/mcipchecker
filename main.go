package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
)


type status struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	Favicon Icon
	Delay   time.Duration
}

type Icon string

func (i Icon) ToImage() (icon image.Image, err error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(string(i), prefix) {
		return nil, fmt.Errorf("server icon should prepended with %q", prefix)
	}
	base64png := strings.TrimPrefix(string(i), prefix)
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))
	icon, err = png.Decode(r)
	return
}

var outTemp = template.Must(template.New("output").Parse(`
	Version: [{{ .Version.Protocol }}] {{ .Version.Name }}
	Description:
{{ .Description }}
	Delay: {{ .Delay }}
	Players: {{ .Players.Online }}/{{ .Players.Max }}{{ range .Players.Sample }}
	- [{{ .Name }}] {{ .ID }}{{ end }}
`))

func (s *status) String() string {
	var sb strings.Builder
	err := outTemp.Execute(&sb, s)
	if err != nil {
		panic(err)
	}
	return sb.String()
}

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n%s [-f] [-p] <address>[:port]\n", os.Args[0])
	flag.PrintDefaults()
}
func main() {
	os.Create("connect.txt")
	runtime.GOMAXPROCS(runtime.NumCPU())
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func(i int) {
			for {
				defer wg.Done()
				var existingIPs []string
				ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
				ipandport := fmt.Sprint(ip + ":25565")
				for _, existingIP := range existingIPs {
					if ip == existingIP {
						fmt.Println("IP already exists")
						continue
					}
				}
				parsedIP := net.ParseIP(ip)
				if parsedIP == nil {
					continue
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
						continue
					}
					conn.SetDeadline(time.Now().Add(5 * time.Second))
					fmt.Fprintf(conn, "\xFE\x01")

					scanner := bufio.NewScanner(conn)
					for scanner.Scan() {
						line := scanner.Text()
						if len(line) > 0 && line[0] == 0xFF {
							resp, delay, err := bot.PingAndList(ipandport)
							if err != nil {
								fmt.Printf("Ping and list server fail: %v", err)
								continue
							}

							var s status
							err = json.Unmarshal(resp, &s)
							if err != nil {
								fmt.Print("Parse json response fail:", err)
								continue
							}
							s.Delay = delay

							fmt.Print(&s)
							sum := fmt.Sprintln(&s)
							file, err := os.OpenFile("connect.txt", os.O_WRONLY|os.O_APPEND, 0644)
							if err != nil {
								fmt.Println(err)
							}
							file.WriteString(ipandport + "\n" + sum)
							continue
						}
					}
				}
			}
		}(i)
	}
	wg.Wait()
}
