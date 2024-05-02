import { axiosClient } from "./client";

export interface Camera {
    cameraId: string;
    name: string;
    ip: string;
    port: number;
    username: string;
    password: string;
    enabled: boolean;
    openGateCameraName: string;
    groupId: string;
    transcoderId: string;
    settingsId: string;
}

/**
 * Get cameras
 * @api GET /api/cameras
 * @param ids 
 * @returns List of Cameras
 */
export async function getCameras(ids: string[]): Promise<Camera[]> {
    let uri = '/api/cameras';
    if (ids.length > 0) {
        uri += '?id=' + ids.join(',');
    }
    const resp = await axiosClient.get(uri);
    return resp.data["cameras"];
}

export interface AddCameraParams {
    name: string;
    ip: string;
    port: number;
    username: string;
    password: string;

    transcoderId: string;
}

/**
 * Add camera
 * @api POST /api/cameras
 * @param camera Camera parameters
 */
export async function addCamera(camera: AddCameraParams): Promise<void> {
    await axiosClient.post('/api/cameras', camera);
}

/**
 * Delete camera
 * @api DELETE /api/cameras
 * @param id Camera ID
 */
export async function deleteCamera(id: string): Promise<void> {
    const uri = `/api/cameras?id=${id}`;
    await axiosClient.delete(uri);
}