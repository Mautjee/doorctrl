// Time utility functions for Unix timestamp handling

function formatUnixTimestamp(unixTimestamp, format = 'datetime') {
    const date = new Date(unixTimestamp * 1000);
    
    const options24h = {
        hour12: false
    };
    
    switch (format) {
        case 'datetime':
            return date.toLocaleString('en-GB', { ...options24h, day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
        case 'date':
            return date.toLocaleDateString('en-GB', { day: '2-digit', month: '2-digit', year: 'numeric' });
        case 'time':
            return date.toLocaleTimeString('en-GB', { ...options24h, hour: '2-digit', minute: '2-digit' });
        case 'short':
            return date.toLocaleString('en-GB', {
                ...options24h,
                day: '2-digit',
                month: 'short',
                hour: '2-digit',
                minute: '2-digit'
            });
        case 'full':
            return date.toLocaleString('en-GB', {
                ...options24h,
                weekday: 'long',
                year: 'numeric',
                month: 'long',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        default:
            return date.toLocaleString('en-GB', options24h);
    }
}

function getUnixTimestamp(date = new Date()) {
    return Math.floor(date.getTime() / 1000);
}

function createUnixTimestampFromInputs(dateStr, timeStr) {
    const datetime = new Date(`${dateStr}T${timeStr}`);
    return Math.floor(datetime.getTime() / 1000);
}
