document.addEventListener('DOMContentLoaded', function() {
    // Save scroll position before form submission
    document.querySelectorAll('form').forEach(form => {
        form.addEventListener('submit', function() {
            localStorage.setItem('scrollPos', window.scrollY);
        });
    });
    
    // Restore scroll position after page load
    const scrollPos = localStorage.getItem('scrollPos');
    if (scrollPos) {
        window.scrollTo(0, parseInt(scrollPos));
        localStorage.removeItem('scrollPos');
    }
});