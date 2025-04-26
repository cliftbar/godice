package main

import (
	"fmt"
	"godice/config"
	ha "godice/homeassistiant"
	"godice/pixel"
	"golang.org/x/image/colornames"
	"image/color"
	"time"
	"tinygo.org/x/bluetooth"
)

func main() {
	conf, err := config.LoadConfig("config.yaml")
	if err != nil {
		panic(err)
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
	red := color.RGBA{255, 0, 255, 255}
	must("blink", die.SendBlink(pixel.BlinkMessage{
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
	target := "light.blamp"

	//haClient.HaOffLight(target)
	//time.Sleep(500 * time.Millisecond)
	haClient.HaTempLight(2500, target, "turn_on")

	lastUpdated := die.LastUpdated
	for true {
		if die.LastUpdated.After(lastUpdated) {
			fmt.Printf("Roll: %d\n", die.CurrentFaceValue)
			if die.CurrentFaceValue == 20 {
				haClient.HaColorLight(colornames.Purple, target, "turn_on")
				time.Sleep(500 * time.Millisecond)
				haClient.HaColorLight(colornames.Green, target, "turn_on")
				time.Sleep(500 * time.Millisecond)
				haClient.HaColorLight(colornames.Royalblue, target, "turn_on")
			} else if die.CurrentFaceValue >= 15 {
				haClient.HaColorLight(colornames.Blue, target, "turn_on")
			} else if die.CurrentFaceValue >= 5 {
				haClient.HaColorLight(colornames.Green, target, "turn_on")
			} else if die.CurrentFaceValue > 1 {
				haClient.HaColorLight(colornames.Red, target, "turn_on")
			} else {
				haClient.HaColorLight(colornames.Red, target, "turn_on")
				time.Sleep(500 * time.Millisecond)
				haClient.HaOffLight(target)
				time.Sleep(500 * time.Millisecond)
				haClient.HaColorLight(colornames.Red, target, "turn_on")
			}
			lastUpdated = die.LastUpdated
		}
		time.Sleep(5 * time.Second)
	}

}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
