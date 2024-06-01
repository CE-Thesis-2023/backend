import { default as duration } from 'dayjs/plugin/duration';
import { axiosClient } from "./client";

export interface ObjectTrackingEvent {
    eventId: string;
    openGateEventId: string;
    eventType: string;
    cameraId: string;
    CameraName: string;
    frameTime: string;
    label: string;
    topScore: number;
    score: number;
    hasSnapshot: boolean;
    hasClip: boolean;
    stationary: boolean;
    falsePositive: boolean;
    startTime: string;
    endTime: string;
    snapshotId: string;
}

/**
 * Get object tracking events
 * @api GET /api/events/object_tracking
 * @param ids Event IDs
 * @returns Object tracking events
 */
export async function getObjectTrackingEvents(options:
    {
        ids: string[],
        cameraId: string,
        limit: number | undefined,
        within: number | undefined,
        latest: boolean | undefined
    }): Promise<ObjectTrackingEvent[]> {
    let uri = "/api/events/object_tracking";
    let queryCount = 0;
    if (options.ids.length > 0) {
        if (queryCount > 0) {
            uri += '&ids=' + options.ids.join(',');
        } else {
            uri += '?ids=' + options.ids.join(',');
        }
        queryCount++;
    }
    if (options.cameraId.length > 0 && options.cameraId != null) {
        if (queryCount > 0) {
            uri += '&camera_id=' + options.cameraId;
        } else {
            uri += '?camera_id=' + options.cameraId;
        }
        queryCount++;
    }
    if (options.limit != undefined) {
        if (queryCount > 0) {
            uri += '&limit=' + options.limit;
        } else {
            uri += '?limit=' + options.limit;
        }
        queryCount++;
    }
    if (options.within != undefined) {
        if (queryCount > 0) {
            uri += '&within=' + options.within + "s";
        } else {
            uri += '?within=' + options.within + "s";
        }
        queryCount++;
    }
    if (options.latest != undefined) {
        let str = options.latest ? 'true' : 'false';
        if (queryCount > 0) {
            uri += '&latest=' + str;
        } else {
            uri += '?latest=' + str;
        }
        queryCount++;
    }
    const resp = await axiosClient.get(uri);
    return resp.data["objectTrackingEvents"];
}

export interface EventSnapshot {
    snapshot: Snapshot[];
    presignedUrl: Object;
}

export interface Snapshot {
    snapshotId: string;
    timestamp: string;
    transcoderId: string;
    openGateEventId: string;
    detectedPeopleId: string | undefined;
}

/**
 * Get snapshots
 * @api GET /api/snapshots
 * @param snapshotIds Snapshot IDs
 * @returns Snapshots
 */
export async function getSnapshots(snapshotIds: string[]): Promise<EventSnapshot> {
    let uri = "/api/snapshots";
    if (snapshotIds.length > 0) {
        uri += '?snapshot_id=' + snapshotIds.join(',');
    }
    const resp = await axiosClient.get(uri);
    return resp.data;
}