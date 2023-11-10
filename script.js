$(function () {
    var log = function () {
        var n = 0
        return function () {
            let concatenatedMsg = "";
            for (let i = 0; i < arguments.length; i++) {
                concatenatedMsg += arguments[i] + " ";
            }
            const now = new Date();
            const iso8601WithMilliseconds = now.toISOString();
            n++;
            $('#log').prepend('<p>' + n + ". " + iso8601WithMilliseconds + ": " + concatenatedMsg + '</p>');
            console.log(concatenatedMsg);
        };
    }();
    const socket = new WebSocket('ws://' + window.location.host + '/ws');
    socket.onopen = function () {
        log('WebSocket connection established');
        if (navigator.mediaDevices) {
            navigator.mediaDevices.getUserMedia({ audio: true })
                .then(stream => {
                    const mediaRecorder = new MediaRecorder(stream);

                    mediaRecorder.onstart = function (e) {
                        log("onstart")
                    };

                    mediaRecorder.ondataavailable = function (e) {
                        log("got chunk of " + e.data.size + " bytes");
                        socket.send(e.data);
                    };

                    mediaRecorder.onstop = function (e) {
                        log("onstop")
                        socket.send("stop");
                    };

                    mediaRecorder.start(100);

                    setTimeout(() => mediaRecorder.stop(), 10000);
                })
                .catch(error => {
                    error('Error accessing the microphone', error);
                });
        } else {
            error('navigator.mediaDevices not supported');
        }
    };
    socket.onmessage = function (message) {
        log('Message received from server:', message.data);
    };
    socket.onerror = function (error) {
        error('WebSocket error:', error);
    };
    socket.onclose = function (event) {
        let reason;
        if (event.wasClean) {
            reason = 'Closed cleanly';
        } else {
            reason = 'Connection died';
        }
        let message = 'WebSocket connection closed: ' + reason +
            ', Code: ' + event.code;
        if (event.reason) {
            message += ', Reason: ' + event.reason;
        }
        log(message);
    };
})