if (document.readyState !== 'loading') initTheme();
document.addEventListener('DOMContentLoaded', initTheme);

function initTheme() {
    const toggle = document.getElementById('theme-toggle');
    if (!toggle) return;
    const body = document.body;
    if (localStorage.getItem('theme') === 'dark') {
        body.classList.add('dark-mode');
        toggle.innerText = '☀️';
    }

    toggle.addEventListener('click', function () {
        body.classList.toggle('dark-mode');
        if (body.classList.contains('dark-mode')) {
            localStorage.setItem('theme', 'dark');
            toggle.innerText = '☀️';
        } else {
            localStorage.setItem('theme', 'light');
            toggle.innerText = '🌙';
        }
    });
}
