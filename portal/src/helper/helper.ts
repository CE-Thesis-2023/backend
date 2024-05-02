import { Camera, getCameras } from "../clients/backend/cameras";
import { RemoteControl, remoteControl } from "../clients/backend/commands";
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


export enum PTZDirection {
    Up = 1,
    Down,
    Left,
    Right,
}

/**
 * Perform remote control of camera
 * @param dir Direction to do PTZ Control
 * @param cameraId Camera ID
 * @returns 
 */
export async function doPtzCtrl(dir: PTZDirection, cameraId: string): Promise<void> {
    let rc: RemoteControl = {
        cameraId: cameraId,
        pan: 0,
        tilt: 0,
    };
    switch (dir) {
        case PTZDirection.Up:
            rc.pan = 0
            rc.tilt = 30
        case PTZDirection.Down:
            rc.pan = 0
            rc.tilt = -30
        case PTZDirection.Left:
            rc.pan = -30
            rc.tilt = 0
        case PTZDirection.Right:
            rc.pan = 30
            rc.tilt = 0
    }
    await remoteControl(rc);
    return;
}

export interface AggregatedListCamera {
    items: CameraItem[];
}

export interface CameraItem {
    camera: Camera;
    transcoder: Transcoder | undefined;
    settings: OpenGateCameraSettings;
    configs: OpenGateIntegration;
}

/**
 * Get camera information for list camera page
 * @param cameraIds Camera IDs
 * @returns 
 */
export async function getListCameras(cameraIds: string[]): Promise<AggregatedListCamera> {
    const cameras = await getCameras(cameraIds);
    const transcoderIds = cameras.map(camera => camera.transcoderId);

    const transcoders = await getTranscoders(transcoderIds);
    const transcoderMap = new Map<string, Transcoder>();
    for (let i = 0; i < transcoders.length; i++) {
        transcoderMap.set(transcoders[i].deviceId, transcoders[i]);
    }

    const foundCameraIds = cameras.map(c => c.cameraId);
    const openGateSettings = await getOpenGateCameraSettings(foundCameraIds);

    const settingsMap = new Map<string, OpenGateCameraSettings>();
    for (let i = 0; i < openGateSettings.length; i += 1) {
        settingsMap.set(cameras[i].cameraId, openGateSettings[i]);
    }

    const openGateConfigsMap = new Map<string, OpenGateIntegration>();
    for (let i = 0; i < transcoders.length; i += 1) {
        const configs = await getOpenGateConfigurations(transcoders[i].openGateIntegrationId);
        openGateConfigsMap.set(transcoders[i].deviceId, configs);
    }

    const aggregated: CameraItem[] = cameras.map(camera => {
        const ltd = transcoderMap.get(camera.transcoderId);
        const settings = settingsMap.get(camera.cameraId);
        const configs = openGateConfigsMap.get(camera.transcoderId);
        return {
            camera: camera,
            transcoder: ltd,
            settings: settings!,
            configs: configs!,
        };
    });

    return {
        items: aggregated,
    };
}