// Store the MaxVideoAge from the server in hours (read from data attribute to avoid editor warnings)
let maxVideoAge;

// Initialize the application when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    // Get max video age from HTML data attribute
    maxVideoAge = parseInt(document.documentElement.getAttribute('data-max-video-age') || '24');
    
    // Register event listeners for paste and clear buttons
    document.getElementById('pasteUrlBtn').addEventListener('click', pasteFromClipboard);
    document.getElementById('clearUrlBtn').addEventListener('click', clearUrlField);
    
    // Setup theme toggle functionality
    setupThemeToggle();
    
    // Clean expired entries and display history on page load
    cleanExpiredHistory();
    displayHistory();
    
    // Update countdown timers immediately and then every second
    updateCountdownTimers();
    setInterval(updateCountdownTimers, 1000);
    
    // Set up periodic cleaning of expired history entries (every 5 minutes)
    setInterval(cleanExpiredHistory, 5 * 60 * 1000);
    
    // Set up HTMX event handlers
    document.body.addEventListener('htmx:beforeRequest', function(event) {
        // Get the target button that initiated the request
        const targetId = event.detail.elt.id;
        
        if (targetId === 'downloadBtn' || !targetId) {
            document.getElementById('spinner').classList.remove('hidden');
            document.getElementById('downloadBtn').disabled = true;
            if (document.getElementById('downloadAndCopyBtn')) {
                document.getElementById('downloadAndCopyBtn').disabled = true;
            }
        } else if (targetId === 'downloadAndCopyBtn') {
            document.getElementById('spinner-copy').classList.remove('hidden');
            document.getElementById('downloadAndCopyBtn').disabled = true;
            document.getElementById('downloadBtn').disabled = true;
        }
    });

    document.body.addEventListener('htmx:afterRequest', function(event) {
        // Get the target button that initiated the request
        const targetId = event.detail.elt.id;
        
        // Re-enable buttons and hide spinners
        document.getElementById('spinner').classList.add('hidden');
        if (document.getElementById('spinner-copy')) {
            document.getElementById('spinner-copy').classList.add('hidden');
        }
        document.getElementById('downloadBtn').disabled = false;
        if (document.getElementById('downloadAndCopyBtn')) {
            document.getElementById('downloadAndCopyBtn').disabled = false;
        }
        
        // Clear the URL input field immediately after request completes
        const urlField = document.getElementById('videoUrl');
        if (urlField) {
            urlField.value = '';
            console.log("URL field cleared");
        }
        
        // Process the download result from the temporary container
        const tempResult = document.getElementById('temp-result');
        if (tempResult && tempResult.innerHTML.trim() !== '') {
            console.log("Processing download result");
            
            // Check if this is a success message
            if (!tempResult.innerHTML.includes('Error') && 
                tempResult.innerHTML.includes('alert-success')) {
                
                // If this was triggered by the download & copy button, copy the URL to clipboard
                if (targetId === 'downloadAndCopyBtn') {
                    const linkElement = tempResult.querySelector('.download-link');
                    if (linkElement && linkElement.textContent) {
                        navigator.clipboard.writeText(linkElement.textContent.trim())
                            .then(() => {
                                console.log('URL copied to clipboard');
                                // Show a notification
                                const resultElement = document.getElementById('result');
                                resultElement.innerHTML = '<div class="alert alert-success"><strong>Copied!</strong> Link copied to clipboard.</div>';
                                resultElement.classList.remove('hidden');
                                
                                // Hide the notification after a few seconds
                                setTimeout(() => {
                                    resultElement.classList.add('hidden');
                                    resultElement.innerHTML = '';
                                }, 10000);
                            })
                            .catch(err => {
                                console.error('Failed to copy URL: ', err);
                            });
                    }
                }
                
                // Add to history first before showing result
                addToHistory(tempResult.innerHTML);
                
                // Skip showing in result area since it will already be in history
                // Clear temp container immediately
                tempResult.innerHTML = '';
            } else {
                // For errors, show them in the result area and don't add to history
                const resultElement = document.getElementById('result');
                resultElement.innerHTML = tempResult.innerHTML;
                resultElement.classList.remove('hidden');
                tempResult.innerHTML = '';
                
                // Hide the error message after a delay
                setTimeout(() => {
                    resultElement.classList.add('hidden');
                    resultElement.innerHTML = '';
                }, 10000);
            }
        }
    });
});

