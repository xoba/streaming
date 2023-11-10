$(function () {

    var n = 0
    var log = function (msg) {
        const now = new Date();
        const iso8601WithMilliseconds = now.toISOString();
        n++;
        $('#log').prepend('<p>' + n + ". " + iso8601WithMilliseconds + ": " + msg + '</p>');
    };

    const socket = new WebSocket('ws://' + window.location.host + '/echo');

    socket.onopen = function () {
        log('WebSocket connection established');
    };

    socket.onmessage = function (message) {
        log('Message received from server', message.data);
    };

    socket.onerror = function (error) {
        error('WebSocket error', error);
    };

    socket.onclose = function (event) {
        log('WebSocket connection closed', event);
    };

    // Check for mediaDevices API support
    if (navigator.mediaDevices) {
        navigator.mediaDevices.getUserMedia({ audio: true })
            .then(stream => {
                const mediaRecorder = new MediaRecorder(stream);

                mediaRecorder.onstart = function (e) {
                    log("onstart")
                    this.chunks = [];
                };

                mediaRecorder.ondataavailable = function (e) {
                    log("got chunk of " + e.data.size + " bytes");
                    this.chunks.push(e.data);
                    socket.send(e.data);
                };

                mediaRecorder.onstop = function (e) {
                    log("onstop")
                    socket.send("stop");
                };

                // Start recording
                mediaRecorder.start(100);

                // Example: Stop recording after 5 seconds
                setTimeout(() => mediaRecorder.stop(), 5000);
            })
            .catch(error => {
                error('Error accessing the microphone', error);
            });
    } else {
        error('navigator.mediaDevices not supported');
    }


})