function main() {
    "use strict";

    const socket = new WebSocket(`ws://${location.host}`);
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

    close();

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

    async function procMsg(data) {
        if (message.innerHTML === data) {
            return;
        }

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


    socket.onopen = () => {
        statusLabel.innerHTML = "Server running";
    };

    socket.onmessage = (event) => {
        let data = event.data.toString();
        data = JSON.parse(data);

        clipLength.innerHTML = data.cliplength;
        data.invert === false ? timecodeMinus[0].innerHTML = "-" : timecodeMinus[0].innerHTML = "+";
        timecodeClipName.innerHTML = data.clipname;

        timecodeHours.innerHTML = data.hour;
        timecodeMinutes.innerHTML = data.minute;
        timecodeSeconds.innerHTML = data.second;
        timecodeMS.innerHTML = data.ms;

        if (parseInt(data.second) <= 10 && parseInt(data.hour) < 1 && parseInt(data.minute) < 1) {
            table.style.color = "#ff4545";
            tableBorder.style.borderColor = "#ff4545";
        } else {
            table.style.color = "#45ff45";
            tableBorder.style.borderColor = "#4b5457";
        }

        procMsg(data.message);
    };

    socket.onclose = () => {
        close();
    };
}

main();