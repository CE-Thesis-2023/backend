import { getCameras, getObjectTrackingEvents, getPeople, getTranscoders } from "./client";

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

test('get people', async () => {
    expect(async () => {
        await getPeople([]);
    }).not.toThrow();
})

test('get object tracking events', async () => {
    expect(async () => {
        await getObjectTrackingEvents([]);
    }).not.toThrow();
})