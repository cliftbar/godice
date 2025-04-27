package pixel

import (
	"fmt"
	"image/color"
	"log"
	"time"
	"tinygo.org/x/bluetooth"
)

type Die struct {
	device           bluetooth.Device
	writeChar        bluetooth.DeviceCharacteristic
	notifyChar       bluetooth.DeviceCharacteristic
	ledCount         uint8
	pixelId          uint32
	CurrentFaceIndex uint8
	CurrentFaceValue uint8
	rollState        uint8
	batteryLevel     uint8
	batteryCharging  bool
	buildTimestamp   uint32
	designAndColor   uint8
	LastUpdated      time.Time
}

func rgbaToUint32(c color.RGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}

func (die *Die) Connect(adapter *bluetooth.Adapter) error {
	pixelServiceUuid, _ := bluetooth.ParseUUID(PixelsService)
	notifyCharacterUuid, _ := bluetooth.ParseUUID(PixelNotifyCharacteristic)
	writeCharacteristicUUid, _ := bluetooth.ParseUUID(PixelWriteCharacteristic)

	ch := make(chan bluetooth.ScanResult, 1)
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {

		if device.HasServiceUUID(pixelServiceUuid) {
			_ = adapter.StopScan()
			ch <- device
		}
	})
	if err != nil {
		return fmt.Errorf("scan failed: %v", err)
	}

	result := <-ch
	device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	die.device = device

	services, err := device.DiscoverServices([]bluetooth.UUID{pixelServiceUuid})
	if err != nil {
		return fmt.Errorf("service discovery failed: %v", err)
	}

	for _, service := range services {
		if service.UUID().String() != PixelsService {
			continue
		}

		chars, _ := service.DiscoverCharacteristics([]bluetooth.UUID{notifyCharacterUuid, writeCharacteristicUUid})
		for _, char := range chars {
			if char.UUID().String() == PixelNotifyCharacteristic {
				die.notifyChar = char
				err = die.notifyChar.EnableNotifications(die.PixelCharacteristicReceiver)
			} else if char.UUID().String() == PixelWriteCharacteristic {
				die.writeChar = char
			}
		}
	}

	if err != nil {
		return fmt.Errorf("notification failed: %v", err)
	}
	return nil
}

func (die *Die) PixelCharacteristicReceiver(buf []byte) {
	if len(buf) == 0 {
		return
	}

	switch buf[0] {
	case MsgTypeIAmADie:
		msg := parseIAmADieMessage(buf)
		die.readIAmADieMsg(msg)

		log.Printf("Received IAmADie: %+v", msg)
	case MsgTypeRollState:
		msg := parseRollStateMessage(buf)
		log.Printf("Received RollState: %+v", msg)
		if msg.RollState == RollStateOnFace || msg.RollState == RollStateRolled {
			die.CurrentFaceIndex = msg.CurrentFaceIndex
			die.CurrentFaceValue = msg.CurrentFaceValue
			die.LastUpdated = time.Now()
		}
	case MsgTypeBlinkAck:
		log.Printf("Blink Ack: %x", buf)
	case MsgTypeBatteryLevel:
		msg := parseBatteryLevelMessage(buf)
		die.readBatteryBuffer(buf)
		log.Printf("Received BatteryLevel: %+v", msg)
	default:

		log.Printf("received %d: %x", buf[0], buf)

	}

}
