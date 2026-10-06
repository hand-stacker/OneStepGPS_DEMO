

function getDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
    return Math.sqrt(Math.pow(lat1 - lat2, 2) + Math.pow(lon1 - lon2, 2))
}
// Finds the lat and long of the device which is closest to all other devices.
// I know its O(n^2) but it should be fine for the small number of demo devices prvided.
function friendliestNodePos(devices: any[]): any {
    let bestDevice;
    let minTotalDistance = Infinity;

    for (const device of devices) {
        let totalDistance = 0;
        for (const otherDevice of devices) {
            if (device !== otherDevice) {
                totalDistance += getDistance(
                    Number(device.lat),
                    Number(device.lng),
                    Number(otherDevice.lat),
                    Number(otherDevice.lng)
                );
            }
        }
        if (totalDistance < minTotalDistance) {
            minTotalDistance = totalDistance;
            bestDevice = device;
        }
    }

    return {'lat': bestDevice.lat, 'lng': bestDevice.lng};
}

// UTC strings like "2026-10-05T19:47:19Z" shown in Pacific time (PDT/PST)
function formatPacificTime(utc: string): string {
    if (!utc) return ''
    const date = new Date(utc)
    if (isNaN(date.getTime())) return utc
    // dateStyle/timeStyle can't be combined with timeZoneName, so spell the fields out
    return date.toLocaleString('en-US', {
        timeZone: 'America/Los_Angeles',
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit',
        second: '2-digit',
        timeZoneName: 'short',
    })
}

// time between a UTC string (or ms timestamp) and now, e.g. "1h 02m 05s"
// now is passed in so callers can tick it to keep the text live
function elapsedSince(start: string | number, now: number): string {
    if (!start) return ''
    const startMs = typeof start === 'number' ? start : new Date(start).getTime()
    const totalSeconds = Math.max(0, Math.floor((now - startMs) / 1000))
    const days = Math.floor(totalSeconds / 86400)
    const hours = Math.floor((totalSeconds % 86400) / 3600)
    const minutes = Math.floor((totalSeconds % 3600) / 60)
    const seconds = totalSeconds % 60
    const pad = (n: number) => String(n).padStart(2, '0')
    if (days > 0) return `${days}d ${hours}h ${pad(minutes)}m`
    if (hours > 0) return `${hours}h ${pad(minutes)}m ${pad(seconds)}s`
    if (minutes > 0) return `${minutes}m ${pad(seconds)}s`
    return `${seconds}s`
}

function formatFuelPercentage(str : string): string {
    if (str) {
        var f = parseFloat(str).toFixed(2)
        return String(f) + "%"
    }
    return "Data Unavailable"
}

function formatSpeed(f : Number) {
    return String(f) + " mph"
}

export { friendliestNodePos, formatPacificTime, elapsedSince, formatFuelPercentage, formatSpeed }