package conf

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

var (
	configOnce sync.Once
	config     *viper.Viper
)

func GetConfig() *viper.Viper {
	configOnce.Do(func() {
		config = viper.New()
		config.SetConfigFile("config.json")
		err := config.ReadInConfig()
		if err != nil {
			log.Fatalf("Lỗi khi đọc tệp cấu hình: %v", err)
		}
	})
	return config
}
