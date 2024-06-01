import { axiosClient } from "./client";

export interface Transcoder {
    deviceId: string;
    name: string;
    openGateIntegrationId: string;
}

/**
 * Get transcoders or LTDs
 * @api GET /api/devices
 * @param ids 
 * @returns List of Transcoders 
 */
export async function getTranscoders(ids: string[]): Promise<Transcoder[]> {
    let uri = '/api/devices';
    if (ids.length > 0) {
        uri += '?id=' + ids.join(',');
    }
    const resp = await axiosClient.get(uri);
    return resp.data["transcoders"];
}

export interface UpdateTranscoder {
    id: string;
    name: string;
    logLevel: string;
    hardwareAccelerationType: string;
    edgeTpuEnabled: boolean;
}

/**
 * Get transcoders
 * @api PUT /api/devices
 * @param transcoder Changes to make
 * @returns 
 */
export async function updateTranscoder(transcoder: UpdateTranscoder): Promise<void> {
    const uri = `/api/devices`
    await axiosClient.put(uri, transcoder);
    return;
}

export interface TranscoderStatus {
    statusId: string;
    transcoderId: string;
    cameraId: string;
    objectDetection: boolean;
    audioDetection: boolean;
    openGateRecordings: boolean;
    snapshots: boolean;
    motionDetection: boolean;
    improveContrast: boolean;
    autotracker: boolean;
    birdseyeView: boolean;
    openGateStatus: boolean;
    transcoderStatus: boolean;
}

/**
 * Get transcoder status
 * @api GET /api/devices/status
 * @param ids Transcoder IDs
 * @returns Transcoder status
 */
export async function getTranscoderStatus(transcoderIds: string[], cameraIds: string[]): Promise<TranscoderStatus[]> {
    const uri = `/api/devices/status?transcoder_id=${transcoderIds.join(',')}&camera_id=${cameraIds.join(',')}`;
    const resp = await axiosClient.get(uri);
    return resp.data["status"];
}

/**
 * Perform device healthcheck
 * @api GET /api/devices/{transcoderId}/healthcheck
 * @param transcoderId Transcoder ID
 * @returns 
 */
export async function doDeviceHealthcheck(transcoderId: string): Promise<void> {
    const uri = `/api/devices/${transcoderId}/healthcheck`;
    await axiosClient.get(uri);
    return;
}

export interface CameraStats {
    cameraStatId: string;
    transcoderId: string;
    cameraName: string;
    cameraFps: string;
    detectionFps: string;
    capturePid: number;
    processId: number;
    processFps: string;
    skippedFps: string;
    timestamp: string;
}

export interface DetectorStats {
    detectorStatId: string;
    detectorName: string;
    transcoderId: string;
    detectorStart: number;
    inferenceSpeed: number;
    processId: number;
    timestamp: string;
}

export interface Stats {
    cameraStats: CameraStats;
    detectorStats: DetectorStats;
}

/**
 * Get camera stats
 * @api GET /api/stats
 * @param transcoderId Transcoder ID
 * @param cameraNames OpenGate camera names
 * @returns 
 */
export async function getCameraStats(transcoderId: string, cameraNames: string[]): Promise<Stats> {
    const uri = `/api/stats?transcoder_id=${transcoderId}&camera_name=${cameraNames.join(',')}`;
    const resp = await axiosClient.get(uri);
    return resp.data;
}
