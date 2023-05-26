package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	registerConfig("configViper", new(configViper))
}

type configViper struct {
}

func (config *configViper) ReadConfig(pflagSet *pflag.FlagSet) error {
	viper.BindPFlags(pflagSet)
	viper.SetConfigFile(viper.GetString("config"))
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		//同一层之间的库尽量不相互调用，
		/*if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.GetLogInstance().Errorf("config[%s] not found", config.GetString("config"))
		} else {
			log.GetLogInstance().Errorf("read config[%s] error[%s]", config.GetString("config"), err.Error())
		}*/
		return err
	}
	return nil
}
func (config *configViper) Stop() {

}
func (config *configViper) GetString(key string) string {
	return viper.GetString(key)
}
func (config *configViper) GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}
func (config *configViper) GetInt(key string) int {
	return viper.GetInt(key)
}
func (config *configViper) GetInt32(key string) int32 {
	return viper.GetInt32(key)
}
func (config *configViper) GetInt64(key string) int64 {
	return viper.GetInt64(key)
}
func (config *configViper) GetUint(key string) uint {
	return viper.GetUint(key)

}
func (config *configViper) GetUint16(key string) uint16 {
	return viper.GetUint16(key)
}
func (config *configViper) GetUint32(key string) uint32 {
	return viper.GetUint32(key)
}
func (config *configViper) GetUint64(key string) uint64 {
	return viper.GetUint64(key)
}
func (config *configViper) Set(key string, value interface{}) {
	viper.Set(key, value)
}
func (config *configViper) SetDefault(key string, value interface{}) {
	viper.SetDefault(key, value)
}

func (config *configViper) GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func (config *configViper) GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}
func (config *configViper) GetBool(key string) bool {
	return viper.GetBool(key)
}
