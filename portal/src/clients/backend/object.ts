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
export async function getObjectTrackingEvents(ids: string[], cameraId: string, limit: number | undefined): Promise<ObjectTrackingEvent[]> {
    let uri = "/api/events/object_tracking";
    let queryCount = 0;
    if (ids.length > 0) {
        if (queryCount > 0) {
            uri += '&ids=' + ids.join(',');
        } else {
            uri += '?ids=' + ids.join(',');
        }
        queryCount++;
    }
    if (cameraId.length > 0 && cameraId != null) {
        if (queryCount > 0) {
            uri += '&camera_id=' + cameraId;
        } else {
            uri += '?camera_id=' + cameraId;
        }
        queryCount++;
    }
    if (limit != null) {
        if (queryCount > 0) {
            uri += '&limit=' + limit;
        } else {
            uri += '?limit=' + limit;
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