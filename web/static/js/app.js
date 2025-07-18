// Go Proverbs Interactive Features

// DOM Content Loaded
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

function initializeApp() {
    // Initialize search functionality
    initializeSearch();
    
    // Initialize copy functionality
    initializeCopyButtons();
    
    // Initialize share functionality
    initializeShareButtons();
    
    // Initialize keyboard shortcuts
    initializeKeyboardShortcuts();
    
    // Initialize animations
    initializeAnimations();
    
    // Initialize theme
    initializeTheme();
}

// Search Functionality
function initializeSearch() {
    const searchInput = document.querySelector('.search-input');
    const searchForm = document.querySelector('.search-form');
    
    if (searchInput) {
        // Auto-focus search input on search page
        if (window.location.pathname === '/search' && !searchInput.value) {
            searchInput.focus();
        }
        
        // Search suggestions (simple implementation)
        searchInput.addEventListener('input', debounce(handleSearchInput, 300));
        
        // Clear search functionality
        const clearButton = document.querySelector('.search-clear');
        if (clearButton) {
            clearButton.addEventListener('click', function() {
                searchInput.value = '';
                searchInput.focus();
            });
        }
    }
    
    if (searchForm) {
        searchForm.addEventListener('submit', function(e) {
            const query = searchInput.value.trim();
            if (!query) {
                e.preventDefault();
                searchInput.focus();
            }
        });
    }
}

function handleSearchInput(event) {
    const query = event.target.value.trim();
    if (query.length < 2) return;
    
    // Simple search suggestions could be implemented here
    // For now, we'll just update the URL without making requests
    console.log('Search query:', query);
}

// Copy Functionality
function initializeCopyButtons() {
    document.addEventListener('click', function(e) {
        if (e.target.closest('[data-copy]')) {
            const button = e.target.closest('[data-copy]');
            const text = button.getAttribute('data-copy');
            copyToClipboard(text, button);
        }
    });
}

function copyToClipboard(text, button) {
    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(() => {
            showCopySuccess(button);
        }).catch(err => {
            console.error('Failed to copy text: ', err);
            fallbackCopyTextToClipboard(text, button);
        });
    } else {
        fallbackCopyTextToClipboard(text, button);
    }
}

function fallbackCopyTextToClipboard(text, button) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        document.execCommand('copy');
        showCopySuccess(button);
    } catch (err) {
        console.error('Fallback: Oops, unable to copy', err);
    }
    
    document.body.removeChild(textArea);
}

function showCopySuccess(button) {
    const originalTitle = button.title || button.getAttribute('aria-label') || 'Copy';
    const originalText = button.textContent;
    
    // Update button state
    button.title = 'Copied!';
    if (button.textContent && !button.querySelector('svg')) {
        button.textContent = 'Copied!';
    }
    
    // Add success class
    button.classList.add('copy-success');
    
    // Reset after 2 seconds
    setTimeout(() => {
        button.title = originalTitle;
        if (button.textContent && !button.querySelector('svg')) {
            button.textContent = originalText;
        }
        button.classList.remove('copy-success');
    }, 2000);
}

// Share Functionality
function initializeShareButtons() {
    document.addEventListener('click', function(e) {
        if (e.target.closest('[data-share]')) {
            const button = e.target.closest('[data-share]');
            const shareData = JSON.parse(button.getAttribute('data-share'));
            shareContent(shareData);
        }
    });
}

function shareContent(shareData) {
    if (navigator.share && /Android|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)) {
        navigator.share(shareData).catch(err => {
            console.log('Error sharing:', err);
            fallbackShare(shareData);
        });
    } else {
        fallbackShare(shareData);
    }
}

function fallbackShare(shareData) {
    const shareText = `${shareData.title} - ${shareData.url}`;
    copyToClipboard(shareText, document.activeElement);
}

