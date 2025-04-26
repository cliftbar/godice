package homeassistiant

import (
	"fmt"
	"image/color"
)

func (haClient *HAClient) HaColorLight(color color.RGBA, entityId string, action string) {
	data := map[string]interface{}{
		"rgb_color": []int{int(color.R), int(color.G), int(color.B)},
		"entity_id": entityId,
	}
	// TODO fix the returns
	_, err := haClient.CallService("light", action, data, false)
	if err != nil {
		fmt.Println(err)
	}
}

func (haClient *HAClient) HaOffLight(entityId string) {
	data := map[string]interface{}{
		"entity_id": entityId,
	}
	// TODO fix the returns
	_, err := haClient.CallService("light", "turn_off", data, false)
	if err != nil {
		fmt.Println(err)
	}
}

func (haClient *HAClient) HaTempLight(temperature int, entityId string, action string) {
	data := map[string]interface{}{
		"color_temp": temperature,
		"entity_id":  entityId,
	}
	// TODO fix the returns
	_, err := haClient.CallService("light", action, data, false)
	if err != nil {
		fmt.Println(err)
	}
}
