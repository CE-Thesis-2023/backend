import { axiosClient, privateClient } from "./client";

export interface OpenGateIntegration {
    openGateId: string
    logLevel: string
    snapshotRetentionDays: number
    hardwareAccelerationType: string
    withEdgeTpu: boolean
    mqttId: string
    transcoderId: string
}

/**
 * Get OpenGate configurations
 * @api GET /api/opengate/{openGateId}
 * @param id OpenGateID
 * @returns OpenGate integration configuration
 */
export async function getOpenGateConfigurations(id: string): Promise<OpenGateIntegration> {
    const uri = `/api/opengate/${id}`;
    const resp = await axiosClient.get(uri);
    return resp.data["openGateIntegration"];
}

export interface OpenGateCameraSettings {
    settingsId: string;
    height: number;
    width: number;
    fps: number;
    mqttEnabled: boolean;
    timestamp: boolean;
    boundingBox: boolean;
    crop: boolean;
    openGateId: string;
    cameraId: string;
}

/**
 * Get OpenGate camera settings
 * @api GET /private/opengate/cameras
 * @param id Camera ID
 * @returns OpenGate camera settings
 */
export async function getOpenGateCameraSettings(ids: string[]): Promise<OpenGateCameraSettings[]> {
    let uri = "/private/opengate/cameras";
    if (ids.length > 0) {
        uri += '?camera_id=' + ids.join(',');
    }
    const resp = await privateClient.get(uri);
    return resp.data["openGateCameraSettings"];
}