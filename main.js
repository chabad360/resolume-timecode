"use strict";

const socket            = new WebSocket(`ws://${location.host}/ws`);

const timecodeHours     = document.getElementById("timecode-hours");
const timecodeMinutes   = document.getElementById("timecode-minutes");
const timecodeSeconds   = document.getElementById("timecode-seconds");
const timecodeMS        = document.getElementById("timecode-ms");
const timecodeClipName  = document.getElementById("clipname");
const table             = document.getElementById("table");
const tableBorder       = document.getElementById("tableborder");
const clipLength        = document.getElementById("ms");
const statusLabel       = document.getElementById("status");
const message           = document.getElementById("msg");

reset();

socket.addEventListener("message", function (event) {
    let data    = event.data.toString();

    statusLabel.innerHTML = "Server Running";

    if (data.includes("/time ,ss ")) {
        procTime(data);
    } else if (data.includes("/name ,s ")) {
        procName(data);
    } else if (data.includes("/message ,s ")) {
        procMsg(data);
    } else if (data.includes("/refresh ")) {
        location.reload();
    } else if (data.includes("/connect")) {
        reset();
    } else if (data.includes("/stop ")) {
        socket.close();
    }
});

socket.addEventListener("close", function () {
    statusLabel.innerHTML = "Server Stopped";

    timecodeHours.innerHTML = "00";
    timecodeMinutes.innerHTML = "00";
    timecodeSeconds.innerHTML = "00";
    timecodeMS.innerHTML = "000";
    clipLength.innerHTML = "0.000s";

    table.style.color = "#ff4545";
    tableBorder.style.borderColor = "#ff4545";
});

function procName(data) {
    timecodeClipName.innerHTML = data.replace("/name ,s ", "");
}

function reset() {
    timecodeHours.innerHTML     = "00";
    timecodeMinutes.innerHTML   = "00";
    timecodeSeconds.innerHTML   = "00";
    timecodeMS.innerHTML        = "000";
    clipLength.innerHTML        = "0.000s";
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
        message.style.color = "#FDFBF7";
        await new Promise(r => setTimeout(r, 500));
    }
}

function procTime(data) {
    data = data.split(" ");
    clipLength.innerHTML = data.pop().toString();

    data = data.pop().split(":");
    timecodeHours.innerHTML     = data[0].substring(1);
    timecodeMinutes.innerHTML   = data[1];
    timecodeSeconds.innerHTML   = data[2].split(".")[0];
    timecodeMS.innerHTML        = data[2].split(".")[1];

    if (parseInt(data[2]) <= 11 && parseInt(data[1]) < 1 && parseInt(data[0].substring(1)) < 1) {
        table.style.color = "#ff4545";
        tableBorder.style.borderColor = "#ff4545";
    } else {
        table.style.color = "#45ff45";
        tableBorder.style.borderColor = "#4b5457";
    }
  }
