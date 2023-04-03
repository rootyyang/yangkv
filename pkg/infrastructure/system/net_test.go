package system

import (
	"strings"
	"testing"
)

func TestGetLocalAddr(t *testing.T) {
	addr, err := GetLocalAddr()
	if err != nil {
		t.Fatalf("GetLocalAddr() =[%s][%s] want ip,nil", addr, err.Error())
	}
	if len(addr) == 0 || addr == "127.0.0.1" {
		t.Fatalf("GetLocalAddr() =[%s]nil want ip,nil", addr)
	}

	sli := strings.Split(addr, ".")
	if len(sli) != 4 {
		t.Fatalf("GetLocalAddr() =[%s][%s] want ip,nil", addr, err)
	}
	t.Logf("IP:%s", addr)
	twiceAddr, err := GetLocalAddr()
	if err != nil {
		t.Fatalf("Second Call GetLocalAddr() =[%s][%s] want [%s],nil", twiceAddr, err.Error(), addr)
	}

	if addr != twiceAddr {
		t.Fatalf("GetLocalAddr() =[%s]nil want [%s],nil", twiceAddr, addr)
	}
}
