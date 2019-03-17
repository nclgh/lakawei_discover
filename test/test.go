package main

import (
	"time"
	"github.com/nclgh/lakawei_discover"
)

func main() {
	lakawei_discover.Register("test","127.0.0.1:6666")
	time.Sleep(10*time.Hour)
}