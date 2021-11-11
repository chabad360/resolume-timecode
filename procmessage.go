package main

import (
	"fmt"
	"github.com/chabad360/go-osc/osc"
	"strings"
	"time"
)

var (
	clipName         = ""
	directionForward = true

	timeLeft   string
	clipLength string

	timePrev           = time.Now()
	posPrev            float32
	samples            int
	posIntervalBuffer  = []float32{0}
	timeIntervalBuffer = []float32{0}
	estSizeBuffer      = []float32{0}
)

//reset();

func maxAppend(array []float32, value float32, limit int) []float32 {
	array = append(array, value)
	if len(array) > limit {
		array = array[1:]
	}
	return array
}

func average(array []float32) float32 {
	var f float32
	for i := 0; i < len(array); i++ {
		f += array[i]
	}
	return f / float32(len(array))
}

func isWithin(original float32, newNum float32, percent float32) bool {
	p := (original / 100) * percent
	return !((newNum > original+p || newNum < original-p) && original != 0)
}

func procMsg(data *osc.Message) {
	if strings.Contains(data.Address, clipPath) {
		if strings.HasSuffix(data.Address, "/position") {
			procPos(data)
		} else if strings.HasSuffix(data.Address, "direction") {
			procDirection(data)
		} else if strings.HasSuffix(data.Address, "/name") {
			procName(data)
		} else if strings.Contains(data.Address, "/connect") {
			reset()
		}
	}
}

func procDirection(data *osc.Message) {
	directionForward = data.Arguments[0].(int32) != 0
	reset()
}

func procName(data *osc.Message) {
	clipName = data.Arguments[0].(string)
	clipNameBinding.Set("Clip Name: " + clipName)
	broadcast.Publish([]byte(fmt.Sprintf("/name ,s %s", clipName)))
}

func reset() {
	message.Address = fmt.Sprintf("%s/name", clipPath)
	b.Reset()
	message.LightMarshalBinary(b)
	client.Write(b.Bytes())

	samples = 0
	posPrev = 0
	posIntervalBuffer = []float32{0}
	timeIntervalBuffer = []float32{0}
	estSizeBuffer = []float32{0}

}

func procPos(data *osc.Message) {
	timeNow := time.Now()

	pos := data.Arguments[0].(float32)

	if !directionForward {
		pos = 1 - pos
	}

	if ((average(estSizeBuffer)*pos)/1000) < 10 && posPrev != 0 {
		posPrev = 0
	}

	currentPosInterval := pos - posPrev
	currentTimeInterval := float32(timeNow.Sub(timePrev).Microseconds())

	if currentPosInterval == 0 || currentTimeInterval == 0 {
		return
	}

	if currentPosInterval < 0 && posPrev > 0 {
		return
	}

	posIntervalBuffer = maxAppend(posIntervalBuffer, currentPosInterval, 100)
	timeIntervalBuffer = maxAppend(timeIntervalBuffer, currentTimeInterval, 100)

	posInterval := average(posIntervalBuffer)
	timeInterval := average(timeIntervalBuffer)

	currentEstSize := timeInterval * (1 / posInterval)
	prevEstSize := average(estSizeBuffer)
	if samples > 1000 && samples < 1500 && isWithin(prevEstSize, currentEstSize, 0.001) {
		estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 500)
	} else if samples > 500 && samples < 1000 && isWithin(prevEstSize, currentEstSize, 1) {
		estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 250)
	} else if samples < 500 {
		estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 100)
	}

	samples++

	posPrev = pos
	timePrev = timeNow

	t := (average(estSizeBuffer) * (1 - pos)) / 1000

	timeActual := time.UnixMilli(int64(t)).UTC()
	timeLeft = fmt.Sprintf("-%02d:%02d:%02d.%03d", timeActual.Hour(), timeActual.Minute(), timeActual.Second(), timeActual.Nanosecond()/1000000)
	clipLength = fmt.Sprintf("%.3fs", average(estSizeBuffer)/1000000)
	message := fmt.Sprintf("/time ,ss %s %s", timeLeft, clipLength)
	broadcast.Publish([]byte(message))

	//fmt.Println(message, clipLength, samples, pos, currentPosInterval, currentTimeInterval, currentEstSize, posInterval, timeInterval, average(estSizeBuffer))

}
