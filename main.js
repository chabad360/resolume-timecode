"use strict";

const socket    = new WebSocket('ws://localhost/ws');

const timecode  = document.getElementById("timecode");
const mult      = 1000000000;

let clipName    = "";
let timePrev    = Date();
let posPrev     = 0;
let samples     = 0;
let posIntervalBuffer   = [];
let timeIntervalBuffer  = [];
let estSizeBuffer       = [];

function maxAppend(array, value, limit) {
    array.unshift(value);
    if (array.length > limit) {
        array.pop();
    }
    return array;
}

function average(array){
    let f = 0;
    for (let i = 0; i < array.length; i++) {
        f = f + array[i];
    }
    return Math.trunc(f / array.length);
}

function within(original, newNum, percent) {
    let p = original / 100 * percent;
    return !((newNum > original + p || newNum < original - p) && original !== 0);
}


socket.addEventListener('message', function (event) {
    let timeNow = new Date();
    let data    = event.data.toString();

    if (data.includes("/composition/selectedclip/transport/position ")) {
        procPos(data, timeNow);
    } else if (data.includes("/composition/selectedclip/name ")) {
        procName(data);
    }
});

function procName(data) {
    data = data.replace("/composition/selectedclip/name ,s ", "");
    if (data !== clipName) {
        clipName = data;
        reset();
    }
}

function reset() {
    samples     = 0;
    posIntervalBuffer   = [];
    timeIntervalBuffer  = [];
    estSizeBuffer = [];
}

function procPos(msg, timeNow) {
    let posInterval  = 0;
    let timeInterval = 0;

    let pos = mult * parseFloat(msg.replace("/composition/selectedclip/transport/position ,f ", ""));
    if (pos < 5) {
        reset();
    }

    // let a = average(posBuffer)
    let prevTimeInterval    = average(timeIntervalBuffer);
    let currentPosInterval  = pos - posPrev;
    let currentTimeInterval = (timeNow - timePrev) * mult;

    if (currentPosInterval === 0 || currentTimeInterval === 0) {
        return;
    }

    if (!within(prevTimeInterval, currentTimeInterval, 50) && samples > 500) {
        //console.log("w");
    } else {
        posIntervalBuffer    = maxAppend(posIntervalBuffer, i, 100);
        timeIntervalBuffer   = maxAppend(timeIntervalBuffer, d, 100);

        posInterval  = average(posIntervalBuffer);
        timeInterval = average(timeIntervalBuffer);

        let td = Math.trunc(timeInterval * (1 / posInterval));
        let tA = average(estSizeBuffer);
        if (within(tA, td, 0.001) && samples > 1000 && samples < 2000) {
            estSizeBuffer = maxAppend(estSizeBuffer, td, 500);
        } else if (within(tA, td, 1) && samples > 500) {
            estSizeBuffer = maxAppend(estSizeBuffer, td, 250);
        } else if (samples < 500) {
            estSizeBuffer = maxAppend(estSizeBuffer, td, 100);
        }
    }

    samples++;

    let t = ((average(estSizeBuffer)) * (mult - pos)) / mult;

    posPrev  = pos;
    timePrev = timeNow;

    let timeActual = new Date(t);
    timecode.innerHTML = `-${
        timeActual.getUTCHours().toString().padStart(3, '0')}:${
        timeActual.getUTCMinutes().toString().padStart(2, '0')}:${
        timeActual.getUTCSeconds().toString().padStart(2, '0')}.${
        timeActual.getUTCMilliseconds().toString().padStart(3, '0')}`;
    // console.log(`pos: ${pos}\taverage: ${a}\ti: ${interval}\ttime: ${d}\ttimeAverage: ${ta}\ttimeActual: ${t}\ttimeTotal: ${average(totalArray)}\ttimeLeft: ${timeLeft}`);
}
