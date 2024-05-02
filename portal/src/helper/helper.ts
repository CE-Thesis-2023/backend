import { Camera, getCameras } from "../clients/backend/cameras";
import { ObjectTrackingEvent, Snapshot, getObjectTrackingEvents, getSnapshots } from "../clients/backend/object";
import { OpenGateCameraSettings, OpenGateIntegration, getOpenGateCameraSettings, getOpenGateConfigurations } from "../clients/backend/opengate";
import { Person, PersonHistory, PersonImage, getPeople, getPeopleImage, getPersonHistory } from "../clients/backend/people";
import { StreamInfo, getCameraStreamInfo } from "../clients/backend/streams";
import { CameraStats, DetectorStats, Transcoder, TranscoderStatus, getCameraStats, getTranscoderStatus, getTranscoders } from "../clients/backend/transcoders";

export interface CameraAggregatedInfo {
    camera: Camera;
    transcoder: Transcoder;
    integration: OpenGateIntegration;
    settings: OpenGateCameraSettings;
    streamInfo: StreamInfo;
    transcoderStatus: TranscoderStatus;
}

/**
 * Gets all information about a camera
 * @param cameraId Camera ID
 */
export async function getCameraViewInfo(cameraId: string): Promise<CameraAggregatedInfo> {
    const cameras = await getCameras([cameraId]);
    if (cameras.length == 0) {
        throw Error("Camera not found");
    }
    const c = cameras[0];

    const transcoderId = c.transcoderId;
    const transcoders = await getTranscoders([transcoderId])
    if (transcoders.length == 0) {
        throw Error("Transcoder not found");
    }
    const t = transcoders[0];
    const integrationId = t.openGateIntegrationId;
    const integration = await getOpenGateConfigurations(integrationId);

    const settings = await getOpenGateCameraSettings([cameraId]);
    if (settings.length == 0) {
        throw Error("Camera settings not found");
    }
    const s = settings[0];

    const streamInfo = await getCameraStreamInfo(cameraId);

    const transcoderStatus = await getTranscoderStatus([transcoderId], [cameraId]);
    if (transcoderStatus.length == 0) {
        throw Error("Transcoder status not found");
    }
    const status = transcoderStatus[0];

    return {
        camera: c,
        transcoder: t,
        integration: integration,
        settings: s,
        streamInfo: streamInfo,
        transcoderStatus: status,
    }
}

export interface UpdatedInfo {
    events: Event[];
    stats: CameraStats;
    detectorStats: DetectorStats;
}

export interface Event {
    tracking: ObjectTrackingEvent;
    snapshot: Snapshot;
    presignedUrl: string;
}

/**
 * Get frequently updated information a camera
 * @param cameraId Camera ID
 * @param cameraName OpenGate camera name
 * @param transcoderId Transcoder ID
 * @returns 
 */
export async function getUpdatedInfo(cameraId: string, cameraName: string, transcoderId: string): Promise<UpdatedInfo> {
    const events = await getObjectTrackingEvents([], cameraId);
    let snapshotIds = events.map(e => e.snapshotId);
    const snapshots = await getSnapshots(snapshotIds);

    let snapshotMap: Map<string, Snapshot> = new Map<string, Snapshot>();
    snapshots.snapshot.forEach(s => snapshotMap.set(s.snapshotId, s))
    const presignedUrls = new Map<string, string>(Object.entries(snapshots.presignedUrl));

    const aggregatedEvent: Event[] = [];
    for (let i = 0; i < events.length; i += 1) {
        const e = events[i];
        aggregatedEvent[i] = {
            tracking: e,
            snapshot: snapshotMap.get(e.snapshotId)!,
            presignedUrl: presignedUrls.get(e.snapshotId)!,
        };
    }

    const stats = await getCameraStats(transcoderId, [cameraName]);
    const detectorStats = stats.detectorStats[0];
    const cameraStats = stats.cameraStats[0];

    return {
        events: aggregatedEvent,
        stats: cameraStats,
        detectorStats: detectorStats,
    };
}

export interface PersonInfo {
    person: Person;
    image: PersonImage;
    history: PersonHistory[];
}

/**
 * Get person information
 * @param personId Person ID
 * @returns 
 */
export async function getPersonInfo(personId: string): Promise<PersonInfo> {
    const people = await getPeople([personId]);
    if (people.length == 0) {
        throw Error("Person not found");
    }
    const image = await getPeopleImage(personId);
    const history = await getPersonHistory([personId]);
    return {
        person: people[0],
        image: image,
        history: history,
    }
}