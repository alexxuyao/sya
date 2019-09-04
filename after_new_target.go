package main

import (
	"encoding/json"
	"github.com/alexxuyao/chrome-dominate"
	"log"
)

type SyaAfterNewTarget struct {
	ChromeDominate           *chromedominate.ChromeDominate
	ResponseReceivedListener []ResponseReceivedListener
}

func (s *SyaAfterNewTarget) AfterNewChromeDominateTargetCreate(c *chromedominate.ChromeTargetDominate) {

	// 启动页面事件监听
	err := c.EnablePage()
	if err != nil {
		return
	}

	// 启动网络监听
	err = c.EnableNetwork(chromedominate.NetworkEnableParam{})
	if err != nil {
		return
	}

	c.AddListener(s)
}

func (s *SyaAfterNewTarget) OnMessage(method string, message []byte) {
	if chromedominate.EventPageWindowOpen == method {
		if nil != s.ChromeDominate {
			if err := s.ChromeDominate.InitPageTargets(); err != nil {
				log.Println(err)
			}
		}
	} else if chromedominate.EventNetworkResponseReceived == method {
		data := chromedominate.NetworkResponseReceived{}
		ret := chromedominate.CmdRootType{
			Params: &data,
		}

		err := json.Unmarshal(message, &ret)
		if err != nil {
			log.Println(err)
			return
		}

		for _, listener := range s.ResponseReceivedListener {
			listener.OnResponseReceived(&data)
		}
	}

}