// Function to paste URL from clipboard
function pasteFromClipboard() {
    navigator.clipboard.readText()
        .then(text => {
            document.getElementById('videoUrl').value = text;
            console.log("URL pasted from clipboard");
        })
        .catch(err => {
            console.error('Failed to read clipboard: ', err);
            alert('Unable to paste from clipboard. Please check permissions.');
        });
}

// Function to clear URL input field
function clearUrlField() {
    document.getElementById('videoUrl').value = '';
    console.log("URL field cleared");
}

// Function to clear all history
function clearHistory() {
    if (confirm('Are you sure you want to clear all download history?')) {
        localStorage.removeItem('downloadHistory');
        displayHistory();
        console.log("History cleared");
    }
}

// Function to copy text to clipboard
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(function() {
        alert('Link copied to clipboard!');
    }, function(err) {
        console.error('Could not copy text: ', err);
    });
}

// Make these functions globally accessible for direct HTML event handlers
// Must be done AFTER the functions are defined
window.clearHistory = clearHistory;
window.copyToClipboard = copyToClipboard;

// Helper to convert timestamp strings to Date objects
function getTimestampFromElement(element) {
    const timestampEl = element.querySelector('.timestamp');
    return timestampEl ? parseInt(timestampEl.getAttribute('data-time')) : 0;
}

