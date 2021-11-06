"use strict";

const socket            = new WebSocket(`ws://${location.host}/ws`);

const timecodehours     = document.getElementById("timecode-hours");
const timecodeminutes   = document.getElementById("timecode-minutes");
const timecodeseconds   = document.getElementById("timecode-seconds");
const timecodems        = document.getElementById("timecode-ms");
const timecodeclipname  = document.getElementById("clipname");
const table             = document.getElementById('table');
const tableborder       = document.getElementById('tableborder');
const cliplength        = document.getElementById("ms");
const status            = document.getElementById("status");
const message          = document.getElementById("msg");

const mult      = 10000000000; // This constant is used to avoid JSs famous floating point pitfalls

let clipName    = "";
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

function within(original, newNum, percent) {
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
    } else if (data.includes("/stop ")) {
        socket.close();
    }
});

socket.addEventListener('close', function () {
    status.innerHTML = "Server Stopped";

    timecodehours.innerHTML = "00";
    timecodeminutes.innerHTML = "00";
    timecodeseconds.innerHTML = "00";
    timecodems.innerHTML = "000";
    cliplength.innerHTML = '0.000s'

    table.style.color = "#ff4545";
    tableborder.style.borderColor = "#ff4545";
})

function procDirection(data) {
    directionForward = data.substring(data.length - 1) !== "0";
    reset();
}

function procName(data) {
    data = data.replace("/name ,s ", "");
    timecodeclipname.innerHTML = data;
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

    timecodehours.innerHTML     = '00';
    timecodeminutes.innerHTML   = '00';
    timecodeseconds.innerHTML   = '00';
    timecodems.innerHTML        = '000';
    cliplength.innerHTML        = '0.000s';
}

async function procMsg(data) {
    message.innerHTML = data.replace("/message ,s ", "");
    message.style.color = "#ff4545"
    await new Promise(r => setTimeout(r, 500));
    message.style.color = "#FDFBF7"
    await new Promise(r => setTimeout(r, 500));
    message.style.color = "#ff4545"
    await new Promise(r => setTimeout(r, 500));
    message.style.color = "#FDFBF7"
    await new Promise(r => setTimeout(r, 500));
}

function procPos(msg, timeNow) {
    let pos = mult * parseFloat(msg.split(" ").pop());

    if (!directionForward) {
        pos = mult - pos
    }

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
    timecodehours.innerHTML     = timeActual.getUTCHours().toString().padStart(2, '0');
    timecodeminutes.innerHTML   = timeActual.getUTCMinutes().toString().padStart(2, '0');
    timecodeseconds.innerHTML   = timeActual.getUTCSeconds().toString().padStart(2, '0');
    timecodems.innerHTML        = timeActual.getUTCMilliseconds().toString().padStart(3, '0');
    cliplength.innerHTML        = `${average(estSizeBuffer)/1000}s`;

    if (timeActual.getTime() / 1000 <= 11) {
        table.style.color = "#ff4545";
        tableborder.style.borderColor = "#ff4545";
    } else {
        table.style.color = "#45ff45";
        tableborder.style.borderColor = "#4b5457";
    }
    // console.log(`pos: ${pos}\taverage: ${a}\ti: ${interval}\ttime: ${d}\ttimeAverage: ${ta}\ttimeActual: ${t}\ttimeTotal: ${average(totalArray)}\ttimeLeft: ${timeLeft}`);
}
