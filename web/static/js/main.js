document.addEventListener('DOMContentLoaded', () => {
    const searchForm = document.getElementById('search-form');
    const searchInput = document.getElementById('search-input');
    const imageGrid = document.getElementById('image-grid');
    const pagination = document.getElementById('pagination');
    const prevPageButton = document.getElementById('prev-page');
    const nextPageButton = document.getElementById('next-page');
    const pageInfo = document.getElementById('page-info');

    let currentPage = 1;
    let currentTags = '';
    let totalPosts = 0;

    searchForm.addEventListener('submit', (e) => {
        e.preventDefault();
        currentTags = searchInput.value;
        currentPage = 1;
        fetchImages();
    });

    prevPageButton.addEventListener('click', () => {
        if (currentPage > 1) {
            currentPage--;
            fetchImages();
        }
    });

    nextPageButton.addEventListener('click', () => {
        currentPage++;
        fetchImages();
    });

    async function fetchImages() {
        if (!currentTags) return;

        try {
            const response = await fetch(`/api/search?tags=${encodeURIComponent(currentTags)}&page=${currentPage}&limit=40`);
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            const result = await response.json();
            totalPosts = result.total;
            renderImages(result.posts);
            updatePagination();
        } catch (error) {
            console.error('Failed to fetch images:', error);
            imageGrid.innerHTML = '<p>Failed to load images. Please try again later.</p>';
        }
    }

    function renderImages(posts) {
        imageGrid.innerHTML = '';
        if (!posts || posts.length === 0) {
            imageGrid.innerHTML = '<p>No images found.</p>';
            return;
        }

        posts.forEach(post => {
            const item = document.createElement('div');
            item.className = 'grid-item';

            const link = document.createElement('a');
            link.href = `/api/proxy?url=${encodeURIComponent(post.file_url)}`;
            link.target = '_blank';

            const img = document.createElement('img');
            img.src = `/api/proxy?url=${encodeURIComponent(post.preview_url)}`;
            img.alt = post.tags.join(', ');

            const tags = document.createElement('div');
            tags.className = 'tags';
            tags.textContent = `Source: ${post.source}`;

            link.appendChild(img);
            item.appendChild(link);
            item.appendChild(tags);
            imageGrid.appendChild(item);
        });
    }

    function updatePagination() {
        pageInfo.textContent = `Page ${currentPage}`;
        prevPageButton.disabled = currentPage === 1;
        nextPageButton.disabled = totalPosts < 40;
    }
});