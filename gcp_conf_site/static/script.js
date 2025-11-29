document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('searchInput');
    const cards = document.querySelectorAll('.talk-card');
    const noResults = document.getElementById('no-results');

    searchInput.addEventListener('input', (e) => {
        const searchTerm = e.target.value.toLowerCase();
        let visibleCount = 0;

        cards.forEach(card => {
            const title = card.dataset.title.toLowerCase();
            const category = card.dataset.category.toLowerCase();
            const speakers = card.dataset.speakers.toLowerCase();

            if (title.includes(searchTerm) || category.includes(searchTerm) || speakers.includes(searchTerm)) {
                card.style.display = 'flex';
                visibleCount++;
            } else {
                card.style.display = 'none';
            }
        });

        if (visibleCount === 0) {
            noResults.style.display = 'block';
        } else {
            noResults.style.display = 'none';
        }
    });
});
