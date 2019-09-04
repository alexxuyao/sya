package main

const TypeCookie = "cookie"

type Message struct {
	Type string      `json:"type"` // set const type
	Data interface{} `json:"data"`
}
