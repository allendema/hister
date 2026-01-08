type Document = {
    title: string;
    text: string;
    url: string;
    html: string;
    faviconURL: string;
}

function getURL() {
	return window.location.href.replace(window.location.hash, "");
}

function extractData() : Document {
    let d : Document = {
        text: document.body.innerText,
        title: document.querySelector("title").innerText,
        url: getURL(),
        html: document.documentElement.innerHTML,
        faviconURL: new URL("/favicon.ico", getURL()).href,
    };
	let link = document.querySelector("link[rel~='icon']");
	if (link && link.getAttribute("href")) {
        d.faviconURL = new URL(link.getAttribute("href"), d.url).href;
	}
    return d;
}

export {
    Document,
    extractData,
}
