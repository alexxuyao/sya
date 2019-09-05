package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/alexxuyao/chrome-dominate"
	"log"
	"net"
	"strings"
)

func runAsClient(serverAddress string, chromePath string) {

	config := chromedominate.DominateConfig{
		ChromePath: chromePath,
	}

	filter := &SyaAfterNewTarget{}
	config.AfterNewChromeDominateTarget = append(config.AfterNewChromeDominateTarget, filter)

	c, err := chromedominate.NewChromeDominate(config)

	if err != nil {
		log.Println(err)
		return
	}

	server, err := NewShareClient(c, serverAddress)

	if err != nil {
		log.Println(err)
		return
	}

	filter.ChromeDominate = c
	filter.ResponseReceivedListener = append(filter.ResponseReceivedListener, server)

	target, err := c.GetOneTarget()

	if err != nil {
		log.Println(err, "new chrome dominate error")
		fmt.Println(err)
		return
	}

	ret, err := target.OpenPage("https://www.baidu.com/")

	if err != nil {
		log.Println(err, "open baidu error")
		fmt.Println(err)
		return
	}

	log.Println(ret.FrameId)

	ch := make(chan int)
	<-ch

}

type ShareClient struct {
	serverAddress string
	dominate      *chromedominate.ChromeDominate
	domains       []string
	serverConn    *net.TCPConn
}

func NewShareClient(dominate *chromedominate.ChromeDominate, serverAddress string) (*ShareClient, error) {

	s := &ShareClient{
		serverAddress: serverAddress,
		dominate:      dominate,
	}

	if err := s.InitClient(); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *ShareClient) OnResponseReceived(data *chromedominate.NetworkResponseReceived) {
	for k, v := range data.Response.Headers {
		log.Println("received:" + k + ", value :" + v)
	}
}

func (c *ShareClient) InitClient() error {
	return c.InitConn()
}

func (c *ShareClient) InitConn() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.serverAddress)

	if err != nil {
		return err
	}

	c.serverConn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}

	// 处理conn
	go c.handlerServerMessage(c.serverConn)

	return nil
}

//具体处理连接过程方法
func (c *ShareClient) handlerServerMessage(conn *net.TCPConn) {

	//tcp连接的地址
	ipStr := conn.RemoteAddr().String()

	defer func() {
		fmt.Println(" Disconnected : " + ipStr)
		if err := conn.Close(); err != nil {
			log.Println(err)
		}

		// 尝试重新连接
		if err := c.InitConn(); err != nil {
			log.Println(err)
		}
	}()

	//获取一个连接的reader读取流
	reader := bufio.NewReader(conn)
	// writer := bufio.NewWriter(conn)

	// 取int的byte数组长度
	b, err := IntToBytes(int64(123))
	if err != nil {
		log.Println(err)
		return
	}

	intLen := len(b)

	//接收并返回消息
	tmpBuffer := make([]byte, 0) // 用于缓存消息
	messageLen := int64(-1)      // 消息的长度，从协议中解析出来
	appendMessage := false       // 是否追加
	buffer := make([]byte, 1024) // 临时桶，用于从流中装消息

	for {
		mLen, err := reader.Read(buffer)
		if err != nil {
			break
		}

		tmpBuffer = append(tmpBuffer, buffer[:mLen]...)

		for {
			if appendMessage {

				// 如果是追加的，说明上次拿到的消息有粘包或不全的包

				if int64(len(tmpBuffer)) >= messageLen {

					// 如果缓存长度大于等于消息长度，说明最少有一条消息可以进行解析了

					message := tmpBuffer[:messageLen]

					// 取出消息，处理消息
					c.handlerMessage(message)

					// 把追加消息重置为非追加
					appendMessage = false
					tmpBuffer = tmpBuffer[messageLen:]
					messageLen = -1

					// 如果取完消息后，缓存为0，那么break进入下一次读取
					if len(tmpBuffer) <= 0 {
						break
					}

					// 如果取完消息后，缓存不为0，说明缓存中可能还有另一条消息
					// 这时以非追加方式，进入下一次循环

				} else {

					// 如果缓存长度小于消息长度，说明消息没接收完，进入下次读取

					break
				}
			} else {

				if len(tmpBuffer) >= intLen {
					// 解析出消息长度
					messageLen, err = BytesToInt(tmpBuffer[:intLen])
					if err != nil {
						log.Println(err)
					}

					tmpBuffer = tmpBuffer[intLen:]
					appendMessage = true

				} else {
					break
				}
			}
		}

	}
}

func makeCookieUrl(c chromedominate.Cookie) string {
	u := ""

	if c.Secure {
		u = "https://"
	} else {
		u = "http://"
	}

	if strings.HasPrefix(c.Domain, ".") {
		u = u + c.Domain[1:] + c.Path
	} else {
		u = u + c.Domain + c.Path
	}

	return u
}

func (c *ShareClient) handlerMessage(message []byte) {
	log.Println("handle message:" + string(message))

	mString := string(message)
	mString = mString[strings.Index(mString, ":")+1 : strings.Index(mString, ",")]
	mType := mString[1 : len(mString)-1]

	if TypeCookie == mType {
		cookie := chromedominate.Cookie{}
		m := Message{
			Data: &cookie,
		}

		if err := json.Unmarshal(message, &m); err != nil {
			log.Println(err)
			return
		}

		target, err := c.dominate.GetOneTarget()
		if err != nil {
			log.Println(err)
			return
		}

		u := makeCookieUrl(cookie)

		if ret, err := target.SetCookie(chromedominate.CookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   &cookie.Domain,
			Path:     &cookie.Path,
			Secure:   &cookie.Secure,
			HttpOnly: &cookie.HttpOnly,
			SameSite: cookie.SameSite,
			Expires:  &cookie.Expires,
			Url:      &u,
		}); err != nil {
			log.Println(err)
			return
		} else {
			if !ret {
				log.Println("set cookie result:", ret)
			}
		}

	}
}
