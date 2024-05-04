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
    if (ids.length > 0) {
        uri += '?ids=' + ids.join(',');
    }
    if (cameraId.length > 0 && cameraId != null) {
        uri += '?camera_id=' + cameraId;
    }
    if (limit != null) {
        uri += '?limit=' + limit;
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