package main

import (
	"context"
	"fmt"
	"github.com/siongui/godom/wasm"
	"log"
	"nhooyr.io/websocket"
	"os"
	"strconv"
	"strings"
	"time"
)

func startOSC(div *wasm.Value, addr string) {
	timePrev = time.Now()

	ctx := context.Background()

	c, _, err := websocket.Dial(ctx, "ws://"+addr+"/ws", nil)
	if err != nil {
		log.Fatal(err)
	}
	//defer c.Close(websocket.StatusInternalError, "the sky is falling")
	defer c.Close(websocket.StatusNormalClosure, "")

	for {
		//packet, err := server.ReceivePacket(conn)
		_, packet, err := c.Read(ctx)
		timeNow = time.Now()
		if err != nil {
			//fmt.Println("Server error: " + err.Error())
			os.Exit(1)
		}
		msg := string(packet)
		fmt.Println(msg)
		switch {
		case strings.Contains(msg, "/composition/selectedclip/transport/position"):
			procMsg(div)(msg)
		case strings.Contains(msg, "/composition/selectedclip/name"):
			clipName(msg)
		}
	}
}

var (
	posPrev    float32
	timePrev   time.Time
	timeNow    time.Time
	tPrev      int
	samples    int
	interval   float32
	duration   float32 = 10000
	div        float32 = 1000
	array              = []float32{0}
	timeArray          = []float32{0}
	totalArray         = []float32{0}
	name       string
)

func Pop(array []float32) (values []float32, value float32) {
	value, values = array[0], array[1:]
	return
}

func maxAppend(array []float32, value float32, limit int) []float32 {
	array = append(array, value)
	if len(array) > limit {
		array, _ = Pop(array)
	}
	return array
}

func average(array []float32) float32 {
	var f float32
	for i := 0; i < len(array); i++ {
		f = f + array[i]
	}
	return f / float32(len(array))
}

func within(original float32, new float32, percent float32) bool {
	p := original / 100 * percent
	if (new > original+p || new < original-p) && original != 0 {
		return false
	}
	return true
}

func clipName(msg string) {
	nameString := strings.TrimPrefix(msg, "/composition/selectedclip/name ,s ")
	if name != nameString {
		name = nameString
		reset()
	}
}

func reset() {
	samples = 0
	totalArray = []float32{0}
	timeArray = []float32{0}
	array = []float32{0}
}

func procMsg(testdiv *wasm.Value) func(string) {
	return func(msg string) {
		var t int
		var timeActual time.Duration
		var tA float32
		var td float32
		posString := strings.TrimPrefix(msg, "/composition/selectedclip/transport/position ,f ")
		pos64, _ := strconv.ParseFloat(posString, 32)
		pos := float32(pos64)
		if pos < 0.000005 {
			reset()
		}

		//a := average(array)
		ta := average(timeArray)
		i := pos - posPrev
		d := float32(timeNow.Sub(timePrev).Microseconds())
		if i == 0 || d == 0 {
			goto done
		}
		if !within(ta, d, 50) && samples > 500 {
			//fmt.Printf("w")
		} else {
			timeArray = maxAppend(timeArray, d, 100)
			array = maxAppend(array, i, 100)
			interval = average(array)
			duration = average(timeArray) / div

			td = duration * (1 / interval)
			tA = average(totalArray)
			if within(tA, td, 0.001) && samples > 1000 && samples < 2000 {
				totalArray = maxAppend(totalArray, td, 500)
			} else if within(tA, td, 1) && samples > 500 {
				totalArray = maxAppend(totalArray, td, 250)
			} else if samples < 500 {
				totalArray = maxAppend(totalArray, td, 100)
			}
		}
		samples++

		//t = (duration/div)*((1-pos)/interval)
		//t = (duration/div)*((1-pos)/(interval/samples))
		//t = int((duration)*((1-pos)/interval))
		t = int((average(totalArray) / 1) * (1 - pos))
		//t = int(100000/1)*(1-pos))

		posPrev = pos
		timePrev = timeNow
		timeActual, _ = time.ParseDuration(fmt.Sprintf("%dms", int(t)))
		testdiv.Set("innerHTML", fmt.Sprintf("-%02d:%02d:%02d.%03d", int(timeActual.Hours()),
			int(timeActual.Minutes())-(60*int(timeActual.Hours())),
			int(timeActual.Seconds())-(60*int(timeActual.Minutes())),
			int(timeActual.Milliseconds()-(1000*int64(timeActual.Seconds())))))
		//fmt.Printf("pos: %f\taverage: %f\ti: %f\ttime: %f\ttimeAverage: %f\ttimeActual: %d\ttimeTotal: %d\n", pos, a, interval, d, ta, t, int(average(totalArray)))
		//fmt.Printf("%f,%f,%f,%f,%f,%d,%d,%d\n", pos, a, interval, d, ta, t, int(average(totalArray)),timeNow.UnixNano())
	done:
	}
}
