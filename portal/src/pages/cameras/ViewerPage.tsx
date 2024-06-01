
import { useParams } from "@solidjs/router";
import { ArrowDownward, ArrowLeft, ArrowRight, ArrowUpward, CircleRounded, Refresh } from "@suid/icons-material";
import { Alert, Box, Button, Chip, CircularProgress, Divider, FormControlLabel, List, ListItemAvatar, ListItemButton, ListItemText, Modal, Paper, Switch as SWButton, Typography } from "@suid/material";
import { red } from "@suid/material/colors";
import green from "@suid/material/colors/green";
import dayjs from "dayjs";
import 'dayjs/locale/en';
import { default as duration } from 'dayjs/plugin/duration';
import { default as relativeTime } from 'dayjs/plugin/relativeTime';
import { Component, For, Match, Switch, createResource, createSignal } from "solid-js";
import { toggleStream } from "../../clients/backend/streams";
import { CameraAggregatedInfo, Event, PTZDirection, doPtzCtrl, getCameraViewInfo, getPersonInfo, getUpdatedInfo } from "../../helper/helper";

dayjs.extend(duration);
dayjs.extend(relativeTime);
dayjs.locale('en');

const Available = () => {
    return <div class="flex flex-row justify-center items-center gap-1">
        <CircleRounded sx={{ color: green[500], fontSize: "12px" }} />
        <p>Active</p>
    </div>
}

const Unavailable = () => {
    return <div class="flex flex-row justify-center items-center gap-1">
        <CircleRounded sx={{ color: red[500], fontSize: "12px" }} />
        <p>Inactive</p>
    </div>
}

