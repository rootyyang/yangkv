package system

import (
	"fmt"
	"net"
	"sync"
)

type localAddr struct {
	addr  string
	init  bool
	mutex sync.RWMutex
}

var gLocalAddr = localAddr{init: false}

func GetLocalAddr() (rLocalAddr string, rErr error) {
	gLocalAddr.mutex.RLock()
	if gLocalAddr.init {
		rLocalAddr, rErr = gLocalAddr.addr, nil
		gLocalAddr.mutex.RUnlock()
		return
	}
	gLocalAddr.mutex.RUnlock()
	gLocalAddr.mutex.Lock()
	defer gLocalAddr.mutex.Unlock()
	if gLocalAddr.init {
		rLocalAddr, rErr = gLocalAddr.addr, nil
		return
	}
	addrs, rErr := net.InterfaceAddrs()
	if rErr != nil {
		return "", rErr
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				gLocalAddr.addr = ipnet.IP.String()
				gLocalAddr.init = true
				rLocalAddr, rErr = gLocalAddr.addr, nil
				return
			}
		}

	}
	return "", fmt.Errorf("Not Found Available")
}
