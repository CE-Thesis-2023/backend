
import { useParams } from "@solidjs/router";
import { ArrowDownward, ArrowLeft, ArrowRight, ArrowUpward, Refresh, Visibility } from "@suid/icons-material";
import { Box, Button, Chip, CircularProgress, Divider, FormControl, FormControlLabel, IconButton, Input, InputAdornment, InputLabel, List, ListItemAvatar, ListItemButton, ListItemText, Modal, Paper, Switch as SWButton, Typography } from "@suid/material";
import dayjs from "dayjs";
import { Component, For, Match, Switch, createResource, createSignal } from "solid-js";
import { CameraAggregatedInfo, Event, getCameraViewInfo, getPersonInfo, getUpdatedInfo } from "../../helper/helper";

export const CameraViewerPage: Component = () => {
    const routeParams = useParams();
    const [data, { refetch }] = createResource(routeParams.cameraId, getCameraViewInfo);
    const [events, { refetch: eventRefetch }] = createResource(data, async (data: CameraAggregatedInfo) => {
        return await getUpdatedInfo({
            cameraId: data.camera.cameraId,
            cameraName: data.camera.openGateCameraName,
            transcoderId: data.camera.transcoderId,
        });
    });

    const [modalOpen, setModalOpen] = createSignal(false);
    const handleOpen = () => setModalOpen(true);
    const handleClose = () => setModalOpen(false);
    const [currentItem, setCurrentItem] = createSignal<Event | null>(null);

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
                            <For each={events()?.events}>
                                {event => <>
                                    <EventItem event={event} onClick={(item: Event) => {
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

interface EventItemProps {
    event: Event;
    onClick: (item: Event) => void;
}

function getEndTime(endTime: string): number {
    return endTime ?
        dayjs(endTime).second() :
        dayjs(Date.now()).second();
}

function getElapsed(startTime: string, endTime: string): number {
    const currentTime = getEndTime(endTime);
    return currentTime -
        dayjs(startTime).second();
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
                        {snapshot.detectedPeopleId ? <Chip label="Face" color="primary" /> : null}
                        <Typography variant="body2">{`Duration: ${elasped}s`}</Typography>
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