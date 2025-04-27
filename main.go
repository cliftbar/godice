package main

import (
	"godice/apps"
	"godice/config"
)

var conf, confErr = config.LoadConfig("config.yaml")

func main() {

	if confErr != nil {
		panic(confErr)
	}
	//apps.MultiDieRunner(*conf)
	apps.DiceBagRunner(*conf)
	//SingleDiePixelRunner(conf.HAConfig.URL, conf.HAConfig.Token)
}
