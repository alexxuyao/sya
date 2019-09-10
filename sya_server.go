package main

import (
	"bufio"
	"encoding/json"
	"github.com/alexxuyao/chrome-dominate"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func runAsServer(port int, shareDomains []string, chromePath string) {

	config := chromedominate.DominateConfig{
		ChromePath: chromePath,
	}

	filter := &SyaAfterNewTarget{}
	config.AfterNewChromeDominateTarget = append(config.AfterNewChromeDominateTarget, filter)

	c, err := chromedominate.NewServerAgentChromeDominate(config)

	if err != nil {
		log.Println(err)
		return
	}

	server, err := NewShareServer(c, port, shareDomains)

	if err != nil {
		log.Println(err)
		return
	}

	filter.ChromeDominate = c
	filter.ResponseReceivedListener = append(filter.ResponseReceivedListener, server)

	//
	ch := make(chan int)
	<-ch
}

type ShareServer struct {
	port       int
	dominate   *chromedominate.ChromeDominate
	domains    []string
	cookies    []chromedominate.Cookie
	clientConn []*net.TCPConn
	mutex      *sync.Mutex
}

func NewShareServer(dominate *chromedominate.ChromeDominate, port int, domains []string) (*ShareServer, error) {

	s := &ShareServer{
		port:     port,
		dominate: dominate,
		domains:  domains,
		mutex:    new(sync.Mutex),
	}

	if err := s.InitServer(); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *ShareServer) OnResponseReceived(data *chromedominate.NetworkResponseReceived) {
	for k, v := range data.Response.Headers {
		log.Println("received:" + k + ", value :" + v)
	}
}

func (c *ShareServer) InitServer() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+strconv.Itoa(c.port))

	if err != nil {
		return err
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	//defer func() {
	//	err = tcpListener.Close()
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//}()

	//循环接收客户端的连接，创建一个协程具体去处理连接
	go func() {
		for {
			tcpConn, err := tcpListener.AcceptTCP()

			if err != nil {
				log.Println(err)
				continue
			}

			log.Println("A client connected :" + tcpConn.RemoteAddr().String())
			go c.handlerClientConn(tcpConn)
		}
	}()

	// 注册监听，每当页面加载时，都把cookie取出来，发到客户端
	go func() {

		for {

			for {

				target, err := c.dominate.GetOneTarget()
				if err != nil {
					log.Println("can not get one target", err)
					break
				}

				if !target.IsAlive {
					log.Println("target is not alive")
					break
				}

				cookies, err := target.GetAllCookies()
				if err != nil {
					log.Println("can not get all cookies", err)
					break
				}

				c.cookies = cookies

				for _, conn := range c.clientConn {
					c.SendCookies(conn)
				}

				break
			}

			time.Sleep(60 * time.Second)
		}

	}()

	return nil
}

//具体处理连接过程方法
func (c *ShareServer) handlerClientConn(conn *net.TCPConn) {

	c.openClient(conn)

	//tcp连接的地址
	ipStr := conn.RemoteAddr().String()

	defer func() {
		c.closeClient(conn)
		log.Println(" Disconnected : " + ipStr)
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()

	//获取一个连接的reader读取流
	reader := bufio.NewReader(conn)

	// 写入全部cookies
	c.SendCookies(conn)

	//接收并返回消息
	for {
		message, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}

		log.Println(string(message))
	}
}

func (c *ShareServer) SendCookies(conn *net.TCPConn) {
	writer := bufio.NewWriter(conn)

	for _, cookie := range c.cookies {

		share := false
		for _, domain := range c.domains {
			if strings.Contains(cookie.Domain, domain) {
				share = true
			}
		}

		if !share {
			continue
		}

		message := Message{
			Type: TypeCookie,
			Data: cookie,
		}

		log.Println("share cookie :", cookie)

		if jStr, err := json.Marshal(message); err == nil {
			jLen := int64(len(jStr))

			b, err := IntToBytes(jLen)
			if err != nil {
				log.Println(err)
				continue
			}

			b = append(b, jStr...)

			if _, err := writer.Write(b); err != nil {
				log.Println(err)
			}

			if err = writer.Flush(); err != nil {
				log.Println(err)
			}

		} else {
			log.Println(err)
		}

	}
}

func (c *ShareServer) openClient(conn *net.TCPConn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.clientConn = append(c.clientConn, conn)
}

func (c *ShareServer) closeClient(conn *net.TCPConn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for k, v := range c.clientConn {
		if v == conn {
			c.clientConn = append(c.clientConn[:k], c.clientConn[k+1:]...)
			break
		}
	}
}
