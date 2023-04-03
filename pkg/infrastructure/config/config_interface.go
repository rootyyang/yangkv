package config

import (
	"fmt"

	gomock "github.com/golang/mock/gomock"
	"github.com/spf13/pflag"
)

type ConfigInterface interface {
	ReadConfig(*pflag.FlagSet) error
	Stop()
	GetString(key string) string
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint16(key string) uint16
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetBool(key string) bool

	GetStringSlice(key string) []string
	GetIntSlice(key string) []int

	Set(key string, value interface{})
	SetDefault(key string, value interface{})
}

type allConfigAndUsedConfig struct {
	usedConfig        ConfigInterface
	configName2Config map[string]ConfigInterface
}

var gUseConfig = allConfigAndUsedConfig{nil, make(map[string]ConfigInterface)}

func SetConfig(pConfigName string) error {
	if config, ok := gUseConfig.configName2Config[pConfigName]; ok {
		gUseConfig.usedConfig = config
		return nil
	}
	return fmt.Errorf("config[%s] Not Found global[%v]", pConfigName, gUseConfig.configName2Config)
}

//该函数可以并发访问
func GetConfig() (ConfigInterface, error) {
	if gUseConfig.usedConfig == nil {
		return nil, fmt.Errorf("Please call SetConfig(pConfigName string) first")
	}
	return gUseConfig.usedConfig, nil

}

//该函数不可并发访问
func registerConfig(pConfigName string, pConfigInterface ConfigInterface) {
	gUseConfig.configName2Config[pConfigName] = pConfigInterface
}

func RegisterAndUseMockConfig(mockCtl *gomock.Controller) *MockConfigInterface {
	configInterface := NewMockConfigInterface(mockCtl)
	registerConfig("mockConfig", configInterface)
	SetConfig("mockConfig")
	return configInterface
}
