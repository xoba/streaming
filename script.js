$(function () {

    const socket = new WebSocket('ws://localhost:8080/echo');

    socket.onopen = function () {
        console.log('WebSocket connection established');
    };

    socket.onmessage = function (message) {
        console.log('Message received from server', message.data);
    };

    socket.onerror = function (error) {
        console.error('WebSocket error', error);
    };

    socket.onclose = function (event) {
        console.log('WebSocket connection closed', event);
    };

    // Check for mediaDevices API support
    if (navigator.mediaDevices) {
        navigator.mediaDevices.getUserMedia({ audio: true })
            .then(stream => {
                // Initialize MediaRecorder with desired format
                const mediaRecorder = new MediaRecorder(stream);

                mediaRecorder.onstart = function (e) {
                    console.log("onstart")
                    this.chunks = [];
                };

                mediaRecorder.ondataavailable = function (e) {
                    console.log("got chunk");
                    this.chunks.push(e.data);
                    // Send the audio chunk to the server using WebSocket
                    socket.send(e.data);
                };

                mediaRecorder.onstop = function (e) {
                    console.log("onstop")
                };

                // Start recording
                mediaRecorder.start(100);

                // Example: Stop recording after 5 seconds
                setTimeout(() => mediaRecorder.stop(), 5000);
            })
            .catch(error => {
                console.error('Error accessing the microphone', error);
            });
    } else {
        console.error('navigator.mediaDevices not supported');
    }


})