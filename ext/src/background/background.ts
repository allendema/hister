import {
    sendPageData,
    sendResult,
} from '../modules/network';

const missingURLMsg = {"error": "Missing or invalid Hister server URL. Configure it in the addon popup."};
// TODO check source
async function cjsMsgHandler(request, sender, sendResponse) {
    chrome.storage.local.get(['histerURL']).then(data => {
        let u = data['histerURL'] || "";
        if(!u) {
            chrome.tabs.sendMessage(sender.tab.id, missingURLMsg);
            return;
        }
        if(!u.endsWith('/')) {
            u += '/';
        }
        if(request.pageData) {
            sendPageData(u+"add", request.pageData).then((r) => sendResponse({"msg": "ok"})).catch(err => sendResponse({"error": err}));
        }
        if(request.resultData) {
            sendResult(u+"history", request.resultData).then((r) => sendResponse({"msg": "ok"})).catch(err => sendResponse({"error": err}));
        }
    }).catch(error => {
        chrome.tabs.sendMessage(sender.tab.id, missingURLMsg);
    });
}

chrome.runtime.onMessage.addListener(cjsMsgHandler);
