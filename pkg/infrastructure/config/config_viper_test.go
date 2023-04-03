package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestPFlag(t *testing.T) {
	pflagSet := pflag.NewFlagSet("test", pflag.ExitOnError)
	pflagSet.StringP("config", "c", "../config/yangkv", "config path")
	testData := [][]string{{"yangkv"}, {"yangkv", "-c", "./yangkv1.toml"}, {"yangkv", "-c", "./yangkv2.toml"}, {"yangkv", "-c", "./yangkv3.toml"}}
	testDataResult := []string{"../config/yangkv", "./yangkv1.toml", "./yangkv2.toml", "./yangkv3.toml"}
	for i, _ := range testData {
		os.Args = testData[i]
		pflagSet.Parse(os.Args)
		value, err := pflagSet.GetString("config")
		if err != nil {
			t.Fatalf("pflagSet.GetString(config)=[%v][%v], want=[%v]nil", value, err, testDataResult[i])
		}
		if value != testDataResult[i] {
			t.Fatalf("pflagSet.GetString(config)=[%v] nil, want=[%v]nil", value, testDataResult[i])
		}
	}
}

func TestViperToml(t *testing.T) {
	os.Args = []string{"yangkv", "-c", "./data_test.toml"}
	pflagSet := pflag.NewFlagSet("testViperToml", pflag.ExitOnError)
	pflagSet.StringP("config", "c", "../config/yangkv", "config path")
	pflagSet.Parse(os.Args)

	err := SetConfig("configViper")
	if err != nil {
		t.Fatalf("SetConfig(configViper)=[%v], want nil", err)
	}

	ci, err := GetConfig()
	if err != nil || ci == nil {
		t.Fatalf("GetConfig()=[%v][%v] want not nil,nil", ci, nil)
	}
	ci.SetDefault("node.is_master", true)
	ci.SetDefault("node.discovery_list", []string{"127.0.0.1:3888"})
	ci.SetDefault("node.transport", 3888)
	ci.SetDefault("node.cluster_name", "yangkv")

	ci.SetDefault("node2.is_master", true)
	ci.SetDefault("node2.discovery_list", []string{"127.0.0.1:3888"})
	ci.SetDefault("node2.transport", 3888)
	ci.SetDefault("node2.cluster_name", "yangkv")

	err = ci.ReadConfig(pflagSet)
	if err != nil {
		t.Fatalf("GetConfig()=[%v][%v] want not nil,nil", ci, err)
	}

	if ci.GetString("config") != "./data_test.toml" {
		t.Fatalf("GetString(config)=[%v] want ./data_test.toml", ci.GetString("config"))
	}

	if ci.GetBool("node.is_master") == true {
		t.Fatalf("GetBool(node.master)=false want true")
	}

	wantDiscovery := []string{"127.0.0.1:1888", "127.0.0.1:2888", "127.0.0.1:3888"}
	getDiscovery := ci.GetStringSlice("node.discovery_list")
	if len(getDiscovery) != len(wantDiscovery) {
		t.Fatalf("ci.GetStringSlice(node.discovery_list)=[%v] want {127.0.0.1:1888, 127.0.0.1:2888, 127.0.0.1:3888}", getDiscovery)
	}

	for i, _ := range getDiscovery {
		if getDiscovery[i] != wantDiscovery[i] {
			t.Fatalf("ci.GetStringSlice(node.discovery_list)=[%v] want {127.0.0.1:1888, 127.0.0.1:2888, 127.0.0.1:3888}", getDiscovery)
		}
	}

	if ci.GetInt("node.transport") != 4888 {
		t.Fatalf("GetInt(node.transport)=[%v] want 3888", ci.GetInt("node.transport"))
	}

	if ci.GetString("node.cluster_name") != "kvcluster" {
		t.Fatalf("GetString(node.cluster_name)=[%v] want true", ci.GetString("node.cluster_name"))
	}

	//测试默认值
	if ci.GetBool("node2.is_master") != true {
		t.Fatalf("GetBool(node2.master)=false want true")
	}

	wantDiscovery = []string{"127.0.0.1:3888"}
	getDiscovery = ci.GetStringSlice("node2.discovery_list")
	if len(getDiscovery) != len(wantDiscovery) {
		t.Fatalf("ci.GetStringSlice(node2.discovery_list)=[%v] want {127.0.0.1:3888}", getDiscovery)
	}

	for i, _ := range getDiscovery {
		if getDiscovery[i] != wantDiscovery[i] {
			t.Fatalf("ci.GetStringSlice(node2.discovery_list)=[%v] want {127.0.0.1:3888}", getDiscovery)
		}
	}

	if ci.GetInt("node2.transport") != 3888 {
		t.Fatalf("GetInt(node2.transport)=[%v] want 3888", ci.GetInt("node2.transport"))
	}

	if ci.GetString("node2.cluster_name") != "yangkv" {
		t.Fatalf("GetString(node2.cluster_name)=[%v] want yangkv", ci.GetString("node2.cluster_name"))
	}

}
