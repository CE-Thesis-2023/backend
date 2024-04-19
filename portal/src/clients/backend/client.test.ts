import { getCameras, getTranscoders } from "./client";

test('get cameras', async () => {
    expect(async () => {
        await getCameras([]);
    }).not.toThrow();
})

test('get devices', async () => {
    expect(async () => {
        await getTranscoders([]);
    }).not.toThrow();
})