
import { useParams } from "@solidjs/router";
import { ArrowDownward, ArrowLeft, ArrowRight, ArrowUpward, Refresh, Visibility } from "@suid/icons-material";
import { Box, Button, Chip, CircularProgress, Divider, FormControl, FormControlLabel, IconButton, Input, InputAdornment, InputLabel, List, ListItemAvatar, ListItemButton, ListItemText, Modal, Paper, Switch as SWButton, Typography } from "@suid/material";
import dayjs from "dayjs";
import { Component, For, Match, Switch, createResource, createSignal } from "solid-js";
import { ObjectTrackingEvent, Snapshot, getCameraStreamInfo, getCameras, getObjectTrackingEvents, getPeople, getPeopleImage, getSnapshots } from "../../clients/backend/client";

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

    const [modalOpen, setModalOpen] = createSignal(false);
    const handleOpen = () => setModalOpen(true);
    const handleClose = () => setModalOpen(false);
    const [currentItem, setCurrentItem] = createSignal<CameraEvent | null>(null);

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
                                    <EventItem event={event} onClick={(item: CameraEvent) => {
                                        setCurrentItem(item);
                                        handleOpen();
                                    }} />
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
                {currentItem() ?
                    <EventInfoModal isOpen={modalOpen()} onClose={handleClose} data={currentItem()!} />
                    : null}
            </div>
        </Match>
    </Switch>
}

const EventItem: Component<{ event: CameraEvent, onClick: (item: CameraEvent) => void }> = (props) => {
    const currentTime = props.event.event.endTime ?
        dayjs(props.event.event.endTime).second() :
        dayjs(Date.now()).second();
    const duration = currentTime - dayjs(props.event.event.startTime).second();
    return <>
        <ListItemButton onClick={() => {
            props.onClick(props.event);
        }}>
            <ListItemAvatar class="mr-2">
                <img src={props.event.presignedUrl} alt="Snapshot" class="w-16 h-16" />
            </ListItemAvatar>
            <ListItemText primary={
                <div class="flex flex-row justify-between items-center mb-2">
                    {"Person detected"}
                    <div class="flex flex-row justify-start items-center gap-2">
                        {props.event.event.label === "person" ? <Chip label="Person" /> : <Chip label="Object" color="secondary" />}
                        {props.event.snapshot.detectedPersonId ? <Chip label="Face" color="primary" /> : null}
                        <Typography variant="body2">{`Duration: ${currentTime - duration}s`}</Typography>
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

interface EventInfoModalProps {
    isOpen: boolean;
    onClose: () => void;
    data: CameraEvent;
}

interface DetectablePerson {
    personId: string;
    name: string;
    age: string;
    presignedUrl: string;
}

async function fetchDetectablePerson(personId: string) {
    if (personId.length > 0) {
        const person = await getPeople([personId]);
        let p: DetectablePerson = {
            personId: person[0].personId,
            name: person[0].name,
            age: person[0].age,
            presignedUrl: ""
        };
        if (person.length > 0) {
            const presignedUrl = await getPeopleImage(personId);
            p.presignedUrl = presignedUrl.presignedUrl;
        }
        return p;
    }
    return null;
}

const EventInfoModal = (props: EventInfoModalProps) => {
    const [detectablePerson] = createResource(props.data.snapshot.detectedPersonId, fetchDetectablePerson);
    return <Modal sx={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center'
    }} open={props.isOpen} onClose={props.onClose}>
        <Paper class="p-8">
            <Box>
                <img src={props.data.presignedUrl} alt="Snapshot" class="block h-96 w-auto" />
            </Box>
            <Box class="mt-4">
                <div class="flex flex-row justify-between items-center">
                    <Typography variant="h6">Event Information</Typography>
                    <div class="flex flex-row gap-2">
                        {props.data.snapshot.detectedPersonId ? <Chip label="Face" color="primary" /> : <Chip label="No face" />}
                    </div>
                </div>
                <Typography variant="body2">{dayjs(props.data.event.frameTime).format("H:mm:ss A on MMM DD, YYYY")}</Typography>
                <div class="flex flex-row gap-4 justify-between items-center mt-4">
                    <div>
                        <Typography variant="body1">Score: {props.data.event.score}</Typography>
                    </div>
                    <div>
                        {(props.data.event.endTime == null || props.data.event.endTime == "") ?
                            <Chip label="Ongoing" color="success" /> :
                            <Chip label="Ended" color="secondary" />}
                    </div>
                </div>
                <Typography variant="body2" class="mt-4">Snapshot ID: {props.data.snapshot.snapshotId}</Typography>
                <Typography variant="body2">Event ID: {props.data.event.eventId}</Typography>
                {detectablePerson() ?
                    <div class="flex flex-row gap-4 justify-between items-center mt-4">
                        <div>
                            <img src={detectablePerson()!.presignedUrl} alt="Person" class="block h-32 w-auto" />
                        </div>
                        <div>
                            <Typography variant="body1">Person Information</Typography>
                            <Typography variant="body2">Name: {detectablePerson()!.name}</Typography>
                            <Typography variant="body2">Age: {detectablePerson()!.age}</Typography>
                            <Typography variant="body2">Person ID: {detectablePerson()!.personId}</Typography>
                        </div>
                    </div> : null}
            </Box>
        </Paper>
    </Modal>
}