package main

import (
	"fmt"
	"testing"
)

func TestIntToBytes(t *testing.T) {
	i := int64(62)
	b, err := IntToBytes(i)

	if err != nil {
		t.Error(err)
	}

	u, err := BytesToInt(b)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(u)
}