export const CameraViewerPage: Component = () => {
    const routeParams = useParams();
    const [data, { refetch }] = createResource(routeParams.cameraId, getCameraViewInfo);
    const [events, { refetch: eventRefetch }] = createResource(data, async (data: CameraAggregatedInfo) => {
        return await getUpdatedInfo({
            cameraId: data.camera.cameraId,
            cameraName: data.camera.openGateCameraName,
            transcoderId: data.camera.transcoderId,
        },
            10,
            true,
            600); // 10 minutes
    });
    const [isPtzCtrlInProcess, setIsPtzCtrlInProcess] = createSignal(false);

    const [modalOpen, setModalOpen] = createSignal(false);
    const handleOpen = () => setModalOpen(true);
    const handleClose = () => setModalOpen(false);
    const [currentItem, setCurrentItem] = createSignal<Event | null>(null);
    const [ptzError, setPtzError] = createSignal<any | null>(null);

    const timer = setInterval(() => eventRefetch(routeParams.cameraId), 10000);

    const handlePtzCtrl = (direction: PTZDirection) => {
        return async () => {
            setIsPtzCtrlInProcess(true);
            try {
                await doPtzCtrl(
                    direction,
                    data()!.
                        camera.
                        cameraId);
            } catch (e: any) {
                setPtzError(e);
            }
            setIsPtzCtrlInProcess(false);
        }
    }

    return <Switch>
        <Match when={data.loading}>
            <div class="flex flex-row justify-center mt-8">
                <CircularProgress />
            </div>
        </Match>
        <Match when={data.error}>Error: {data.error.message}</Match>
        <Match when={data()}>
            {ptzError() != null && <div class="absolute top-8 right-8 z-50 w-80">
                <Alert severity="error" onClose={() => { setPtzError(null) }}>
                    Unable to complete the request. Please try again later.
                </Alert>
            </div>}
            <div class="flex flex-row h-full w-full">
                <div class="w-full h-full" style="flex: 7">
                    <Switch>
                        <Match when={data()!.camera.enabled}>
                            <iframe src={data()!.streamInfo.streamUrl} width={"100%"} height={"100%"} allowfullscreen />
                        </Match>
                        <Match when={!data()?.camera.enabled}>
                            <div class="flex flex-row justify-center items-center h-full w-full bg-gray-900">
                                <Typography variant="h6" color="white">Camera is disabled</Typography>
                            </div>
                        </Match>
                    </Switch>
                </div>
                <div class="w-full h-full p-4 flex flex-col gap-4" style="flex: 7">
                    <Paper sx={{ width: '100%', height: "fit-content", padding: '1rem' }}>
                        <div class="flex w-full h-fit flex-row justify-between items-center gap-8">
                            <div>
                                <Typography variant="body1" fontWeight={500}>{data()!.camera.name}</Typography>
                                <FormControlLabel
                                    label="Enabled"
                                    checked={data()!.camera.enabled}
                                    onChange={async () => {
                                        await toggleStream(
                                            data()!.
                                                camera.
                                                cameraId,
                                            !data()!.
                                                camera.
                                                enabled);
                                        refetch();
                                        eventRefetch();
                                    }}
                                    control={<SWButton size="small" inputProps={{ "aria-label": "controlled" }} defaultChecked />}
                                />
                            </div>
                            <div class="flex justify-between items-center">
                                <div class="grid grid-cols-4 gap-x-4 gap-y-2 items-center">
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm">Edge TPU</p>
                                        <p>{data()?.integration.withEdgeTpu ? "Yes" : "No"}</p>
                                    </div>
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm"> Hardware Accel.</p>
                                        <p>{data()?.integration.hardwareAccelerationType ?? "N/A"}</p>
                                    </div >
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm">Log Level</p>
                                        <p>{data()?.integration.logLevel}</p>
                                    </div>
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm">Retention (days)</p>
                                        <p>{data()?.integration.snapshotRetentionDays}</p>
                                    </div>
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm">OpenGate</p>
                                        <p>{data()?.transcoderStatus.openGateStatus ? <Available /> : <Unavailable />}</p>
                                    </div>
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm">Transcoder</p>
                                        <p>{data()?.transcoderStatus.transcoderStatus ? <Available /> : <Unavailable />}</p>
                                    </div>
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm">Autotracking</p>
                                        <p>{data()?.transcoderStatus.autotracker ? <Available /> : <Unavailable />}</p>
                                    </div>
                                    <div class="flex flex-col justify-start items-start">
                                        <p class="font-semibold text-sm">Object Det.</p>
                                        <p>{data()?.transcoderStatus.objectDetection ? <Available /> : <Unavailable />}</p>
                                    </div>
                                </div>
                            </div>
                            <div class="flex flex-row justify-end align-middle gap-2">
                                <Button size="small" onClick={() => eventRefetch()} startIcon={<Refresh />} variant="contained">
                                    {"Refresh"}
                                </Button>
                            </div>
                        </div>
                    </Paper>
                    <Paper sx={{ width: '100%' }} class="overflow-y-scroll h-96">
                        <List>
                            <Switch>
                                <Match when={events.loading && events() == undefined}>
                                    <div class="flex flex-row justify-center mt-8">
                                        <CircularProgress />
                                    </div>
                                </Match>
                                <Match when={events.error}>Error: {events.error.message}</Match>
                                <Match when={events()}>
                                    <For each={events()!.events}>
                                        {event => <div>
                                            <EventItem event={event} onClick={(item: Event) => {
                                                setCurrentItem(item);
                                                handleOpen();
                                            }} />
                                            <Divider />
                                        </div>}
                                    </For>
                                </Match>
                            </Switch>
                        </List>
                    </Paper>
                    <Paper sx={{ width: '100%', height: "fit-content", padding: '1rem' }} class="flex flex-row gap-8 items-center">
                        <div class="flex flex-row gap-4 h-fit">
                            <Button variant="contained" sx={{ width: '0.5rem' }} color="primary" disabled={isPtzCtrlInProcess()} onClick={handlePtzCtrl(PTZDirection.Left)}><ArrowLeft /></Button>
                            <div class="flex flex-col gap-4 justify-between items-center">
                                <Button variant="contained" fullWidth color="primary" disabled={isPtzCtrlInProcess()} onClick={handlePtzCtrl(PTZDirection.Up)}><ArrowUpward /></Button>
                                <Button variant="contained" fullWidth color="primary" disabled={isPtzCtrlInProcess()} onClick={handlePtzCtrl(PTZDirection.Down)}><ArrowDownward /></Button>
                            </div>
                            <Button variant="contained" color="primary" disabled={isPtzCtrlInProcess()} onClick={handlePtzCtrl(PTZDirection.Right)}><ArrowRight /></Button>
                        </div>
                        <div class="flex flex-row justify-center items-center">
                            <div class="grid grid-cols-4 gap-y-2 gap-x-4 items-center">
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Camera Name</p>
                                    <p>{events()?.stats?.cameraName ?? data()!.camera.openGateCameraName}</p>
                                </div>
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Camera (fps)</p>
                                    <p>{events()?.stats?.cameraFps ?? "N/A"}</p>
                                </div>
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Detection (fps)</p>
                                    <p>{events()?.stats?.detectionFps ?? "N/A"}</p>
                                </div>
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Capture PID</p>
                                    <p>{events()?.stats?.capturePid ?? "N/A"}</p>
                                </div>
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Last Updated</p>
                                    <p>{events()?.stats?.timestamp ?
                                        getElapsed(events()!.stats!.timestamp, "").humanize(false) :
                                        "N/A"}</p>
                                </div>
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Processed (fps)</p>
                                    <p>{events()?.stats?.processFps ?? "N/A"}</p>
                                </div>
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Skipped (fps)</p>
                                    <p>{events()?.stats?.skippedFps ?? "N/A"}</p>
                                </div>
                                <div class="flex flex-col justify-start items-start">
                                    <p class="font-semibold text-sm">Inference Time (ms)</p>
                                    <p>{events()?.detectorStats?.inferenceSpeed ?? "N/A"}</p>
                                </div>
                            </div>
                        </div>
                    </Paper>
                </div>
                {currentItem() ?
                    <EventInfoModal isOpen={modalOpen()} onClose={() => {
                        handleClose();
                        setCurrentItem(null);
                    }} data={currentItem()!} />
                    : null}
            </div>
        </Match>
    </Switch>
}

