import { getCameraViewInfo, getListCameras, getListTranscoders, getUpdatedInfo } from "./helper";

test("camera aggregated info", async () => {
    const cameraId = "ec92cdfa-a7a0-4b4c-8717-bfb26753cc5d";
    const result = await getCameraViewInfo(cameraId);
    expect(result.camera.cameraId).toBe(cameraId);
    console.log(result);
});

test("get camera updated info", async () => {
    const cameraId = "ec92cdfa-a7a0-4b4c-8717-bfb26753cc5d";
    const cameraName = "ip_camera_03";
    const transcoderId = "test-device-01";
    const result = await getUpdatedInfo({
        cameraId,
        cameraName,
        transcoderId
    });
    expect(result.events.length).toBeGreaterThan(0);
    console.log(JSON.stringify(result));
})

test("get camera list", async () => {
    const cameraIds: string[] = [];
    const result = await getListCameras(cameraIds);
    expect(result.items.length).toBeGreaterThan(0);
    console.log(JSON.stringify(result));
})

test("get transcoder list", async () => {
    const transcoderIds: string[] = [];
    const result = await getListTranscoders(transcoderIds);
    expect(result.items.length).toBeGreaterThan(0);
    console.log(JSON.stringify(result));
})