// Keyboard Shortcuts
function initializeKeyboardShortcuts() {
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K for search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            const searchInput = document.querySelector('.search-input');
            if (searchInput) {
                searchInput.focus();
                searchInput.select();
            }
        }
        
        // Escape to clear search
        if (e.key === 'Escape') {
            const searchInput = document.querySelector('.search-input');
            if (searchInput && document.activeElement === searchInput) {
                searchInput.blur();
            }
        }
        
        // R for random proverb
        if (e.key === 'r' && !e.ctrlKey && !e.metaKey && !isInputFocused()) {
            window.location.href = '/random';
        }
        
        // H for home
        if (e.key === 'h' && !e.ctrlKey && !e.metaKey && !isInputFocused()) {
            window.location.href = '/';
        }
    });
}

function isInputFocused() {
    const activeElement = document.activeElement;
    return activeElement && (
        activeElement.tagName === 'INPUT' ||
        activeElement.tagName === 'TEXTAREA' ||
        activeElement.contentEditable === 'true'
    );
}

// Animations
function initializeAnimations() {
    // Intersection Observer for fade-in animations
    if ('IntersectionObserver' in window) {
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('fade-in');
                    observer.unobserve(entry.target);
                }
            });
        }, {
            threshold: 0.1,
            rootMargin: '0px 0px -50px 0px'
        });
        
        // Observe cards and other elements
        document.querySelectorAll('.card, .proverb-card, .category-card, .stat-card').forEach(el => {
            observer.observe(el);
        });
    }
    
    // Smooth scroll for anchor links
    document.addEventListener('click', function(e) {
        const link = e.target.closest('a[href^="#"]');
        if (link) {
            e.preventDefault();
            const target = document.querySelector(link.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        }
    });
}

// Theme Management
function initializeTheme() {
    // Check for saved theme preference or default to light mode
    const savedTheme = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    
    if (savedTheme) {
        document.documentElement.setAttribute('data-theme', savedTheme);
    } else if (prefersDark) {
        document.documentElement.setAttribute('data-theme', 'dark');
    }
    
    // Listen for theme changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
        if (!localStorage.getItem('theme')) {
            document.documentElement.setAttribute('data-theme', e.matches ? 'dark' : 'light');
        }
    });
}

// Utility Functions
function debounce(func, wait, immediate) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            timeout = null;
            if (!immediate) func(...args);
        };
        const callNow = immediate && !timeout;
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
        if (callNow) func(...args);
    };
}

function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// Global functions for template use
window.copyProverbLink = function(id) {
    const url = window.location.origin + '/proverbs/' + id;
    copyToClipboard(url, document.activeElement);
};

window.shareProverb = function(id, title) {
    const url = window.location.origin + '/proverbs/' + id;
    const shareData = {
        title: `Go Proverb: ${title}`,
        text: `Check out this Go proverb: "${title}"`,
        url: url
    };
    shareContent(shareData);
};

window.clearSearch = function() {
    const searchInput = document.querySelector('.search-input');
    if (searchInput) {
        searchInput.value = '';
        searchInput.focus();
    }
};

// Performance monitoring
if ('performance' in window) {
    window.addEventListener('load', function() {
        setTimeout(() => {
            const perfData = performance.getEntriesByType('navigation')[0];
            if (perfData) {
                console.log('Page load time:', perfData.loadEventEnd - perfData.loadEventStart, 'ms');
            }
        }, 0);
    });
}

// Error handling
window.addEventListener('error', function(e) {
    console.error('JavaScript error:', e.error);
    // Could send to analytics or error reporting service
});

// Service Worker registration (for future PWA features)
if ('serviceWorker' in navigator) {
    window.addEventListener('load', function() {
        // Uncomment when service worker is implemented
        // navigator.serviceWorker.register('/sw.js')
        //     .then(registration => console.log('SW registered:', registration))
        //     .catch(error => console.log('SW registration failed:', error));
    });
}

// Export for module use if needed
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        copyToClipboard,
        shareContent,
        debounce,
        throttle
    };
}