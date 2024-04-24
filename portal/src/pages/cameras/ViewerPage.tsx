// @ts-nocheck

import { useParams } from "@solidjs/router";
import { ArrowDownward, ArrowLeft, ArrowRight, ArrowUpward, Refresh, Visibility } from "@suid/icons-material";
import { Button, Chip, CircularProgress, Divider, FormControl, FormControlLabel, IconButton, Input, InputAdornment, InputLabel, List, ListItemButton, ListItemText, Paper, Switch as SWButton, Typography } from "@suid/material";
import dayjs from "dayjs";
import { Component, For, Match, Switch, createResource } from "solid-js";
import { ObjectTrackingEvent, Snapshot, getCameraStreamInfo, getCameras, getObjectTrackingEvents, getSnapshots } from "../../clients/backend/client";

async function fetchCameraData(cameraId: string) {
    const response = await getCameras([cameraId]);
    const camera = response[0];
    const streamInfo = await getCameraStreamInfo(cameraId);
    return {
        camera: camera,
        streamInfo: streamInfo
    };
}

interface CameraEvent {
    event: ObjectTrackingEvent;
    snapshot: Snapshot;
    presignedUrl: string;
}

async function fetchCameraEvents(cameraId: string) {
    const events = await getObjectTrackingEvents([]);
    let snapshotIds = []
    for (let i = 0; i < events.length; i++) {
        snapshotIds.push(events[i].snapshotId);
    }
    const snapshots = await getSnapshots(snapshotIds);
    const snapshotMap = new Map<string, Snapshot>();
    for (let i = 0; i < snapshots.snapshot.length; i++) {
        snapshotMap.set(
            snapshots.snapshot[i].snapshotId,
            snapshots.snapshot[i]);
    }
    const eventsWithSnapshots: CameraEvent[] = [];
    for (let i = 0; i < events.length; i++) {
        const snapshot = snapshotMap.get(events[i].snapshotId);
        if (snapshot) {
            console.log(snapshots.presignedUrl)
            eventsWithSnapshots.push({
                event: events[i],
                snapshot: snapshot,
                presignedUrl: snapshots.
                    presignedUrl[snapshot.snapshotId]
            });
        }
    }
    return eventsWithSnapshots;
}

export const CameraViewerPage: Component = () => {
    const routeParams = useParams();
    const [data, { refetch }] = createResource(routeParams.cameraId, fetchCameraData);
    const [events, { refetch: eventRefetch }] = createResource(routeParams.cameraId, fetchCameraEvents);

    return <Switch>
        <Match when={data.loading}>
            <div class="flex flex-row justify-center mt-8">
                <CircularProgress />
            </div>
        </Match>
        <Match when={data.error}>Error: {data.error.message}</Match>
        <Match when={data()}>
            <div class="flex flex-row h-full w-full">
                <div class="w-full h-full" style="flex: 7">
                    <iframe src={data()!.streamInfo.streamUrl} width={"100%"} height={"100%"} allowfullscreen />
                </div>
                <div class="w-full h-full p-4 flex flex-col gap-4" style="flex: 3">
                    <Paper sx={{ width: '100%', height: "fit-content", padding: '1rem' }}>
                        <div class="flex w-full h-fit flex-row justify-between items-center">
                            <div>
                                <Typography variant="body1">{data()!.camera.name}</Typography>
                                <FormControlLabel
                                    label="Enabled"
                                    control={<SWButton size="small" inputProps={{ "aria-label": "controlled" }} defaultChecked />}
                                />
                            </div>
                            <div class="flex flex-row justify-end align-middle gap-2">
                                <IconButton size="small" aria-label="camera-info-button">
                                    <Visibility />
                                </IconButton>
                                <Button size="small" onClick={() => eventRefetch()} startIcon={<Refresh />} variant="contained">
                                    {"Refresh"}
                                </Button>
                            </div>
                        </div>
                    </Paper>
                    <Paper sx={{ width: '100%', height: "100%" }}>
                        <List>
                            <For each={events()}>
                                {event => <>
                                    <EventItem event={event} />
                                    <Divider />
                                </>}
                            </For>
                        </List>
                    </Paper>
                    <Paper sx={{ width: '100%', height: "fit-content", padding: '1rem' }} class="flex flex-row gap-4">
                        <div class="flex flex-row gap-4 flex-1">
                            <Button variant="contained" sx={{ width: '1rem' }} color="primary"><ArrowLeft /></Button>
                            <div class="flex flex-col gap-4 justify-between items-center">
                                <Button variant="contained" fullWidth color="primary"><ArrowUpward /></Button>
                                <Button variant="contained" fullWidth color="primary"><ArrowDownward /></Button>
                            </div>
                            <Button variant="contained" color="primary"><ArrowRight /></Button>
                        </div>
                        <div class="flex-1">
                            <FormControl fullWidth variant="standard" size="small" margin="dense">
                                <InputLabel for="standard-adornment-duration">Reset after</InputLabel>
                                <Input
                                    id="standard-adornment-duration"
                                    endAdornment={<InputAdornment position="end">seconds</InputAdornment>}
                                />
                            </FormControl>
                        </div>
                    </Paper>
                </div>
            </div>
        </Match>
    </Switch>
}

const EventItem: Component<{ event: CameraEvent }> = (props) => {
    console.log(props.event);
    return <>
        <ListItemButton>
            <ListItemText primary={
                <div class="flex flex-row justify-between items-center">
                    {"Person detected"}
                    <div class="flex flex-row justify-start items-center">
                        {props.event.event.label === "person" ? <Chip label="Person" /> : <Chip label="Object" color="secondary" />}
                    </div>
                </div>
            } secondary={
                <div class="flex flex-row justify-between items-center">
                    <Typography variant="body2">
                        {`Score: ${Math.round(props.event.event.score * 100) / 100}`}
                    </Typography>
                    {`${dayjs(props.event.event.frameTime).format("H:mm:ss A on MMM DD, YYYY")}`}
                </div>
            } />
        </ListItemButton>
    </>
}