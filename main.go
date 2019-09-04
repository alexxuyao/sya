package main

import (
	"flag"
	"strings"
)

var (
	server      bool   // 以服务器方式运行
	port        int    // 以服务器运行时的端口
	shareDomain string // 以服务端运行时，要共享cookie的域名
	client      bool   // 以客户端运行
	s           string //以客户端运行时，要连接的服务器
	chromePath  string //chrome在本机的路径
)

func init() {
	flag.BoolVar(&server, "server", false, "run as server")
	flag.IntVar(&port, "port", 8332, "the port to listen when run as server, gather than 1024")
	flag.StringVar(&shareDomain, "shareDomain", "baidu.com,sina.com", "domains to share, split by ',', when run as server")

	flag.BoolVar(&client, "client", false, "run as client")
	flag.StringVar(&s, "s", "localhost:8332", "the server to connect when run as client, eg. localhost:8332")
	flag.StringVar(&chromePath, "chromePath", "", "the chrome exe to run")
}

func main() {

	if !client && !server {
		flag.Usage()
		return
	}

	if "" == chromePath {
		chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}

	if server {
		if port < 1024 {
			flag.Usage()
			return
		}

		if "" == shareDomain {
			flag.Usage()
			return
		}
	}

	if server {
		runAsServer(port, strings.Split(shareDomain, ","), chromePath)
	} else {
		runAsClient()
	}

}

func runAsClient() {

}
