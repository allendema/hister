import {
    Document,
    extractData,
} from '../modules/extract';

let d : Document;

window.addEventListener("load", extract, false);

function extract() {
    d = extractData();
    chrome.runtime.sendMessage({data:  d}, resp => {});
    setInterval(update, 10*1000);
}


function update() {
    let d2 = extractData();
    if(d2.html != d.html) {
        d = d2;
        chrome.runtime.sendMessage({data:  d}, resp => {});
    }
}

// Get message from background page
// TODO check sender
chrome.runtime.onMessage.addListener(function(request, sender, sendResponse) {
    if(!request) {
        return;
    }
    if(request.error) {
        alert(request.error);
        return;
    }
    if(request.action == "reindex") {
        extract();
        sendResponse({"action": "reindex", "status": "ok"});
		return;
    }
    console.log("message received", request)
});
