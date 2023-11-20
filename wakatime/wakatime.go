package wakatime

import (
	"fmt"
	"math/rand"
	"os"

	"time"

	"github.com/spf13/viper"
	"github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"gopkg.in/ini.v1"
)

func Execute(lang string, key string) {
	iniOption := viper.IniLoadOptions(ini.LoadOptions{
		AllowPythonMultilineValues: true,
	})
	viper := viper.NewWithOptions(iniOption)

	file, err := os.CreateTemp(fmt.Sprintf("%s", os.TempDir()), fmt.Sprintf("*.%s", lang))
	if err != nil {
		fmt.Printf("Error creating temp file: %s", err)
		os.Exit(1)
	}

	fileName := file.Name()

	viper.Set("entity", fileName)
	viper.Set("plugin", "\"wakarizer-wakatime/0.1\"")
	viper.Set("write", "")
	viper.Set("key", key)

	file.WriteString(fmt.Sprintf("YOLO BABY %d", rand.Intn(999999999)))
	file.Close()

	_, err = heartbeat.Run(viper)
	if err != nil {
		//fmt.Printf("ERROR: %s",err) uncomment to see debug info
	}

	os.Remove(file.Name())

	time.Sleep(time.Second * 3)
	//fmt.Printf("KEY: %s FILE: %s",key, file) //print where the file is located
	// Comment out bellow for testing execution
	//os.Exit(1)
}
