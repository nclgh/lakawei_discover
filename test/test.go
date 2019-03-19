package main

import (
	"time"
	"github.com/nclgh/lakawei_discover"
	"fmt"
)

func main() {
	lakawei_discover.Register("test", "127.0.0.1:6666")
	addr := lakawei_discover.GetServiceAddr("test")
	fmt.Printf("%v", addr)
	time.Sleep(10 * time.Hour)
}
