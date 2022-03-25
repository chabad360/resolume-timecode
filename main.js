// import OSC from "./osc.min.js";

function main(){
    "use strict";

    const plugin = new OSC.WebsocketClientPlugin({ host: location.hostname, port: location.port });
    const osc = new OSC({ plugin: plugin });

    osc.on('open', () => {statusLabel.innerHTML = "Server Running";});
    osc.on('/name', (message) => procName(message));
    osc.on('/message', (message) => procMsg(message));
    osc.on('/time', (message) => procTime(message));
    osc.on('/refresh', () => location.reload());
    osc.on('/connect', () => reset());
    osc.on('/stop', () => plugin.close());
    osc.on('/tminus', (message) => procTminus(message));
    osc.on('close', () => close());

    const timecodeHours = document.getElementById("timecode-hours");
    const timecodeMinutes = document.getElementById("timecode-minutes");
    const timecodeSeconds = document.getElementById("timecode-seconds");
    const timecodeMS = document.getElementById("timecode-ms");
    const timecodeMinus = document.getElementsByClassName("minus");
    const timecodeClipName = document.getElementById("clipname");
    const table = document.getElementById("table");
    const tableBorder = document.getElementById("tableborder");
    const clipLength = document.getElementById("ms");
    const statusLabel = document.getElementById("status");
    const message = document.getElementById("msg");

    reset();
    osc.open();

    function close() {
        statusLabel.innerHTML = "Server Stopped";

        timecodeHours.innerHTML = "00";
        timecodeMinutes.innerHTML = "00";
        timecodeSeconds.innerHTML = "00";
        timecodeMS.innerHTML = "000";
        clipLength.innerHTML = "0.000s";

        table.style.color = "#ff4545";
        tableBorder.style.borderColor = "#ff4545";
    }

    function procName(data) {
        timecodeClipName.innerHTML = data.args[0];
    }

    function procTminus(data) {
        data.args[0] === true ? timecodeMinus[0].innerHTML = "-" : timecodeMinus[0].innerHTML = '+'
    }

    function reset() {
        timecodeHours.innerHTML = "00";
        timecodeMinutes.innerHTML = "00";
        timecodeSeconds.innerHTML = "00";
        timecodeMS.innerHTML = "000";
        clipLength.innerHTML = "0.000s";
    }

    async function procMsg(data) {
        data = data.args[0];
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
        clipLength.innerHTML = data.args[1];

        data = data.args[0].split(":");
        timecodeHours.innerHTML = data[0].substring(1);
        timecodeMinutes.innerHTML = data[1];
        timecodeSeconds.innerHTML = data[2].split(".")[0];
        timecodeMS.innerHTML = data[2].split(".")[1];

        if (parseInt(data[2]) <= 10 && parseInt(data[1]) < 1 && parseInt(data[0].substring(1)) < 1) {
            table.style.color = "#ff4545";
            tableBorder.style.borderColor = "#ff4545";
        } else {
            table.style.color = "#45ff45";
            tableBorder.style.borderColor = "#4b5457";
        }
    }
}

main();