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
    stats: CameraStats | undefined;
    detectorStats: DetectorStats | undefined;
}

export interface Event {
    tracking: ObjectTrackingEvent;
    snapshot: Snapshot;
    presignedUrl: string;
}

export interface UpdatedInfoParams {
    cameraId: string,
    cameraName: string,
    transcoderId: string
}

/**
 * Get frequently updated information a camera
 * @param cameraId Camera ID
 * @param cameraName OpenGate camera name
 * @param transcoderId Transcoder ID
 * @returns 
 */
export async function getUpdatedInfo(params: UpdatedInfoParams, limit: number): Promise<UpdatedInfo> {
    const events = await getObjectTrackingEvents([], params.cameraId, limit);
    const aggregatedEvent: Event[] = [];
    if (events.length > 0) {
        let snapshotIds = events.map(e => e.snapshotId);
        const snapshots = await getSnapshots(snapshotIds);

        let snapshotMap: Map<string, Snapshot> = new Map<string, Snapshot>();
        snapshots.snapshot.forEach(s => snapshotMap.set(s.snapshotId, s))
        const presignedUrls = new Map<string, string>(Object.entries(snapshots.presignedUrl));

        for (let i = 0; i < events.length; i += 1) {
            const e = events[i];
            aggregatedEvent[i] = {
                tracking: e,
                snapshot: snapshotMap.get(e.snapshotId)!,
                presignedUrl: presignedUrls.get(e.snapshotId)!,
            };
        }
    }


    const stats = await getCameraStats(params.transcoderId, [params.cameraName]);
    const detectorStats = stats.detectorStats.length > 0 ?
        stats.detectorStats[0] :
        undefined;
    const cameraStats = stats.cameraStats.length > 0 ?
        stats.cameraStats[0] :
        undefined;
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

export interface PeopleList {
    items: FrontPagePersonInfo[];
}

export interface FrontPagePersonInfo {
    person: Person;
    image: PersonImage;
}

/**
 * Get list of people
 * @param personIds Person IDs
 * @returns 
 */
export async function getListPeople(personIds: string[]): Promise<PeopleList> {
    const people = await getPeople(personIds);
    const peopleIds = people.map(p => p.personId);
    const imagesAsync = peopleIds.map(id => getPeopleImage(id));
    const images = await Promise.all(imagesAsync);
    const imageMap = new Map<string, PersonImage>();
    for (let i = 0; i < peopleIds.length; i += 1) {
        imageMap.set(peopleIds[i], images[i]);
    }
    const frontPageInfo = people.map(p => {
        return {
            person: p,
            image: imageMap.get(p.personId)!,
        }
    });

    return {
        items: frontPageInfo,
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
    let openGateSettings: OpenGateCameraSettings[] = [];
    if (foundCameraIds.length > 0) {
        openGateSettings = await getOpenGateCameraSettings(foundCameraIds);
    }

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

export interface ListTranscoders {
    items: TranscoderInfo[];
}

export interface TranscoderInfo {
    transcoder: Transcoder;
    integration: OpenGateIntegration;
}

/**
 * Get list of transcoders
 * @param transcoderIds Transcoder IDs
 * @returns 
 */
export async function getListTranscoders(transcoderIds: string[]): Promise<ListTranscoders> {
    const transcoders = await getTranscoders(transcoderIds);
    const foundIds = transcoders.map(t => t.deviceId);

    const integrationMap = new Map<string, OpenGateIntegration>();
    for (let i = 0; i < foundIds.length; i += 1) {
        const id = foundIds[i];
        const integration = await getOpenGateConfigurations(transcoders[i].openGateIntegrationId);
        integrationMap.set(id, integration);
    }

    const infos: TranscoderInfo[] = transcoders.map(t => {
        return {
            transcoder: t,
            integration: integrationMap.get(t.deviceId)!,
        }
    })

    return {
        items: infos,
    }
}

export interface SummarizedHistory {
    history: PersonHistory;
    event: ObjectTrackingEvent;
    snapshot: string;
    person: Person;
}

/**
 * Get summarized person history
 * @param personId Person ID
 * @returns 
 */
export async function getSummarizedPersonHistory(personId: string): Promise<SummarizedHistory[]> {
    const people = await getPeople([personId]);
    if (people.length == 0) {
        throw Error("Person not found");
    }
    let summarized: SummarizedHistory[] = [];

    const history = await getPersonHistory([personId]);
    const eventsId = history.map(h => h.eventId);
    const events = await getObjectTrackingEvents(eventsId, "", undefined);
    const person = people[0];
    const eventsMap = new Map<string, ObjectTrackingEvent>();
    for (let i = 0; i < events.length; i += 1) {
        eventsMap.set(events[i].snapshotId, events[i]);
    }

    const snapshotIds: string[] = [];
    events.forEach(e => {
        if (e.snapshotId != undefined) {
            snapshotIds.push(e.snapshotId);
        }
    });
    const snapshots = await getSnapshots(snapshotIds);
    const snapshotMap = new Map<string, Snapshot>();
    snapshots.snapshot.forEach(s => snapshotMap.set(s.snapshotId, s));

    const presignedUrl = snapshots.presignedUrl;
    const presignedMap = new Map<string, string>(Object.entries(presignedUrl));
    history.forEach(h => {
        const snapshot = snapshotMap.get(h.eventId);
        if (snapshot == undefined) {
            return;
        }
        const event = eventsMap.get(snapshot?.snapshotId);
        if (event == undefined) {
            return;
        }
        summarized.push({
            history: h,
            event: event,
            person: person,
            snapshot: presignedMap.get(event.snapshotId) ?? "",
        });
    });

    return summarized;
}

export interface SummarizedEvent {
    event: ObjectTrackingEvent;
    snapshot: Snapshot;
    presignedUrl: string;
    person: Person | undefined;
}

/**
 * Get events
 * @param eventIds Event Ids
 * @returns 
 */
export async function getListEvents(eventIds: string[], limit: number): Promise<SummarizedEvent[]> {
    const events = await getObjectTrackingEvents(eventIds, "", limit);
    const snapshotIds = events.map(e => e.snapshotId);
    const snapshots = await getSnapshots(snapshotIds);
    const snapshotMap = new Map<string, Snapshot>();
    snapshots.snapshot.forEach(s => snapshotMap.set(s.snapshotId, s));

    const presignedUrl = snapshots.presignedUrl;
    const presignedMap = new Map<string, string>(Object.entries(presignedUrl));


    const personIds: string[] = [];
    snapshots.snapshot.forEach(s => {
        if (s.detectedPeopleId !== undefined) {
            personIds.push(s.detectedPeopleId)
        }
    });
    const people = await getPeople(personIds);
    const peopleMap = new Map<string, Person>();
    for (let i = 0; i < people.length; i += 1) {
        peopleMap.set(people[i].personId, people[i]);
    }

    const summarized: SummarizedEvent[] = events.map(e => {
        const snapshot = snapshotMap.get(e.snapshotId)!
        return {
            event: e,
            snapshot: snapshot,
            presignedUrl: presignedMap.get(e.snapshotId) ?? "",
            person: snapshot.detectedPeopleId ? peopleMap.get(snapshot.detectedPeopleId) : undefined,
        }
    });

    return summarized;
}

export interface SummarizedTranscoderInfo {
    transcoder: Transcoder;
    status: TranscoderStatus;
    integration: OpenGateIntegration;
}
