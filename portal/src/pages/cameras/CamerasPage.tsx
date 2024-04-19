import createDebounce from "@solid-primitives/debounce";
import { Add, CheckCircleRounded, CircleRounded, Delete, MoreVert, Refresh, Settings } from "@suid/icons-material";
import { Box, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, IconButton, Link, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from "@suid/material";
import green from "@suid/material/colors/green";
import red from "@suid/material/colors/red";
import { Component, Match, Show, Switch, createResource, createSignal } from "solid-js";
import { Transcoder, deleteCamera, getCameras, getOpenGateCameraSettings, getOpenGateConfigurations, getTranscoders } from "../../clients/backend/client";
import Codeblock from "../../components/Codeblock";

async function fetchData(cameraIds: string[]) {
    const cameras = await getCameras(cameraIds);
    const transcoderIds = cameras.map(camera => camera.transcoderId);
    const transcoders = await getTranscoders(transcoderIds);
    const transcoderMap = new Map<string, Transcoder>();
    for (let i = 0; i < transcoders.length; i++) {
        transcoderMap.set(transcoders[i].deviceId, transcoders[i]);
    }
    const aggregated = cameras.map(camera => {
        const ltd = transcoderMap.get(camera.transcoderId);
        return {
            camera: camera,
            transcoder: ltd,
        };
    });

    return aggregated;
}

export const CamerasPage: Component = () => {
    const [cameraIds, setCamerasIds] = createSignal<string[]>([]);

    const [menuAnchorEl, setMenuAnchorEl] = createSignal<null | HTMLElement>(null);
    const [menuCurrentlySelectedId, setMenuCurrentlySelectedId] = createSignal<string>("");
    const [menuDialogCurrentItem, setMenuDialogCurrentItem] = createSignal<any | null>(null);

    const [settingsDialogStatus, setSettingsDialogStatus] = createSignal<boolean>(false);
    const [deleteDialogStatus, setDeleteDialogStatus] = createSignal<boolean>(false);

    const open = () => Boolean(menuAnchorEl());
    const handleClose = () => setMenuAnchorEl(null);

    const filterDebouncer = createDebounce((m: string) => {
        if (m === "") {
            setCamerasIds([]);
        } else {
            setCamerasIds(m.split(","));
        }
    }, 350);
    const [cameras, { refetch }] = createResource(cameraIds, fetchData);

    const moreMenuItems = [
        {
            id: 'delete',
            name: 'Delete',
            icon: <Delete />,
            onClick: () => {
                setDeleteDialogStatus(true);
            }
        },
        {
            id: 'settings',
            name: 'Settings',
            icon: <Settings />,
            onClick: () => {
                setSettingsDialogStatus(true);
            }
        }
    ]

    return <div class="w-full h-full p-8">
        <Paper class="h-max w-full">
            <div class="flex flex-row justify-between items-center p-4">
                <Typography variant="h6" component="h1">Cameras</Typography>
                <div class="flex flex-row items-end gap-2">
                    <TextField id="camera-id-filter" label="Camera ID" variant="standard" size="small" color="primary" margin="none" sx={{ marginRight: '1rem' }} onChange={(e) => {
                        const value = e.target.value;
                        filterDebouncer(value);
                    }} />
                    <Button variant="contained" size="medium" color="primary" startIcon={<Add />}>
                        Add Camera
                    </Button>
                    <Button variant="contained" size="medium" color="primary" startIcon={<Refresh />} onClick={() => {
                        refetch();
                    }}>
                        Refresh
                    </Button>
                </div>
            </div>
            <Show when={cameras.loading}>
                <div class="flex flex-row justify-center align-middle w-full">
                    <CircularProgress color="primary" sx={{ margin: '2rem' }} />
                </div>
            </Show>
            <Switch>
                <Match when={cameras.error}>
                    <p>Error when retrieving data: {cameras.error}</p>
                </Match>
                <Match when={cameras() && !cameras.loading}>
                    <Box sx={{
                        display: "flex",
                        flexWrap: "wrap",
                    }}>
                        <TableContainer>
                            <Table aria-label="camera-list">
                                <TableHead>
                                    <TableRow>
                                        <TableCell align="center">Status</TableCell>
                                        <TableCell sx={{ width: '20%' }}>Name</TableCell>
                                        <TableCell align="left">ID</TableCell>
                                        <TableCell align="left">IP Address</TableCell>
                                        <TableCell align="left">Transcoder</TableCell>
                                        <TableCell align="center"></TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {cameras()?.map((camera, _) => {
                                        return <>
                                            <TableRow sx={{ "&:last-child td, &:last-child th": { border: 0 } }}>
                                                <TableCell align="center">{camera.camera.enabled === true ?
                                                    <CheckCircleRounded fontSize="small" sx={{ color: green[500] }} /> :
                                                    <CircleRounded fontSize="small" sx={{ color: red[500] }} />}</TableCell>
                                                <TableCell align="left">
                                                    <Link variant="body2" underline="none" href={`/cameras/${camera.camera.cameraId}`}>
                                                        {camera.camera.name}
                                                    </Link>
                                                </TableCell>
                                                <TableCell align="left">
                                                    {camera.camera.cameraId}
                                                </TableCell>
                                                <TableCell align="left">{camera.camera.ip}</TableCell>
                                                <TableCell align="left">{camera.transcoder?.name}</TableCell>
                                                <TableCell align="center">
                                                    <IconButton
                                                        aria-label="more"
                                                        id="long-button"
                                                        aria-controls={open() ? 'long-menu' : undefined}
                                                        aria-expanded={open() ? 'true' : undefined}
                                                        aria-haspopup="true"
                                                        onClick={(e) => {
                                                            setMenuDialogCurrentItem(camera);
                                                            setMenuAnchorEl(e.currentTarget);
                                                        }}
                                                    >
                                                        <MoreVert />
                                                    </IconButton>
                                                    <Menu
                                                        id="long-menu"
                                                        anchorEl={menuAnchorEl()}
                                                        MenuListProps={{ "aria-labelledby": "long-button" }}
                                                        open={open()}
                                                        onClose={() => {
                                                            setMenuCurrentlySelectedId("");
                                                            setMenuDialogCurrentItem(null);
                                                            handleClose();
                                                        }}
                                                        PaperProps={{
                                                            elevation: 2,
                                                        }}
                                                    >
                                                        {moreMenuItems.map((item, _) => {
                                                            return <MenuItem selected={menuCurrentlySelectedId() === item.id}
                                                                onClick={() => {
                                                                    setMenuCurrentlySelectedId(item.id);
                                                                    item.onClick();
                                                                }}>
                                                                <ListItemIcon>
                                                                    {item.icon}
                                                                </ListItemIcon>
                                                                <ListItemText>
                                                                    {item.name}
                                                                </ListItemText>
                                                            </MenuItem>
                                                        })}
                                                    </Menu>
                                                </TableCell>
                                            </TableRow>
                                        </>;
                                    })}
                                </TableBody>
                            </Table>
                        </TableContainer>
                    </Box>
                </Match>
            </Switch>
            {
                settingsDialogStatus() &&
                <CameraSettingsDialog
                    onClose={() => {
                        setSettingsDialogStatus(false);
                        handleClose();
                    }}
                    open={settingsDialogStatus()}
                    cameraId={menuDialogCurrentItem().
                        camera.
                        cameraId}
                    openGateConfigurationsId={menuDialogCurrentItem().
                        transcoder.
                        openGateIntegrationId}
                />
            }
            {
                deleteDialogStatus() &&
                <DeleteCameraConfirmDialog
                    cameraId={menuDialogCurrentItem().camera.cameraId}
                    open={deleteDialogStatus()}
                    onClose={() => {
                        setDeleteDialogStatus(false);
                        handleClose();
                        refetch();
                    }}
                />
            }
        </Paper>
    </div>
}

interface CameraSettingsDialogProps {
    cameraId: string;
    openGateConfigurationsId: string;
    onClose: () => void;
    open: boolean;
}

async function fetchCameraSettings(ids: { cameraId: string, openGateConfigurationsId: string }) {
    const openGateConfigs = await getOpenGateConfigurations(ids.openGateConfigurationsId);
    const openGateSettings = await getOpenGateCameraSettings([ids.cameraId]);
    return {
        openGateConfigs: openGateConfigs,
        openGateSettings: openGateSettings
    };
}

const CameraSettingsDialog = (props: CameraSettingsDialogProps) => {
    const [cameraSettings] = createResource({
        cameraId: props.cameraId,
        openGateConfigurationsId: props.openGateConfigurationsId
    }, fetchCameraSettings);

    return <Dialog onClose={props.onClose} open={props.open}>
        <DialogTitle>Camera Settings</DialogTitle>
        <DialogContent>
            <DialogContentText>For debugging purposes</DialogContentText>
            <Switch>
                <Match when={cameraSettings.loading}>
                    <div class="flex flex-row justify-center align-middle w-full">
                        <CircularProgress color="primary" sx={{ margin: '2rem' }} />
                    </div>
                </Match>
                <Match when={cameraSettings.error}>
                    <p>Error when retrieving data: {cameraSettings.error}</p>
                </Match>
                <Match when={cameraSettings() && !cameraSettings.loading}>
                    <Codeblock class={"mt-2"} title="OpenGate Configurations">
                        {JSON.stringify(cameraSettings()?.openGateConfigs, null, 2)}
                    </Codeblock>
                    <Codeblock class={"mt-2"} title="OpenGate Camera Settings">
                        {JSON.stringify(cameraSettings()?.openGateSettings, null, 2)}
                    </Codeblock>
                </Match>
            </Switch>
        </DialogContent>
    </Dialog>
}

interface DeleteCameraConfirmDialogProps {
    cameraId: string;
    open: boolean;
    onClose: () => void;
}

const DeleteCameraConfirmDialog = (props: DeleteCameraConfirmDialogProps) => {
    const [waitingDelete, setWaitingDelete] = createSignal<boolean>(false);
    const handleDelete = async () => {
        setWaitingDelete(true);
        await deleteCamera(props.cameraId);
        setWaitingDelete(false);
        props.onClose();
    }

    return <Dialog onClose={props.onClose} open={props.open || waitingDelete()}>
        <DialogTitle id="alert-dialog-title">
            Delete camera?
        </DialogTitle>
        <DialogContent>
            <DialogContentText id="alert-dialog-description">
                {`Requested to delete camera with ID: ${props.cameraId}, proceed?`}
            </DialogContentText>
        </DialogContent>
        <DialogActions>
            <Button onClick={props.onClose}>Disagree</Button>
            <Button onClick={handleDelete} disabled={waitingDelete()}>Agree</Button>
        </DialogActions>
    </Dialog>
}