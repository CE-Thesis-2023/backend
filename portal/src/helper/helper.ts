import { Camera, getCameras } from "../clients/backend/cameras";
import { ObjectTrackingEvent, getObjectTrackingEvents } from "../clients/backend/object";
import { OpenGateCameraSettings, OpenGateIntegration, getOpenGateCameraSettings, getOpenGateConfigurations } from "../clients/backend/opengate";
import { StreamInfo, getCameraStreamInfo } from "../clients/backend/streams";
import { CameraStats, Transcoder, TranscoderStatus, getCameraStats, getTranscoderStatus, getTranscoders } from "../clients/backend/transcoders";

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

    const settingsId = c.settingsId;
    const settings = await getOpenGateCameraSettings([settingsId]);
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

/**
 * Get frequently updated information of cameras
 */
export interface UpdatedInfo {
    events: ObjectTrackingEvent[];
    stats: CameraStats[];
}

export async function getUpdatedInfo(cameraI): Promise<UpdatedInfo> {
    const events = await getObjectTrackingEvents();
    const stats = await getCameraStats();
    return { events, stats };
}