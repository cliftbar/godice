package pixel

import (
	"encoding/binary"
	"image/color"
	"time"
)

type MessageIAmADie struct {
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

func (die *Die) SendWhoAreYou() error {
	msg := []byte{MsgTypeWhoAreYou}
	val, err := die.writeChar.WriteWithoutResponse(msg)
	println("who are you: ", val)
	return err
}

func parseIAmADieMessage(buf []byte) MessageIAmADie {
	msg := MessageIAmADie{
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

func (die *Die) readIAmADieMsg(msg MessageIAmADie) {
	die.pixelId = msg.PixelId
	die.ledCount = msg.LedCount
	die.designAndColor = msg.DesignAndColor
	die.CurrentFaceIndex = msg.CurrentFaceIndex
	die.CurrentFaceValue = msg.CurrentFaceValue
	die.rollState = msg.RollState
	die.batteryLevel = msg.BatteryLevel
	die.buildTimestamp = msg.BuildTimestamp
	die.batteryCharging = msg.BatteryState == BattStateCharging
	die.LastUpdated = time.Now()
}

type MessageBlink struct {
	Count     uint8
	Duration  uint16
	Color     color.RGBA
	FaceMask  uint32
	Fade      uint8
	LoopCount uint8
}

func (msg MessageBlink) ToBuffer() (buf []byte) {
	buf = make([]byte, 14)
	buf[0] = MsgTypeBlink
	buf[1] = msg.Count
	binary.LittleEndian.PutUint16(buf[2:], msg.Duration)
	buf[4] = msg.Color.B
	buf[5] = msg.Color.G
	buf[6] = msg.Color.R
	buf[7] = msg.Color.A
	binary.LittleEndian.PutUint32(buf[8:], msg.FaceMask)
	buf[12] = msg.Fade
	buf[13] = msg.LoopCount

	return buf
}

type Message interface {
	ToBuffer() []byte
}

func (die *Die) SendBlink(blink MessageBlink) error {
	_, err := die.writeChar.WriteWithoutResponse(blink.ToBuffer())
	return err
}

func (die *Die) SendMsg(msg Message) error {
	_, err := die.writeChar.WriteWithoutResponse(msg.ToBuffer())
	return err
}

type MessageLightUpFace struct {
	LedIndex  uint8
	RemapFace uint8
	LedColor  color.RGBA
}
