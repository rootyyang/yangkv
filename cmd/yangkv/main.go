package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/rootyyang/yangkv/pkg/control/role"
	"github.com/rootyyang/yangkv/pkg/domain/cluster"
	"github.com/rootyyang/yangkv/pkg/domain/node"
	"github.com/rootyyang/yangkv/pkg/infrastructure/config"
	"github.com/rootyyang/yangkv/pkg/infrastructure/log"
	"github.com/rootyyang/yangkv/pkg/infrastructure/rpc"
	"github.com/spf13/pflag"
)

func main() {

	err := assemblyComponents()
	if err != nil {
		fmt.Printf("assemblyComponents error %s", err.Error())
		os.Exit(1)
	}

	err = initConfig()
	if err != nil {
		fmt.Printf("initConfig error %s", err.Error())
		os.Exit(1)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("config.GetConfig() error %s", err.Error())
		os.Exit(1)
	}
	//设置使用的RPC
	//这里不将配置透传下去
	if is_master := cfg.GetBool("master"); is_master {
		masterRole := role.GetMasterRole()
		masterRole.Start(cfg.GetString("node.cluster_name"), cfg.GetStringSlice("node.discovery_list"), strconv.Itoa(cfg.GetInt("node.transport")))
		waitStop()
		masterRole.Stop()
	} else {
		dataRole := role.GetDataRole()
		dataRole.Start(cfg.GetString("node.cluster_name"), cfg.GetStringSlice("node.discovery_list"), strconv.Itoa(cfg.GetInt("node.transport")))
		waitStop()
		dataRole.Stop()
	}

}

//思考怎么把依赖关系划分好
func waitStop() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

/*默认的命令行参数读取使用pflag*/
func initConfig() error {
	ci, err := config.GetConfig()
	if err != nil {
		return err
	}
	pflag.StringP("config", "c", "../config/yangkv", "config path")
	pflag.Parse()
	ci.SetDefault("node.is_master", true)
	ci.SetDefault("node.discovery_list", []string{"127.0.0.1:3888"})
	ci.SetDefault("node.transport", 3888)
	ci.SetDefault("node.cluster_name", "yangkv")
	return ci.ReadConfig(pflag.CommandLine)
}

func assemblyComponents() error {
	err := node.SetNodeProvider("rpcnode")
	if err != nil {
		return err
	}
	err = cluster.SetNodeManager("default")
	if err != nil {
		return err
	}
	err = config.SetConfig("configViper")
	if err != nil {
		return err
	}
	err = log.SetUseLog("zaplumberjack")
	if err != nil {
		return err
	}
	err = rpc.SetUseRPC("grpc")
	if err != nil {
		return err
	}
	return nil
}
