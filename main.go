package main

import (
	"fmt"
	"godice/config"
	ha "godice/homeassistiant"
	pix "godice/pixel"
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
	MultiDieRunner()
	//SingleDiePixelRunner(conf.HAConfig.URL, conf.HAConfig.Token)
}

func MultiDieRunner() {
	adapter := bluetooth.DefaultAdapter
	_ = adapter.Enable()
	dieChan := make(chan *pix.Die)
	go pix.WatchForDice(adapter, dieChan)

	allDie := make(map[uint32]*pix.Die)
	go func(dch chan *pix.Die) {
		for {
			newDie := <-dch
			allDie[newDie.PixelId] = newDie
			//fmt.Printf("new die %d on face %d\n", newDie.PixelId, newDie.CurrentFaceValue)
		}
	}(dieChan)

	multipleDiceWatcher(&allDie, ha.NewClient(conf.HAConfig.URL, conf.HAConfig.Token))
}

func getUpdatedDie(dice *map[uint32]*pix.Die, asOf *map[uint32]time.Time) (updated map[uint32]*pix.Die) {
	updated = make(map[uint32]*pix.Die)
	for id, die := range *dice {
		//fmt.Printf("LastRolled %s\n", die.LastRolled)
		val, exists := (*asOf)[id]
		if !exists {
			(*asOf)[id] = time.Now()
		}
		if exists && die.LastRolled.After(val) {
			updated[id] = die
		}
	}
	return updated
}

func lastUpdate(dice *map[uint32]*pix.Die) (latest time.Time) {
	for _, die := range *dice {
		if die.LastRolled.After(latest) {
			latest = die.LastRolled
		}
	}
	return latest
}

func rollTotal(dice *map[uint32]*pix.Die) (total int) {
	for _, die := range *dice {
		total += int(die.CurrentFaceValue)
	}
	return total
}

func multipleDiceWatcher(dice *map[uint32]*pix.Die, haClient *ha.HAClient) {
	target := conf.HAConfig.LightEntities[0]

	haClient.LightOff(target)
	time.Sleep(500 * time.Millisecond)
	haClient.LightTemperature(target, 2500)

	lastRollMap := make(map[uint32]time.Time)
	//lastRollChecked := lastUpdate(dice)
	for {
		updatedDice := getUpdatedDie(dice, &lastRollMap)
		if len(updatedDice) == 0 {
			//fmt.Printf("no dice updated as of %s\n", lastRollChecked)
			time.Sleep(2 * time.Second)
			continue
		}

		rollTotal := rollTotal(&updatedDice)
		fmt.Printf("Roll Total: %d\n", rollTotal)
		if rollTotal == 20 {
			haClient.LightCycleColorsEz(target, []color.RGBA{
				cn.Red,
				cn.Orange,
				cn.Yellow,
				cn.Green,
				cn.Blue,
				cn.Indigo,
				cn.Purple,
			})
		} else if rollTotal >= 15 {
			haClient.LightColor(target, cn.Royalblue)
		} else if rollTotal >= 10 {
			haClient.LightColor(target, cn.Green)
		} else if rollTotal >= 5 {
			haClient.LightColor(target, cn.Orange)
		} else if rollTotal > 1 {
			haClient.LightColor(target, cn.Red)
		} else {
			haClient.LightCycleColors(target, []color.RGBA{cn.Red, cn.Red, cn.Red}, 500*time.Millisecond, true)
		}

		for id := range updatedDice {
			lastRollMap[id] = time.Now()
		}
		time.Sleep(500 * time.Millisecond)
		haClient.LightTemperature(target, 2500)
		time.Sleep(5 * time.Second)
	}

}

func SingleDiePixelRunner(haUrl string, haToken string) {
	adapter := bluetooth.DefaultAdapter
	haClient := ha.NewClient(haUrl, haToken)

	die := &pix.Die{}
	must("enable BLE stack", adapter.Enable())
	must("connect", die.Connect(adapter))
	must("who are you", die.SendMsg(pix.MessageWhoAreYou{}))
	time.Sleep(3 * time.Second)
	red := cn.Purple
	must("blink", die.SendMsg(pix.MessageBlink{
		Count:     3,
		Duration:  1000,
		Color:     red,
		FaceMask:  0xFFFFFF,
		Fade:      128,
		LoopCount: 0,
	}))
	go singleDieWatcher(die, haClient)
	select {}
}

func singleDieWatcher(die *pix.Die, haClient *ha.HAClient) {
	target := conf.HAConfig.LightEntities[0]

	haClient.LightOff(target)
	time.Sleep(500 * time.Millisecond)
	haClient.LightTemperature(target, 2500)

	lastUpdated := die.LastRolled
	for {
		if die.LastRolled.After(lastUpdated) {
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
			} else if die.CurrentFaceValue >= 10 {
				haClient.LightColor(target, cn.Green)
			} else if die.CurrentFaceValue >= 5 {
				haClient.LightColor(target, cn.Orange)
			} else if die.CurrentFaceValue > 1 {
				haClient.LightColor(target, cn.Red)
			} else {
				haClient.LightCycleColors(target, []color.RGBA{cn.Red, cn.Red, cn.Red}, 500*time.Millisecond, true)
			}
			lastUpdated = die.LastRolled
		}
		time.Sleep(500 * time.Millisecond)
		haClient.LightTemperature(target, 2500)
		time.Sleep(5 * time.Second)

	}

}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
