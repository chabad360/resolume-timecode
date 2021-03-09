"use strict";

const socket = new WebSocket('ws://localhost/ws');
const timecode = document.getElementById("timecode")
const mult = 1000000000

var clipName = ""
var posPrev = 0
var timePrev
var timeNow
var tPrev
var samples = 0
var interval = 0
var duration = 10000
const div = 1000
var posArray = []
var timeArray = []
var totalArray = []

function maxAppend(array, value, limit) {
    array.unshift(value)
    if (array.length > limit) {
        array.pop()
    }
    return array
}

function average(array){
    var f = 0
    for (let i = 0; i < array.length; i++) {
        f = f + array[i]
    }
    return Math.trunc(f / array.length)
}

function within(original, newNum, percent) {
    let p = original / 100 * percent
    return !((newNum > original + p || newNum < original - p) && original !== 0);
}


socket.addEventListener('message', function (event) {
    timeNow = new Date()
    let data = event.data.toString();
    if (data.includes("/composition/selectedclip/transport/position")) {
        procPos(data)
    } else if (data.includes("/composition/selectedclip/name")) {
        procName(data)
    }
});

function procName(data) {
    data = data.replace("/composition/selectedclip/name ,s ", "")
    if (data !== clipName) {
        clipName = data
        reset()
    }
    // console.log(clipName)
}

function reset() {
    samples = 0
    totalArray = []
    timeArray = []
    posArray = []
}

function procPos(msg) {
    let pos = parseFloat(msg.replace("/composition/selectedclip/transport/position ,f ", "")) * mult
    if (pos < 5) {
        reset()
    }

    let a = average(posArray)
    let ta = average(timeArray)
    let i = pos - posPrev
    let d = (timeNow - timePrev) * mult
    if (i === 0 || d === 0) {
        return
    }
    if (!within(ta, d, 50) && samples > 500) {
        //fmt.Printf("w")
    } else {
        timeArray = maxAppend(timeArray, d, 100)
        posArray = maxAppend(posArray, i, 100)
        interval = average(posArray)
        duration = average(timeArray)

        let td = duration * (1 / interval)
        let tA = average(totalArray)
        if (within(tA, td, 0.001) && samples > 1000 && samples < 2000) {
            totalArray = maxAppend(totalArray, td, 500)
        } else if (within(tA, td, 1) && samples > 500) {
            totalArray = maxAppend(totalArray, td, 250)
        } else if (samples < 500) {
            totalArray = maxAppend(totalArray, td, 100)
        }
    }
    samples++

    //t = int((duration)*((1-pos)/interval))
    let t = ((average(totalArray)) * (mult - pos)) / mult

    posPrev = pos
    timePrev = timeNow
    let timeActual = new Date(t)
    let timeLeft = `-${timeActual.getUTCHours().toString().padStart(3, '0')}:${timeActual.getUTCMinutes().toString().padStart(2, '0')}:${timeActual.getUTCSeconds().toString().padStart(2, '0')}.${timeActual.getUTCMilliseconds().toString().padStart(3, '0')}`
    timecode.innerHTML = timeLeft
    // console.log(`pos: ${pos}\taverage: ${a}\ti: ${interval}\ttime: ${d}\ttimeAverage: ${ta}\ttimeActual: ${t}\ttimeTotal: ${average(totalArray)}\ttimeLeft: ${timeLeft}`)
    //fmt.Printf("%f,%f,%f,%f,%f,%d,%d,%d\n", pos, a, interval, d, ta, t, int(average(totalArray)),timeNow.UnixNano())
}
