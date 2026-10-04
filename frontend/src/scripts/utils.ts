

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

export {friendliestNodePos }