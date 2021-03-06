"use strict";

const socket    = new WebSocket('ws://'+location.hostname+(location.port ? ':'+location.port: '')+'/ws');
const timecode  = document.getElementById("timecode");
const ms        = document.getElementById("ms");
const mult      = 10000000000;

let clipName    = "";
let timePrev    = Date();
let posPrev     = 0;
let samples     = 0;
let posIntervalBuffer   = [];
let timeIntervalBuffer  = [];
let estSizeBuffer       = [];

reset();

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
        f += array[i];
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

    if (data.includes("/transport/position ")) {
        procPos(data, timeNow);
    } else if (data.includes("/name ")) {
        procName(data);
    } else if (data.includes("/refresh ")) {
        location.reload();
    } else if (data.includes("/stop ")) {
        socket.close();
    }
});

socket.addEventListener('close', function () {
    timecode.innerHTML = "Server Stopped";
    ms.innerHTML       = 'Clip Length: 0.000s'
})

function procName(data) {
    data = data.replace("/name ,s ", "");
    if (data !== clipName) {
        clipName = data;
        reset();
    }
}

function reset() {
    samples            = 0;
    posPrev            = 0;
    posIntervalBuffer  = [];
    timeIntervalBuffer = [];
    estSizeBuffer      = [];

    timecode.innerHTML = '-000:00:00.000'
    ms.innerHTML       = 'Clip Length: 0.000s'
}

function procPos(msg, timeNow) {
    let pos = mult * parseFloat(msg.replace("/transport/position ,f ", ""));
    if (pos < 50) {
        posPrev = 0;
    }

    let currentPosInterval  = pos - posPrev;
    let currentTimeInterval = (timeNow - timePrev) * mult;

    if (currentPosInterval === 0 || currentTimeInterval === 0) {
        return;
    }

    if (currentPosInterval < 0 && posPrev > 0) {
        return;
    }

    posIntervalBuffer  = maxAppend(posIntervalBuffer, currentPosInterval, 100);
    timeIntervalBuffer = maxAppend(timeIntervalBuffer, currentTimeInterval, 100);

    let posInterval  = average(posIntervalBuffer);
    let timeInterval = average(timeIntervalBuffer);

    let currentEstSize = Math.trunc(timeInterval * (1 / posInterval));
    let prevEstSize = average(estSizeBuffer);
    if (samples > 1000 && samples < 1500 && within(prevEstSize, currentEstSize, 0.001)) {
        estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 500);
    } else if (samples > 500 && samples < 1000 && within(prevEstSize, currentEstSize, 1)) {
        estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 250);
    } else if (samples < 500) {
        estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 100);
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
    ms.innerHTML = `Clip Length: ${average(estSizeBuffer)/1000}s`;
    // console.log(`pos: ${pos}\taverage: ${a}\ti: ${interval}\ttime: ${d}\ttimeAverage: ${ta}\ttimeActual: ${t}\ttimeTotal: ${average(totalArray)}\ttimeLeft: ${timeLeft}`);
}
