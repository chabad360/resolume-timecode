package server

import (
	"fmt"
	"resolume-timecode/config"
	"resolume-timecode/services/clients"
	"resolume-timecode/util"
	"strings"
	"time"

	"github.com/chabad360/go-osc/osc"
)

var (
	clipName         = ""
	directionForward = true

	clipLength float32
	posPrev    float32

	message  = &osc.Message{Arguments: []interface{}{"?"}}
	message2 = &osc.Message{Arguments: []interface{}{"?"}}
)

func Reset() {
	clipPath := config.GetString(config.ClipPath)

	message.Address = clipPath + "/name"
	message2.Address = clipPath + "/transport/position/behaviour/duration"
	if _, err := oscServer.WriteTo(osc.NewBundle(message, message2), config.GetString(config.OSCAddr)+":"+config.GetString(config.OSCPort)); err != nil {
		fmt.Println(err)
	}

	posPrev = 0
}

func procMsg(data *osc.Message) {
	if strings.HasPrefix(data.Address, config.GetString(config.ClipPath)) {
		switch {
		case strings.HasSuffix(data.Address, "/position"):
			procPos(data)
		case strings.HasSuffix(data.Address, "direction"):
			procDirection(data)
		case strings.HasSuffix(data.Address, "/name"):
			procName(data)
		case strings.HasSuffix(data.Address, "/duration"):
			procDuration(data)
		case strings.HasSuffix(data.Address, "/connect"):
			Reset()
		case strings.Contains(data.Address, "/select"):
			Reset()
		}
	}
}

func procDirection(data *osc.Message) {
	directionForward = data.Arguments[0].(int32) != 0
	if !directionForward {
		posPrev = 1 - posPrev
	}
}

func procName(data *osc.Message) {
	clipName = data.Arguments[0].(string)
	clients.Publish(&util.Message{ClipName: clipName})
}

func procDuration(data *osc.Message) {
	clipLength = (data.Arguments[0].(float32) * 604800) + 0.001
	clients.Publish(&util.Message{ClipLength: fmt.Sprintf("%.3fs", clipLength)})
}

func procPos(data *osc.Message) {
	pos := data.Arguments[0].(float32)

	if !directionForward {
		pos = 1 - pos
	}

	if posPrev == 0 || posPrev == pos || (pos < posPrev && posPrev > 0.9 && pos < 0.1) {
		posPrev = pos
		return
	}

	currentPosInterval := pos - posPrev

	if currentPosInterval < 0 && posPrev > 0 {
		return
	}

	posPrev = pos

	if config.GetBool(config.ClipInvert) {
		pos = 1 - pos
	}

	t := (clipLength * 1000) * (1 - pos)

	timeActual := time.UnixMilli(int64(t)).UTC()

	clients.Publish(&util.Message{Hour: fmt.Sprintf("%02d", timeActual.Hour()),
		Minute:     fmt.Sprintf("%02d", timeActual.Minute()),
		Second:     fmt.Sprintf("%02d", timeActual.Second()),
		MS:         fmt.Sprintf("%03d", timeActual.Nanosecond()/1000000),
		ClipLength: fmt.Sprintf("%.3fs", clipLength),
		ClipName:   clipName,
		Message:    config.GetString(config.ClientMessage),
		Invert:     config.GetBool(config.ClipInvert),
		AlertTime:  config.GetInt(config.AlertTime),
	})

	//fmt.Println(message, clipLength, samples, pos, currentPosInterval, currentTimeInterval, currentEstSize, posInterval, timeInterval, average(estSizeBuffer))

}
