import axios from 'axios';
import configs from '../../../dev.configs.json';

const axiosClient = axios.create({
    baseURL: configs.backendBaseUrl[0],
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }
})

const privateClient = axios.create({
    baseURL: configs.backendBaseUrl[1],
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'Authorization': 'Basic ZGV2OmRldg=='
    }
})

export interface Transcoder {
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

export interface StreamInfo {
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

export interface CameraGroup {
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

/**
 * Delete camera
 * @param id Camera ID
 */
export async function deleteCamera(id: string): Promise<void> {
    const uri = `/api/cameras?id=${id}`;
    await axiosClient.delete(uri);
}

export interface OpenGateIntegration {
    openGateId: string
    available: boolean
    isRestarting: boolean
    logLevel: string
    snapshotRetentionDays: number
    mqttId: string
    transcoderId: string
}

/**
 * Get OpenGate configurations
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

interface Person {
    personId: string;
    name: string;
    age: string;
    imagePath: string;
}

/**
 * Get detectable people
 * @param ids Person IDs
 * @returns People
 */
export async function getPeople(ids: string[]): Promise<Person[]> {
    let uri = "/api/people";
    if (ids.length > 0) {
        uri += '?ids=' + ids.join(',');
    }
    const resp = await axiosClient.get(uri);
    return resp.data["people"];
}

interface ObjectTrackingEvent {
    eventId: string;
    openGateEventId: string;
    eventType: string;
    cameraId: string;
    cameraName: string;
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
}

/**
 * Get object tracking events
 * @param ids Event IDs
 * @returns Object tracking events
 */
export async function getObjectTrackingEvents(ids: string[]): Promise<ObjectTrackingEvent[]> {
    let uri = "/api/events/object_tracking";
    if (ids.length > 0) {
        uri += '?ids=' + ids.join(',');
    }
    const resp = await axiosClient.get(uri);
    return resp.data["objectTrackingEvents"];
}