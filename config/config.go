package config

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.BindPFlags(pflag.CommandLine)
}

func Cookies() []string {
	return []string{}
}
