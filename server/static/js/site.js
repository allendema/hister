window.addEventListener('load', (e) => {
    const themeToggle = document.getElementById('theme-toggle');
    themeToggle.addEventListener('click', toggleTheme);
    let theme = localStorage.getItem('theme');
    if(!theme) {
        theme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    }
	document.querySelector("html").setAttribute("data-theme", theme);
});

function toggleTheme() {
    const currentTheme = document.querySelector("html").getAttribute("data-theme");
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    document.querySelector("html").setAttribute("data-theme", newTheme);
    localStorage.setItem('theme', newTheme);
}
