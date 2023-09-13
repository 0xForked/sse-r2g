document.addEventListener('DOMContentLoaded', () => {
    const unsupportedStreamScreen = document.getElementById('unsupported-stream-screen');
    const supportedStreamScreen = document.getElementById('supported-stream-screen');

    if(typeof(EventSource) === "undefined" && unsupportedStreamScreen.classList.contains("hidden")) {
        unsupportedStreamScreen.classList.remove("hidden");
        return
    }

    if(supportedStreamScreen.classList.contains("hidden")) {
        supportedStreamScreen.classList.remove("hidden");
    }

    document.getElementById('fire-notify-button').addEventListener(
        'click', async () => fireNotify());

    streamNotify();
});

function generateRandomKey(length) {
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++ ) {
        result += characters.charAt(Math.floor(
            Math.random() * characters.length));
    }
    return result;
}

function getSessionId() {
    const sessionIdKey = "session_id";
    const sessionId  = localStorage.getItem(sessionIdKey);
    if(sessionId === null) {
        const newSessionId = generateRandomKey(16);
        localStorage.setItem(sessionIdKey, newSessionId);
        return newSessionId;
    }
    return sessionId;
}

async function fireNotify() {
    const fireEventURL = `/api/v1/notifications/${getSessionId()}/fire`
    const response = await fetch(fireEventURL);
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    return await response.json();
}

function streamNotify() {
    const streamEventURL = `/api/v1/notifications/${getSessionId()}/stream`
    let streamSource = new EventSource(streamEventURL);
    streamSource.onmessage = (event) => {
        let timerInterval
        // noinspection JSUnresolvedReference,JSUnusedGlobalSymbols
        Swal.fire({
            position: 'bottom-end',
            icon: 'success',
            title: 'New Notification',
            html: `<b>${event.data}</b>`,
            timer: 2500,
            timerProgressBar: true,
            showConfirmButton: false,
            willClose: () =>  clearInterval(timerInterval)
        })
    };
    streamSource.onerror = (error) => console.log(
        "An error occurred while attempting to connect:", error);
}