// Применение тёмной темы сразу — до initTheme()
if (localStorage.getItem('theme') === 'dark') {
    document.documentElement.classList.add('dark-mode');
}

function initTheme() {
    const toggle = document.getElementById('theme-toggle');
    if (!toggle) return;

    const isDark = document.documentElement.classList.contains('dark-mode');
    toggle.innerText = isDark ? '☀️' : '🌙';

    toggle.addEventListener('click', () => {
        document.documentElement.classList.toggle('dark-mode');
        const nowDark = document.documentElement.classList.contains('dark-mode');
        localStorage.setItem('theme', nowDark ? 'dark' : 'light');
        toggle.innerText = nowDark ? '☀️' : '🌙';
    });
}

if (document.readyState !== 'loading') {
    initTheme();
} else {
    document.addEventListener('DOMContentLoaded', initTheme);
}
