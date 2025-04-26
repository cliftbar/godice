package pixel

import (
	"encoding/binary"
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
	LastUpdated      time.Time
}

type IAmADieMessage struct {
	Id               uint8
	LedCount         uint8
	DesignAndColor   uint8
	Reserved         uint8
	DataSetHash      uint32
	PixelId          uint32
	AvailableFlash   uint16
	BuildTimestamp   uint32
	RollState        uint8
	CurrentFaceIndex uint8
	CurrentFaceValue uint8
	BatteryLevel     uint8
	BatteryState     uint8
}

type RollStateMessage struct {
	Id               uint8
	RollState        uint8
	CurrentFaceIndex uint8
	CurrentFaceValue uint8
}

type BatteryLevelMessage struct {
	Id           uint8
	BatteryLevel uint8
	BatteryState uint8
}

type BlinkMessage struct {
	Count     uint8
	Duration  uint16
	Color     color.RGBA
	FaceMask  uint32
	Fade      uint8
	LoopCount uint8
}

func rgbaToUint32(c color.RGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}

func (p *Die) Connect(adapter *bluetooth.Adapter) error {
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
	p.device = device

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
				p.notifyChar = char
				err = p.notifyChar.EnableNotifications(p.PixelCharacteristicReceiver)
			} else if char.UUID().String() == PixelWriteCharacteristic {
				p.writeChar = char
			}
		}
	}

	if err != nil {
		return fmt.Errorf("notification failed: %v", err)
	}
	return nil
}

func (p *Die) PixelCharacteristicReceiver(buf []byte) {
	if len(buf) == 0 {
		return
	}

	switch buf[0] {
	case MsgTypeIAmADie:
		msg := parseIAmADieMessage(buf)
		log.Printf("Received IAmADie: %+v", msg)
	case MsgTypeRollState:
		msg := parseRollStateMessage(buf)
		log.Printf("Received RollState: %+v", msg)
		if msg.RollState == RollStateOnFace || msg.RollState == RollStateRolled {
			p.CurrentFaceIndex = msg.CurrentFaceIndex
			p.CurrentFaceValue = msg.CurrentFaceValue
			p.LastUpdated = time.Now()
		}
	case MsgTypeBlinkAck:
		log.Printf("Blink Ack: %x", buf)
	case MsgTypeBatteryLevel:
		msg := parseBatteryLevelMessage(buf)
		log.Printf("Received BatteryLevel: %+v", msg)
	default:

		log.Printf("received %d: %x", buf[0], buf)

	}

}

func parseIAmADieMessage(buf []byte) IAmADieMessage {
	msg := IAmADieMessage{
		Id:               buf[0],
		LedCount:         buf[1],
		DesignAndColor:   buf[2],
		Reserved:         buf[3],
		DataSetHash:      binary.LittleEndian.Uint32(buf[4:]),
		PixelId:          binary.LittleEndian.Uint32(buf[8:]),
		AvailableFlash:   binary.LittleEndian.Uint16(buf[12:]),
		BuildTimestamp:   binary.LittleEndian.Uint32(buf[14:]),
		RollState:        buf[18],
		CurrentFaceIndex: buf[19],
		CurrentFaceValue: buf[19] + 1,
		BatteryLevel:     buf[20],
		BatteryState:     buf[21],
	}
	return msg
}

func parseRollStateMessage(buf []byte) RollStateMessage {
	msg := RollStateMessage{
		Id:               buf[0],
		RollState:        buf[1],
		CurrentFaceIndex: buf[2],
		CurrentFaceValue: buf[2] + 1,
	}
	return msg
}

func parseBatteryLevelMessage(buf []byte) BatteryLevelMessage {
	msg := BatteryLevelMessage{
		Id:           buf[0],
		BatteryLevel: buf[1],
		BatteryState: buf[2],
	}
	return msg
}

func (p *Die) SendWhoAreYou() error {
	msg := []byte{MsgTypeWhoAreYou}
	val, err := p.writeChar.WriteWithoutResponse(msg)
	println("who are you: ", val)
	return err
}

func (p *Die) SendBlink(blink BlinkMessage) error {
	msg := make([]byte, 14)
	msg[0] = MsgTypeBlink
	msg[1] = blink.Count
	binary.LittleEndian.PutUint16(msg[2:], blink.Duration)
	msg[4] = blink.Color.B
	msg[5] = blink.Color.G
	msg[6] = blink.Color.R
	msg[7] = blink.Color.A
	binary.LittleEndian.PutUint32(msg[8:], blink.FaceMask)
	msg[12] = blink.Fade
	msg[13] = blink.LoopCount

	_, err := p.writeChar.WriteWithoutResponse(msg)
	return err
}
