package apps

import (
	"context"
	"fmt"
	"godice/config"
	pix "godice/pixel"
	"time"
	"tinygo.org/x/bluetooth"
)

type RollGate struct {
	elements map[uint32]*pix.Die
}

func (q *RollGate) load(lc <-chan *pix.Die) {
	for die := range lc {
		q.elements[die.PixelId] = die
	}
}
func (q *RollGate) drain(dc chan<- *pix.Die, timeout time.Duration) {
	end := time.Now().Add(timeout)
	for len(q.elements) > 0 && time.Now().Before(end) {
		out := q.elements[0]
		delete(q.elements, out.PixelId)
		dc <- out
	}
}

func rollCallback(d *pix.Die) {
	fmt.Printf("roll callback %d\n", d.PixelId)
}

func StateQueueRunner(conf config.AppConfig) {
	adapter := bluetooth.DefaultAdapter
	_ = adapter.Enable()
	dieChan := make(chan *pix.Die)
	go pix.WatchForDice(adapter, dieChan, rollCallback)

	allDie := make(map[uint32]*pix.Die)
	go storeDie(dieChan, &allDie)

	for {
		fmt.Printf("dice: %v\n", allDie)
		for {
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

func StateQueueRunner2(conf config.AppConfig) {
	adapter := bluetooth.DefaultAdapter
	_ = adapter.Enable()
	dieChan := make(chan *pix.Die)
	go pix.WatchForDice(adapter, dieChan, rollCallback)

	allDie := make(map[uint32]*pix.Die)
	go storeDie(dieChan, &allDie)

	// poll for roll updates
	for {
		if rollHappened(&allDie) {
			// interrupt roll scanning
			// start pushing rolling die into Gate
			ctx, cancel := context.WithCancel(context.Background())

			rolling := make(map[uint32]*pix.Die)
			go func(ctx *context.Context) {

				for {
					select {
					case <-(*ctx).Done():
						return
					default:
						for k, v := range allDie {
							if v.RollState == pix.RollStateRolling {
								rolling[k] = v
							}
						}
					}
				}
			}(&ctx)
			// start polling for the first die to finish rolling

			rollTimeout := 5 * time.Second
			end := time.Now().Add(rollTimeout)
			for time.Now().Before(end) && rollEnding(&allDie) {
				fmt.Printf("Dice rolling: %v\n", rolling)
				time.Sleep(10 * time.Millisecond)
			}
			cancel()
			endSettle := time.Now().Add(rollTimeout)
			for time.Now().Before(endSettle) {
				for _, v := range rolling {
					if v.RollState != pix.RollStateRolled {
						time.Sleep(10 * time.Millisecond)
						continue
					}
				}
			}

			fmt.Printf("Dice rolled: %v\n", rolling)

		}

		time.Sleep(500 * time.Millisecond)
	}

}

func rollHappened(watched *map[uint32]*pix.Die) bool {
	for _, v := range *watched {
		if v.RollState == pix.RollStateRolling {
			return true
		}
	}

	return false
}

func rollEnding(watched *map[uint32]*pix.Die) bool {
	for _, v := range *watched {
		if v.RollState == pix.RollStateRolled {
			return true
		}
	}

	return false
}

func storeDie(dieReadChannel chan *pix.Die, allDie *map[uint32]*pix.Die) {
	for {
		newDie := <-dieReadChannel
		(*allDie)[newDie.PixelId] = newDie
	}
}