interface EventItemProps {
    event: Event;
    onClick: (item: Event) => void;
}

function getEndTime(endTime: string): dayjs.Dayjs {
    if (endTime != null) {
        return dayjs(endTime);
    }
    return dayjs(Date.now());
}

function getElapsed(startTime: string, endTime: string): duration.Duration {
    const currentTime = getEndTime(endTime);
    const start = dayjs(startTime);
    const dur = currentTime.diff(start, 'second');
    return dayjs.duration(dur);
}

const EventItem: Component<EventItemProps> = (props: EventItemProps) => {
    const { tracking, snapshot } = props.event;
    const elasped = getElapsed(tracking.startTime, tracking.endTime);

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
                        {tracking.label === "person" ? <Chip label="Person" /> : <Chip label="Object" color="secondary" />}
                        {snapshot?.detectedPeopleId ? <Chip label="Face" color="primary" /> : null}
                        <Typography variant="body2">{`Duration: ${elasped.humanize(false)}`}</Typography>
                    </div>
                </div>
            } secondary={
                <div class="flex flex-row justify-between items-center">
                    <Typography variant="body2">
                        {`Score: ${Math.round(tracking.score * 100) / 100}`}
                    </Typography>
                    {`${dayjs(tracking.frameTime).format("H:mm:ss A on MMM DD, YYYY")}`}
                </div>
            } />
        </ListItemButton>
    </>
}

interface EventInfoModalProps {
    isOpen: boolean;
    onClose: () => void;
    data: Event;
}

const EventInfoModal = (props: EventInfoModalProps) => {
    const { tracking, snapshot, presignedUrl } = props.data;
    const [personInfo, setPersonInfo] = createResource(
        snapshot.detectedPeopleId,
        getPersonInfo);
    return <Modal sx={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center'
    }} open={props.isOpen} onClose={props.onClose}>
        <Paper class="p-8">
            <Box>
                <img src={presignedUrl} alt="Snapshot" class="block h-96 w-auto" />
            </Box>
            <Box class="mt-4">
                <div class="flex flex-row justify-between items-center">
                    <Typography variant="h6">Event Information</Typography>
                    <div class="flex flex-row gap-2">
                        {snapshot.detectedPeopleId ? <Chip label="Face" color="primary" /> : <Chip label="No face" />}
                    </div>
                </div>
                <Typography variant="body2">{dayjs(tracking.frameTime).format("H:mm:ss A on MMM DD, YYYY")}</Typography>
                <div class="flex flex-row gap-4 justify-between items-center mt-4">
                    <div>
                        <Typography variant="body1">Score: {tracking.score}</Typography>
                    </div>
                    <div>
                        {(tracking.endTime == null || tracking.endTime == "") ?
                            <Chip label="Ongoing" color="success" /> :
                            <Chip label="Ended" color="secondary" />}
                    </div>
                </div>
                <Typography variant="body2" class="mt-4">Snapshot ID: {snapshot.snapshotId}</Typography>
                <Typography variant="body2">Event ID: {tracking.eventId}</Typography>
                {personInfo() ?
                    <div class="flex flex-row gap-4 justify-between items-center mt-4">
                        <div>
                            <img src={personInfo()!.image.presignedUrl} alt="Person" class="block h-32 w-auto" />
                        </div>
                        <div>
                            <Typography variant="body1">Person Information</Typography>
                            <Typography variant="body2">Name: {personInfo()!.person.name}</Typography>
                            <Typography variant="body2">Age: {personInfo()!.person.age}</Typography>
                            <Typography variant="body2">Person ID: {personInfo()!.person.personId}</Typography>
                        </div>
                    </div> : null}
            </Box>
        </Paper>
    </Modal>
}