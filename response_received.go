package main

import "github.com/alexxuyao/chrome-dominate"

type ResponseReceivedListener interface {
	OnResponseReceived(data *chromedominate.NetworkResponseReceived)
}
