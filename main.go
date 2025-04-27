package main

import (
	"fmt"
	"godice/config"
	ha "godice/homeassistiant"
	"godice/pixel"
	cn "golang.org/x/image/colornames"
	"image/color"
	"time"
	"tinygo.org/x/bluetooth"
)

var conf, confErr = config.LoadConfig("config.yaml")

func main() {

	if confErr != nil {
		panic(confErr)
	}
	MainPixelSdk(conf.HAConfig.URL, conf.HAConfig.Token)
}

func MainPixelSdk(haUrl string, haToken string) {
	adapter := bluetooth.DefaultAdapter
	haClient := ha.NewClient(haUrl, haToken)

	die := &pixel.Die{}
	must("enable BLE stack", adapter.Enable())
	must("connect", die.Connect(adapter))
	must("who are you", die.SendWhoAreYou())
	//red := color.RGBA{255, 0, 0, 255}
	time.Sleep(3 * time.Second)
	red := cn.Purple
	must("blink", die.SendMsg(pixel.MessageBlink{
		Count:     3,
		Duration:  1000,
		Color:     red,
		FaceMask:  0xFFFFFF,
		Fade:      128,
		LoopCount: 0,
	}))
	go dieWatcher(die, haClient)
	select {}
}

func dieWatcher(die *pixel.Die, haClient *ha.HAClient) {
	target := conf.HAConfig.LightEntities[0]

	haClient.LightOff(target)
	time.Sleep(500 * time.Millisecond)
	haClient.LightTemperature(target, 2500)

	lastUpdated := die.LastUpdated
	for true {
		if die.LastUpdated.After(lastUpdated) {
			fmt.Printf("Roll: %d\n", die.CurrentFaceValue)
			if die.CurrentFaceValue == 20 {
				haClient.LightCycleColorsEz(target, []color.RGBA{
					cn.Red,
					cn.Orange,
					cn.Yellow,
					cn.Green,
					cn.Blue,
					cn.Indigo,
					cn.Purple,
				})
			} else if die.CurrentFaceValue >= 15 {
				haClient.LightColor(target, cn.Royalblue)
			} else if die.CurrentFaceValue >= 5 {
				haClient.LightColor(target, cn.Green)
			} else if die.CurrentFaceValue > 1 {
				haClient.LightColor(target, cn.Red)
			} else {
				haClient.LightCycleColors(target, []color.RGBA{cn.Red, cn.Red, cn.Red}, 500*time.Millisecond, true)
			}
			lastUpdated = die.LastUpdated
		}
		haClient.LightTemperature(target, 2500)
		time.Sleep(5 * time.Second)
	}

}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