// Function to add download result to history
function addToHistory(html) {
    // Create a temporary div to parse the HTML
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = html;
    
    // Check if this is a success message
    const successAlert = tempDiv.querySelector('.alert-success');
    if (!successAlert) return;  // Don't add errors to history
    
    // Get download info
    const videoInfo = tempDiv.querySelector('.video-info');
    if (!videoInfo) return;
    
    // Extract data from the HTML
    const url = videoInfo.querySelector('.download-link')?.textContent?.trim() || '';
    const title = videoInfo.querySelector('h3')?.textContent?.trim() || 'YouTube Video';
    const timestamp = Date.now();
    const timestampFormatted = new Date().toLocaleString();
    
    // Extract the quality setting from the URL
    const quality = url.match(/\/Q\d+P(?:h264Forced)?\//) ? 
        url.match(/\/Q\d+P(?:h264Forced)?\//).toString().replace(/\//g, '') : 
        document.getElementById('quality')?.value || 'Unknown';
    
    // Format quality for display
    let qualityDisplay = quality.replace('Q', '').replace('P', 'p');
    if (quality.includes('h264Forced')) {
        qualityDisplay = qualityDisplay.replace('h264Forced', ' (H264)');
    }
    
    // Calculate when entry will expire
    const maxAgeMs = maxVideoAge * 60 * 60 * 1000;
    
    // Create a clean history entry HTML with properly styled buttons
    const historyHTML = `\
        <div class="video-info">
            <h3>${title} <span class="quality-badge">${qualityDisplay}</span></h3>
            <div class="download-link">${url}</div>
            <div class="video-actions">
                <button class="btn btn-sm copy-btn" onclick="copyToClipboard('${url}')"><i class="fas fa-copy"></i> Copy Link</button>
                <a href="${url}" target="_blank" class="btn btn-sm open-btn"><i class="fas fa-external-link-alt"></i> Open Video</a>
            </div>
            <div class="timestamp" data-time="${timestamp}">
                ${timestampFormatted} 
                <span class="countdown" data-expires="${timestamp + maxAgeMs}"></span>
            </div>
        </div>
    `;
    
    // Create history entry object
    const historyEntry = {
        url,
        title,
        quality,
        timestamp,
        timestampFormatted,
        html: historyHTML
    };
    
    // Get existing history or create new array
    let history = JSON.parse(localStorage.getItem('downloadHistory') || '[]');
    
    // Check if this URL already exists in history to avoid duplicates
    const isDuplicate = history.some(item => item.url === url);
    if (!isDuplicate) {
        // Add new entry at the beginning (newest first)
        history.unshift(historyEntry);
        
        // Limit history to 50 items to avoid localStorage issues
        if (history.length > 50) {
            history = history.slice(0, 50);
        }
        
        // Save to localStorage
        localStorage.setItem('downloadHistory', JSON.stringify(history));
    }
    
    // Update the display
    displayHistory();
    
    // Log for debugging
    console.log("Added to history:", historyEntry);
}

// Function to clean expired entries from history
function cleanExpiredHistory() {
    // Convert maxVideoAge from hours to milliseconds
    const maxAgeMs = maxVideoAge * 60 * 60 * 1000;
    const currentTime = Date.now();
    
    // Get history from localStorage
    let history = JSON.parse(localStorage.getItem('downloadHistory') || '[]');
    
    if (history.length === 0) return;
    
    // Filter out expired entries
    const filteredHistory = history.filter(entry => {
        const entryAge = currentTime - entry.timestamp;
        return entryAge < maxAgeMs;
    });
    
    // If items were removed, update localStorage
    if (filteredHistory.length !== history.length) {
        console.log(`Removed ${history.length - filteredHistory.length} expired entries from history`);
        localStorage.setItem('downloadHistory', JSON.stringify(filteredHistory));
    }
}

// Function to display download history
function displayHistory() {
    const historySection = document.getElementById('history');
    
    // Clear existing content
    historySection.innerHTML = '';
    
    // Clean expired entries first
    cleanExpiredHistory();
    
    // Get history from localStorage
    const history = JSON.parse(localStorage.getItem('downloadHistory') || '[]');
    
    if (history.length === 0) {
        historySection.innerHTML = '<div class="empty-history">No download history yet</div>';
        return;
    }
    
    // Add each history item to the display
    history.forEach((entry, index) => {
        const entryElement = document.createElement('div');
        entryElement.classList.add('history-entry');
        entryElement.innerHTML = entry.html;
        historySection.appendChild(entryElement);
    });
}

// Function to update all countdown timers on the page
function updateCountdownTimers() {
    const countdowns = document.querySelectorAll('.countdown');
    const now = Date.now();
    
    countdowns.forEach(countdown => {
        const expiresTime = parseInt(countdown.getAttribute('data-expires'));
        const timeRemaining = expiresTime - now;
        
        if (timeRemaining <= 0) {
            // The entry has expired, it will be removed on next cleanExpiredHistory
            countdown.textContent = '(expired)';
            countdown.className = 'countdown danger';
            return;
        }
        
        // Calculate hours, minutes, seconds
        const hours = Math.floor(timeRemaining / (1000 * 60 * 60));
        const minutes = Math.floor((timeRemaining % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((timeRemaining % (1000 * 60)) / 1000);
        
        // Format the countdown text
        let countdownText = '';
        if (hours > 0) {
            countdownText = `(expires in ${hours}h ${minutes}m)`;
        } else if (minutes > 0) {
            countdownText = `(expires in ${minutes}m ${seconds}s)`;
        } else {
            countdownText = `(expires in ${seconds}s)`;
        }
        
        // Update the countdown text
        countdown.textContent = countdownText;
        
        // Color coding based on time remaining
        if (timeRemaining < 1000 * 60 * 30) { // Less than 30 minutes
            countdown.className = 'countdown danger';
        } else if (timeRemaining < 1000 * 60 * 60) { // Less than 1 hour
            countdown.className = 'countdown warning';
        } else {
            countdown.className = 'countdown';
        }
    });
}

// Function to setup theme toggle
function setupThemeToggle() {
    const checkbox = document.getElementById('checkbox');
    const themeLabel = document.getElementById('theme-label');
    
    // Check for saved theme preference or use default (dark)
    const savedTheme = localStorage.getItem('theme') || 'dark';
    
    // Apply the saved theme
    document.documentElement.setAttribute('data-theme', savedTheme);
    
    // Update checkbox and label based on saved theme
    if (savedTheme === 'dark') {
        checkbox.checked = true;
        themeLabel.textContent = 'Dark Mode';
    } else {
        checkbox.checked = false;
        themeLabel.textContent = 'Light Mode';
    }
    
    // Add event listener for theme toggle
    checkbox.addEventListener('change', function() {
        if (checkbox.checked) {
            document.documentElement.setAttribute('data-theme', 'dark');
            themeLabel.textContent = 'Dark Mode';
            localStorage.setItem('theme', 'dark');
        } else {
            document.documentElement.setAttribute('data-theme', 'light');
            themeLabel.textContent = 'Light Mode';
            localStorage.setItem('theme', 'light');
        }
    });
}