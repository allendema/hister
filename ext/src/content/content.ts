import {
    PageData,
    extractPageData,
    registerResultExtractor,
} from '../modules/extract';

let d : PageData;
// ms
const defaultSleepTime = 10*1000;
let sleepTime = defaultSleepTime;
const sleepIncrementRatio = 2;

window.addEventListener("load", extract, false);

function extract() {
    registerResultExtractor(window, r => chrome.runtime.sendMessage({resultData:  r}));
    try {
        d = extractPageData();
    } catch(e) {
        console.log("failed to extract page data:", e);
        return;
    }
    chrome.runtime.sendMessage({pageData:  d}, resp => {});
    setTimeout(update, sleepTime);
}


function update() {
    let d2;
    try {
        d2 = extractPageData();
    } catch(e) {
        console.log("failed to extract page data", e);
        return;
    }
    if(d2.html != d.html) {
        sleepTime = defaultSleepTime;
        d = d2;
        chrome.runtime.sendMessage({pageData:  d}, resp => {});
    } else {
        sleepTime *= sleepIncrementRatio;
    }
    setTimeout(update, sleepTime);
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
