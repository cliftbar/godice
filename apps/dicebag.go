package apps

import (
	"fmt"
	"godice/config"
	pix "godice/pixel"
	"time"
	"tinygo.org/x/bluetooth"
)

type DiceBag struct {
	Dice     map[uint32]*pix.Die
	LastRoll time.Time
}

func bagRollCallback(d *pix.Die) {
	fmt.Printf("bag roll callback %d\n", d.PixelId)
}

func (db *DiceBag) AddDie(die *pix.Die) {
	db.Dice[die.PixelId] = die
}

func (db *DiceBag) BlockForRolls(timeout time.Duration) []*pix.Die {
	end := time.Now().Add(timeout)
	for time.Now().Before(end) && db.isRolling() {
		time.Sleep(100 * time.Millisecond)
	}

	var rolled []*pix.Die
	for _, v := range db.Dice {
		if v.LastRolledState.After(db.LastRoll) {
			rolled = append(rolled, v)
		}
	}

	db.LastRoll = time.Now()

	return rolled
}

func (db *DiceBag) isRolling() bool {
	for _, v := range db.Dice {
		if v.RollState == pix.RollStateRolling {
			return true
		}
	}

	return false
}

func DiceBagRunner(conf config.AppConfig) {
	adapter := bluetooth.DefaultAdapter
	_ = adapter.Enable()
	dieChan := make(chan *pix.Die)

	bag := DiceBag{make(map[uint32]*pix.Die), time.Now()}

	go pix.WatchForDice(adapter, dieChan, func(d *pix.Die) {
		rolls := bag.BlockForRolls(5 * time.Second)
		fmt.Printf("bag roll callback %d, %d\n", d.PixelId, len(rolls))
	})

	go storeDieInBag(dieChan, &bag)

	for {
		//fmt.Printf("dice: %v\n", bag.Dice)

		time.Sleep(1000 * time.Millisecond)
	}
}

func storeDieInBag(dieReadChannel chan *pix.Die, bag *DiceBag) {
	for {
		newDie := <-dieReadChannel
		bag.AddDie(newDie)
	}
}
