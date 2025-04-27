package homeassistiant

import (
	"fmt"
	"image/color"
	"time"
)

func (haClient *HAClient) LightColor(entityId string, color color.RGBA) {
	data := map[string]interface{}{
		"rgb_color": []int{int(color.R), int(color.G), int(color.B)},
		"entity_id": entityId,
	}
	// TODO fix the returns
	_, err := haClient.CallService("light", "turn_on", data, false)
	if err != nil {
		fmt.Println(err)
	}
}

func (haClient *HAClient) LightOff(entityId string) {
	data := map[string]interface{}{
		"entity_id": entityId,
	}
	// TODO fix the returns
	_, err := haClient.CallService("light", "turn_off", data, false)
	if err != nil {
		fmt.Println(err)
	}
}

func (haClient *HAClient) LightTemperature(entityId string, temperature int) {
	data := map[string]interface{}{
		"entity_id":  entityId,
		"color_temp": temperature,
	}
	// TODO fix the returns
	_, err := haClient.CallService("light", "turn_on", data, false)
	if err != nil {
		fmt.Println(err)
	}
}

func (haClient *HAClient) LightCycleColors(entityId string, colors []color.RGBA, interval time.Duration, blink bool) {
	for _, c := range colors[:len(colors)-1] {
		haClient.LightColor(entityId, c)
		if blink {
			time.Sleep(interval)
			haClient.LightOff(entityId)
		}
		time.Sleep(interval)
	}
	haClient.LightColor(entityId, colors[len(colors)-1])
}

func (haClient *HAClient) LightCycleColorsEz(entityId string, colors []color.RGBA) {
	haClient.LightCycleColors(entityId, colors, 500*time.Millisecond, true)
}
