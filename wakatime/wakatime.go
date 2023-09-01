package wakatime

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/wakatime/wakatime-cli/cmd/heartbeat"
)

func Execute(lang string, key string){
  viper :=  viper.New()
  home_dir, err := os.UserHomeDir()
  if err!=nil {
    fmt.Println("Error getting homedir")
    os.Exit(1)
  }
  viper.Set("entity", fmt.Sprintf("%s%s%s%s", home_dir, string(os.PathSeparator), "init.", lang))
  viper.Set("plugin", fmt.Sprintf("wakarizer/0.1"))
  viper.Set("key", key)
  viper.Set("write", "")

  heartbeat.Run(viper)
}
