"use strict";

const socket            = new WebSocket(`ws://${location.host}/ws`);

const timecodeHours     = document.getElementById("timecode-hours");
const timecodeMinutes   = document.getElementById("timecode-minutes");
const timecodeSeconds   = document.getElementById("timecode-seconds");
const timecodeMS        = document.getElementById("timecode-ms");
const timecodeClipName  = document.getElementById("clipname");
const table             = document.getElementById('table');
const tableBorder       = document.getElementById('tableborder');
const clipLength        = document.getElementById("ms");
const status            = document.getElementById("status");
const message           = document.getElementById("msg");

const multiplier = 100000000000; // This constant is used to avoid JSs famous floating point pitfalls

let clipName         = "";
let directionForward = true;

let timePrev    = Date();
let posPrev     = 0;
let samples     = 0;

let posIntervalBuffer   = [0];
let timeIntervalBuffer  = [0];
let estSizeBuffer       = [0];

reset();

function maxAppend(array, value, limit) {
    if (array.unshift(value) > limit) {
        array.pop();
    }
    return array;
}

function average(array){
    return Math.trunc(array.reduce((a,b) => (a+b)) / array.length);
}

function isWithin(original, newNum, percent) {
    let p = (original / 100) * percent;
    return !((newNum > original + p || newNum < original - p) && original !== 0);
}

socket.addEventListener('message', function (event) {
    let timeNow = new Date();
    let data    = event.data.toString();

    status.innerHTML = "Server Running";

    if (data.includes("/position ,f ")) {
        procPos(data, timeNow);
    } else if (data.includes("direction ,i ")) {
        procDirection(data);
    } else if (data.includes("/name ,s ")) {
        procName(data);
    } else if (data.includes("/message ,s ")) {
        procMsg(data);
    } else if (data.includes("/refresh ")) {
        location.reload();
    } else if (data.includes("/connect")) {
        reset()
    } else if (data.includes("/stop ")) {
        socket.close();
    }
});

socket.addEventListener('close', function () {
    status.innerHTML = "Server Stopped";

    timecodeHours.innerHTML = "00";
    timecodeMinutes.innerHTML = "00";
    timecodeSeconds.innerHTML = "00";
    timecodeMS.innerHTML = "000";
    clipLength.innerHTML = '0.000s'

    table.style.color = "#ff4545";
    tableBorder.style.borderColor = "#ff4545";
})

function procDirection(data) {
    directionForward = data.substring(data.length - 1) !== "0";
    reset();
}

function procName(data) {
    data = data.replace("/name ,s ", "");
    timecodeClipName.innerHTML = data;
    if (data !== clipName) {
        clipName = data;
        reset();
    }
}

function reset() {
    samples            = 0;
    posPrev            = 0;
    posIntervalBuffer  = [0];
    timeIntervalBuffer = [0];
    estSizeBuffer      = [0];

    timecodeHours.innerHTML     = '00';
    timecodeMinutes.innerHTML   = '00';
    timecodeSeconds.innerHTML   = '00';
    timecodeMS.innerHTML        = '000';
    clipLength.innerHTML        = '0.000s';
}

async function procMsg(data) {
    data = data.replace("/message ,s ", "");
    message.innerHTML = (data === "") ? "Timecode Monitor" : data;
    if (data === "") {
        return;
    }
    for (let i = 0; i < 3; i++) {
        message.style.color = "#ff4545";
        await new Promise(r => setTimeout(r, 500));
        message.style.color = "#FDFBF7"
        await new Promise(r => setTimeout(r, 500));
    }
}

function procPos(msg, timeNow) {
    let pos = multiplier * parseFloat(msg.split(" ").pop());

    if (!directionForward) {
        pos = multiplier - pos
    }

    if (pos < 50) {
        posPrev = 0;
    }

    let currentPosInterval  = pos - posPrev;
    let currentTimeInterval = (timeNow - timePrev) * multiplier;

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
    if (samples > 1000 && samples < 1500 && isWithin(prevEstSize, currentEstSize, 0.001)) {
        estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 500);
    } else if (samples > 500 && samples < 1000 && isWithin(prevEstSize, currentEstSize, 1)) {
        estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 250);
    } else if (samples < 500) {
        estSizeBuffer = maxAppend(estSizeBuffer, currentEstSize, 100);
    }

    samples++;

    let t = ((average(estSizeBuffer)) * (multiplier - pos)) / multiplier;

    posPrev  = pos;
    timePrev = timeNow;

    let timeActual = new Date(t);
    timecodeHours.innerHTML     = timeActual.getUTCHours().toString().padStart(2, '0');
    timecodeMinutes.innerHTML   = timeActual.getUTCMinutes().toString().padStart(2, '0');
    timecodeSeconds.innerHTML   = timeActual.getUTCSeconds().toString().padStart(2, '0');
    timecodeMS.innerHTML        = timeActual.getUTCMilliseconds().toString().padStart(3, '0');
    clipLength.innerHTML        = `${average(estSizeBuffer)/1000}s`;

    if (timeActual.getTime() / 1000 <= 11) {
        table.style.color = "#ff4545";
        tableBorder.style.borderColor = "#ff4545";
    } else {
        table.style.color = "#45ff45";
        tableBorder.style.borderColor = "#4b5457";
    }
    // console.log(`pos: ${pos}\taverage: ${a}\ti: ${interval}\ttime: ${d}\ttimeAverage: ${ta}\ttimeActual: ${t}\ttimeTotal: ${average(totalArray)}\ttimeLeft: ${timeLeft}`);
}
