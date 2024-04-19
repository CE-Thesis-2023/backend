import axios from 'axios';
import configs from '../../../dev.configs.json';

const axiosClient = axios.create({
    baseURL: configs.backendBaseUrl,
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }
})

interface Transcoder {
    deviceId: string;
    name: string;
    openGateIntegrationId: string;
}

/**
 * Get transcoders or LTDs
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

interface Camera {
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

interface StreamInfo {
    streamUrl: string;
    protocol: string;
    transcoderId: string;
    transcoderName: string;
    enabled: boolean;
}

/**
 * Get the stream info of a camera
 * @param id Camera ID
 * @returns Camera's StreamInfo
 */
export async function getCameraStreamInfo(id: string): Promise<StreamInfo> {
    const uri = `/api/cameras/${id}/streams`;
    const resp = await axiosClient.get(uri);
    return resp.data;
}

interface CameraGroup {
    groupId: string;
    name: string;
    createdDate: string;
}

/**
 * Get camera groups
 * @param ids List of CameraGroup IDs
 * @returns List of CameraGroups
 */
export async function getCameraGroup(ids: string[]): Promise<CameraGroup[]> {
    let uri = '/api/groups';
    if (ids.length > 0) {
        uri += '?ids=' + ids.join(',');
    }
    const resp = await axiosClient.get(uri);
    return resp.data["cameraGroups"];
}
