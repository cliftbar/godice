package apps

import (
	"fmt"
	"godice/config"
	ha "godice/homeassistiant"
	pix "godice/pixel"
	cn "golang.org/x/image/colornames"
	"image/color"
	"slices"
	"time"
	"tinygo.org/x/bluetooth"
)

func MultiDieRunner(conf config.AppConfig) {
	adapter := bluetooth.DefaultAdapter
	_ = adapter.Enable()
	dieChan := make(chan *pix.Die)
	go pix.WatchForDice(adapter, dieChan, func(d *pix.Die) { fmt.Printf("roll callback %d", d.PixelId) })

	allDie := make(map[uint32]*pix.Die)
	go func(dch chan *pix.Die) {
		for {
			newDie := <-dch
			allDie[newDie.PixelId] = newDie
		}
	}(dieChan)
	target := conf.HAConfig.LightEntities[0]
	multipleDiceWatcher(&allDie, ha.NewClient(conf.HAConfig.URL, conf.HAConfig.Token), target)
}

func getUpdatedDie(dice *map[uint32]*pix.Die, asOf *map[uint32]time.Time) (updated map[uint32]*pix.Die) {
	updated = make(map[uint32]*pix.Die)
	for id, die := range *dice {
		val, exists := (*asOf)[id]
		if !exists {
			(*asOf)[id] = time.Now()
		}
		if exists && die.RollState == pix.RollStateRolled && die.LastRolledState.After(val) {
			updated[id] = die
		}
	}
	return updated
}

func lastUpdate(dice *map[uint32]*pix.Die) (latest time.Time) {
	for _, die := range *dice {
		if die.LastRolledState.After(latest) {
			latest = die.LastRolledState
		}
	}
	return latest
}

func getRolls(dice *map[uint32]*pix.Die) (total int, rolls []int) {

	for _, die := range *dice {
		rolls = append(rolls, int(die.CurrentFaceValue))
		total += int(die.CurrentFaceValue)
	}

	fmt.Printf("Rolls: %v\n", rolls)
	return total, rolls
}

func multipleDiceWatcher(dice *map[uint32]*pix.Die, haClient *ha.HAClient, target string) {

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

		rollTotal, rolls := getRolls(&updatedDice)
		fmt.Printf("Roll Total: %d, Roll Adv: %d\n", rollTotal, slices.Max(rolls))
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
