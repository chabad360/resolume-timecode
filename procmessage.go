package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/chabad360/go-osc/osc"
)

var (
	clipName         = ""
	directionForward = true

	timeLeft string

	clipLength float32
	posPrev    float32
)

func procMsg(data *osc.Message) {
	if strings.Contains(data.Address, clipPath) {
		switch {
		case strings.HasSuffix(data.Address, "/position"):
			procPos(data)
		case strings.HasSuffix(data.Address, "direction"):
			procDirection(data)
		case strings.HasSuffix(data.Address, "/name"):
			procName(data)
		case strings.HasSuffix(data.Address, "/duration"):
			procDuration(data)
		case strings.Contains(data.Address, "/connect"):
			reset()
		case strings.Contains(data.Address, "/select"):
			reset()
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
	clipNameBinding.Set("Clip Name: " + clipName)
	broadcast.Publish(osc.NewMessage("/name", clipName))
}

func procDuration(data *osc.Message) {
	clipLength = (data.Arguments[0].(float32) * 604800) + 0.001
	clipLengthBinding.Set(fmt.Sprintf("Clip Length: %.3fs", clipLength))
	broadcast.Publish(osc.NewMessage("/duration", clipLength))
}

func reset() {
	lightReset()

	posPrev = 0
}

func lightReset() {
	message.Address = clipPath + "/name"
	message2.Address = clipPath + "/transport/position/behaviour/duration"
	oscServer.WriteTo(osc.NewBundle(message, message2), OSCAddr+":"+OSCPort)
}

func procPos(data *osc.Message) {
	pos := data.Arguments[0].(float32)

	if !directionForward {
		pos = 1 - pos
	}

	if posPrev == 0 || posPrev == pos || pos < 0.002 {
		posPrev = pos
		return
	}

	currentPosInterval := pos - posPrev

	if currentPosInterval < 0 && posPrev > 0 {
		return
	}

	posPrev = pos

	if clipInvert {
		pos = 1 - pos
	}

	t := (clipLength * 1000) * (1 - pos)

	timeActual := time.UnixMilli(int64(t)).UTC()

	timeLeft = fmt.Sprintf("-%02d:%02d:%02d.%03d", timeActual.Hour(), timeActual.Minute(), timeActual.Second(), timeActual.Nanosecond()/1000000)
	broadcast.Publish(osc.NewMessage("/time", timeLeft, fmt.Sprintf("%.3fs", clipLength)))
	broadcast.Send()

	//fmt.Println(message, clipLength, samples, pos, currentPosInterval, currentTimeInterval, currentEstSize, posInterval, timeInterval, average(estSizeBuffer))

}
