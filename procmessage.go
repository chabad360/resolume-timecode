package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	clipLength = "0.000s"
)

const multiplier = 100000000000 // This constant is used to avoid the mess that is floating point numbers.

var (
	clipName         = ""
	directionForward = true

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

func procMsg(data string) {
	if strings.Contains(data, clipPath) {
		if strings.Contains(data, "/position ,f ") {
			procPos(data)
		} else if strings.Contains(data, "direction ,i ") {
			procDirection(data)
		} else if strings.Contains(data, "/name ,s ") {
			procName(data)
		} else if strings.Contains(data, "/connect") {
			reset()
		}
	}
}

func procDirection(data string) {
	directionForward = data[len(data)-1] != []byte("0")[0]
	reset()
}

func procName(data string) {
	data = strings.Split(data, ",s ")[1]
	if data != clipName {
		clipName = data
	}
}

func reset() {
	message.Address = fmt.Sprintf("%s/name", clipPath)
	b.Reset()
	message.LightMarshalBinary(b)
	client.Write(b.Bytes())

	fmt.Println(samples, posPrev, posIntervalBuffer, timeIntervalBuffer, estSizeBuffer)

	samples = 0
	posPrev = 0
	posIntervalBuffer = []float32{0}
	timeIntervalBuffer = []float32{0}
	estSizeBuffer = []float32{0}

	fmt.Println(samples, posPrev, posIntervalBuffer, timeIntervalBuffer, estSizeBuffer)

}

func procPos(data string) {
	timeNow := time.Now()

	p, _ := strconv.ParseFloat(strings.Split(data, " ")[2], 32)
	pos := float32(p)

	if !directionForward {
		pos = 1 - pos
	}

	if ((average(estSizeBuffer)*pos)/1000) < 70 && posPrev != 0 {
		fmt.Println(average(estSizeBuffer), (average(estSizeBuffer) * pos))
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
	message := fmt.Sprintf("/time ,s %02d:%02d:%02d:%03d", timeActual.Hour(), timeActual.Minute(), timeActual.Second(), timeActual.Nanosecond()/1000000)
	clipLength = fmt.Sprintf("/length, %fs", average(estSizeBuffer)/1000000)
	//broadcast.Publish([]byte(message))
	//broadcast.Publish([]byte(clipLength))
	fmt.Println(message, clipLength, samples, pos, currentPosInterval, currentTimeInterval, currentEstSize, posInterval, timeInterval, average(estSizeBuffer))

}
