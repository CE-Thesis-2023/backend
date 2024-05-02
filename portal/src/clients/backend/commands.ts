import { axiosClient } from "./client";

export interface RemoteControl {
    cameraId: string;
    pan: number;
    tilt: number;
}

/**
 * PTZ remote control a camera
 * @api POST /api/rc
 * @param rc Remote control parameters
 * @returns 
 */
export async function remoteControl(rc: RemoteControl): Promise<void> {
    await axiosClient.post('/api/rc', rc);
    return;
}

export interface ISAPIDeviceInfo {
    cameraId: string;
    deviceName: string;
    deviceLocation: string;
    model: string;
    serialNumber: string;
    firmwareVersion: string;
    firmwareReleasedDate: string;
    capacity: number;
    usedCapacity: number;
    status: DeviceStatus;
}

export interface DeviceStatus {
    status: string;
    detailAbnormalStatus: CameraAbnormality;
}

export interface CameraAbnormality {
    hardDiskFull: boolean;
    hardDiskError: boolean;
    ethernetBroken: boolean;
    ipAddrConflict: boolean;
    illegalAccess: boolean;
    recordError: boolean;
    raidLogicDiskError: boolean;
    spareWorkDeviceError: boolean;
}

/**
 * Get device information
 * @api GET /api/cameras/info/{cameraId}
 * @param cameraId Camera ID
 * @returns Device information
 */
export async function getDeviceInfo(cameraId: string): Promise<ISAPIDeviceInfo> {
    const uri = `/api/cameras/info/${cameraId}`;
    const resp = await axiosClient.get(uri);
    return resp.data;
}