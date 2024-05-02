import { axiosClient } from './client';

/**
 * Toggle stream
 * @api PUT /api/cameras/{cameraId}/streams
 * @param cameraId Camera ID
 * @param enabled Whether to enable/disable stream
 */
export async function toggleStream(cameraId: string, enabled: boolean): Promise<void> {
    const uri = `/api/cameras/${cameraId}/streams?enabled=${enabled}`;
    await axiosClient.put(uri, { enabled });
}

export interface StreamInfo {
    streamUrl: string;
    protocol: string;
    transcoderId: string;
    transcoderName: string;
    enabled: boolean;
}

/**
 * Get the stream info of a camera
 * @api GET /api/cameras/{cameraId}/streams
 * @param id Camera ID
 * @returns Camera's StreamInfo
 */
export async function getCameraStreamInfo(id: string): Promise<StreamInfo> {
    const uri = `/api/cameras/${id}/streams`;
    const resp = await axiosClient.get(uri);
    return resp.data;
}