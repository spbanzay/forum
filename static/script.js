// === Темная тема ===
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

// === Клиентский поиск по постам ===
function initSearch() {
    const input = document.querySelector('input[name="q"]');
    const form = document.getElementById('searchForm');
    if (!form || !input) return;

    form.addEventListener('submit', function (e) {
        e.preventDefault();
        const query = input.value.trim().toLowerCase();
        const posts = document.querySelectorAll('.post-card');

        posts.forEach(post => {
            const title = post.dataset.title?.toLowerCase() || '';
            const content = post.dataset.content?.toLowerCase() || '';
            const matches = title.includes(query) || content.includes(query);
            post.style.display = matches ? '' : 'none';
        });
    });
}

// === Инициализация после загрузки ===
function init() {
    initTheme();
    initSearch();
}

if (document.readyState !== 'loading') {
    init();
} else {
    document.addEventListener('DOMContentLoaded', init);
}

document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('.show-login-popup').forEach(btn => {
      btn.addEventListener('click', () => {
        const modal = new bootstrap.Modal(document.getElementById('loginModal'));
        modal.show();
      });
    });
  });
