package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	//"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/hypebeast/go-osc/osc"


	//"golang.org/x/exp/shiny/materialdesign/icons"
)

func main() {
	timePrev = time.Now()
	addr := ":7001"
	client := osc.NewClient("192.168.254.165", 7000)
	server := &osc.Server{}
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println("Couldn't listen: ", err)
	}
	defer conn.Close()

	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/composition/selectedclip/transport/position", procMsg)
	d.AddMsgHandler("/composition/selectedclip/name", clipName)

	ticker2 := time.NewTicker(time.Second)

	go func() {
		fmt.Println("Start listening on", addr)

		for {
			packet, err := server.ReceivePacket(conn)
			timeNow = time.Now()
			if err != nil {
				fmt.Println("Server error: " + err.Error())
				os.Exit(1)
			}

			d.Dispatch(packet)
		}
	}()

	go func() {
		message := osc.NewMessage("/composition/selectedclip/name")
		message.Append("?")
		for {
			<- ticker2.C
			client.Send(message)
		}
	}()

	go func() {
		w = app.NewWindow(app.Size(unit.Dp(750), unit.Dp(120)), app.Title("Resolume Time"))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())

	var ops op.Ops
	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				// Reset the layout.Context for a new frame.
				gtx := layout.NewContext(&ops, e)

				// Draw the state into ops based on events in e.Queue.
				draw(gtx, th)

				// Update the display.
				e.Frame(gtx.Ops)
			}
		}
	}
}


func draw(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(material.H1(th, timeLeft).Layout),
	)
}

var (
	posPrev  float32
	timePrev time.Time
	timeNow time.Time
	tPrev int
	samples int
	timeLeft string
	interval float32
	duration float32 = 10000
	div float32 = 1000
	array = []float32{0}
	timeArray = []float32{0}
	totalArray = []float32{0}
	name string
	w *app.Window
)

//const duration float32 = 10

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
	for i := 0; i<len(array); i++ {
		f = f + array[i]
	}
	return f/float32(len(array))
}

func within(original float32, new float32, percent float32) bool {
	p := original / 100 * percent
	if (new > original + p || new < original - p) && original != 0 {
		return false
	}
	return true
}

func clipName(msg *osc.Message) {
	nameString := strings.TrimPrefix(msg.String(), "/composition/selectedclip/name ,s ")
	if name != nameString {
		name = nameString
		samples = 0
	}
}

func procMsg(msg *osc.Message) {
	var t int
	var timeActual time.Duration
	posString := strings.TrimPrefix(msg.String(), "/composition/selectedclip/transport/position ,f ")
	pos64, _ := strconv.ParseFloat(posString, 32)
	pos := float32(pos64)

	a := average(array)
	ta := average(timeArray)
	i := pos-posPrev
	d := float32(timeNow.Sub(timePrev).Microseconds())
	if (i == 0 || d == 0) {
		goto done
	}
	if !within(ta, d, 50) && samples > 500{
		fmt.Printf("w")
	} else {
		timeArray = maxAppend(timeArray, d, 100)
		array = maxAppend(array, i, 100)
		interval = average(array)
		duration = average(timeArray)/div

		td := duration*(1/interval)
		tA := average(totalArray)
		if within(tA, td, 0.001) && samples > 1000 && samples < 3000{
			totalArray = maxAppend(totalArray, td, 500)
		} else if within(tA, td, 1) && samples > 500 {
			totalArray = maxAppend(totalArray, td, 500)
		} else if samples < 500{
			totalArray = maxAppend(totalArray, td, 100)
		}
	}
	samples++


	//t = (duration/div)*((1-pos)/interval)
	//t = (duration/div)*((1-pos)/(interval/samples))
	//t = int((duration)*((1-pos)/interval))
	t = int((average(totalArray)/1)*(1-pos))

	posPrev = pos
	timePrev = timeNow
	timeActual, _ = time.ParseDuration(fmt.Sprintf("%dms", int(t)))
	timeLeft = fmt.Sprintf("-%02d:%02d:%02d.%03d", int(timeActual.Hours()),
		int(timeActual.Minutes())-(60*int(timeActual.Hours())),
		int(timeActual.Seconds())-(60*int(timeActual.Minutes())),
		int(timeActual.Milliseconds()-(1000*int64(timeActual.Seconds()))))
	fmt.Printf("pos: %f\taverage: %f\t timeAverage: %f\ti: %f\ttimePrev: %d\ttimeActual: %d\ttimeTotal: %d\ttimeLeft: %s\tname: %s\n", pos, a, ta, interval, tPrev, t, int(average(totalArray)), timeLeft, name)
done:
	if w != nil {
		w.Invalidate()
	}
}
