import createDebounce from "@solid-primitives/debounce";
import { Add, CheckCircleRounded, CircleRounded, Delete, MoreVert, Refresh, Settings } from "@suid/icons-material";
import { Alert, Box, Button, Chip, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormControl, FormControlLabel, IconButton, InputLabel, Link, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Select, Switch as SwitchButton, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from "@suid/material";
import green from "@suid/material/colors/green";
import red from "@suid/material/colors/red";
import { Component, Match, Show, Switch, createResource, createSignal } from "solid-js";
import { AddCameraParams, addCamera, deleteCamera } from "../../clients/backend/cameras";
import { Transcoder, getTranscoders } from "../../clients/backend/transcoders";
import Codeblock from "../../components/Codeblock";
import { CameraItem, getListCameras } from "../../helper/helper";

export const CamerasPage: Component = () => {
    const [cameraIds, setCamerasIds] = createSignal<string[]>([]);

    const [menuAnchorEl, setMenuAnchorEl] = createSignal<null | HTMLElement>(null);
    const [menuCurrentlySelectedId, setMenuCurrentlySelectedId] = createSignal<string>("");
    const [menuDialogCurrentItem, setMenuDialogCurrentItem] = createSignal<any | null>(null);

    const [settingsDialogStatus, setSettingsDialogStatus] = createSignal<boolean>(false);
    const [deleteDialogStatus, setDeleteDialogStatus] = createSignal<boolean>(false);
    const [addCameraDialogStatus, setAddCameraDialogStatus] = createSignal<boolean>(false);

    const open = () => Boolean(menuAnchorEl());
    const handleClose = () => setMenuAnchorEl(null);

    const filterDebouncer = createDebounce((m: string) => {
        if (m === "") {
            setCamerasIds([]);
        } else {
            setCamerasIds(m.split(","));
        }
    }, 350);
    const [cameras, { refetch }] = createResource(cameraIds, getListCameras);

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
                    <Button variant="contained" size="medium" color="primary" startIcon={<Add />} onClick={() => {
                        setAddCameraDialogStatus(true);
                    }}>
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
                                    {cameras()?.items.map((camera, _) => {
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
                    item={menuDialogCurrentItem()}
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
            {
                addCameraDialogStatus() &&
                <AddCameraDialog
                    open={addCameraDialogStatus()}
                    transcoders={cameras()!.items.map((item) => item.transcoder!)}
                    onClose={() => {
                        setAddCameraDialogStatus(false);
                        handleClose();
                        refetch();
                    }}
                />
            }
        </Paper>
    </div>
}

interface CameraSettingsDialogProps {
    item: CameraItem;
    onClose: () => void;
    open: boolean;
}

const CameraSettingsDialog = (props: CameraSettingsDialogProps) => {
    return <Dialog onClose={props.onClose} open={props.open}>
        <DialogTitle>Camera Settings</DialogTitle>
        <DialogContent>
            <DialogContentText>For debugging purposes</DialogContentText>
            <Codeblock class={"mt-2"} title="OpenGate Configurations">
                {JSON.stringify(props.item.configs, null, 2)}
            </Codeblock>
            <Codeblock class={"mt-2"} title="OpenGate Camera Settings">
                {JSON.stringify(props.item.settings, null, 2)}
            </Codeblock>
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

interface AddCameraDialogProps {
    open: boolean;
    transcoders: Transcoder[];
    onClose: () => void;
}

const AddCameraDialog = (props: AddCameraDialogProps) => {
    const [waitingAdd, setWaitingAdd] = createSignal<boolean>(false);
    const handleAdd = async () => {
        setWaitingAdd(true);
        try {
            await addCamera(formValue());
        } catch (e: any) {
            setFormErr(e.message);
            setWaitingAdd(false);
            return;
        }
        setWaitingAdd(false);
        props.onClose();
    }
    const [formErr, setFormErr] = createSignal<string | null>(null);
    const [formValue, setFormValue] = createSignal<AddCameraParams>(
        {
            name: "",
            ip: "",
            port: 80,
            username: "",
            password: "",
            transcoderId: "",
            autotracking: false,
        }
    );
    const [trancoders, { refetch: fetchTranscoders }] = createResource([], getTranscoders);
    return <Switch>
        <Match when={trancoders.loading}>
            <CircularProgress color="primary" sx={{ margin: '2rem' }} />
        </Match>
        <Match when={trancoders.error}>
            <Alert severity="error">{trancoders.error}</Alert>
        </Match>
        <Match when={trancoders() && !trancoders.loading}>
            <Dialog open={props.open} onClose={props.onClose}>
                <DialogTitle>Add Camera</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Add a new camera
                    </DialogContentText>
                    <DialogContent>
                        <TextField
                            autoFocus
                            margin="dense"
                            id="name"
                            label="Name"
                            type="text"
                            variant="filled"
                            fullWidth
                            required
                            value={formValue().name}
                            onChange={(e) => {
                                const value = e.target.value;
                                setFormValue((prev) => {
                                    return {
                                        ...prev,
                                        name: value
                                    }
                                });
                            }}
                        />
                        <TextField
                            margin="dense"
                            id="ip"
                            label="IP Address"
                            type="url"
                            variant="filled"
                            fullWidth
                            required
                            value={formValue().ip}
                            onChange={(e) => {
                                const value = e.target.value;
                                setFormValue((prev) => {
                                    return {
                                        ...prev,
                                        ip: value
                                    }
                                });
                            }}
                        />
                        <TextField
                            margin="dense"
                            id="port"
                            label="Port"
                            type="number"
                            variant="filled"
                            value={formValue().port}
                            onChange={(e) => {
                                const value = e.target.value;
                                setFormValue((prev) => {
                                    return {
                                        ...prev,
                                        port: parseInt(value)
                                    }
                                });
                            }}
                            fullWidth
                        />
                        <FormControl
                            variant="filled"
                            margin="dense"
                            sx={{
                                marginTop: '8px',
                                width: '100%',
                            }}>
                            <InputLabel id="select-filled-label">
                                Transcoder
                            </InputLabel>
                            <Select
                                id="demo-simple-select-filled"
                                labelId="select-filled-label"
                                fullWidth
                                variant="filled"
                                value={formValue().transcoderId}
                                onChange={(e) => {
                                    const value = e.target.value;
                                    setFormValue((prev) => {
                                        return {
                                            ...prev,
                                            transcoderId: value
                                        }
                                    });
                                }}
                            >
                                {trancoders()!.map((transcoder, _) => {
                                    return <MenuItem value={transcoder.deviceId}>
                                        <div class="flex flex-row justify-start items-center gap-2">
                                            {transcoder.name}
                                            <Chip label={transcoder.deviceId} />
                                        </div>
                                    </MenuItem>
                                })}
                            </Select>
                        </FormControl>
                        <TextField
                            margin="dense"
                            id="username"
                            label="Username"
                            type="text"
                            variant="filled"
                            fullWidth
                            required
                            value={formValue().username}
                            onChange={(e) => {
                                const value = e.target.value;
                                setFormValue((prev) => {
                                    return {
                                        ...prev,
                                        username: value
                                    }
                                });
                            }}
                        />
                        <TextField
                            margin="dense"
                            id="password"
                            label="Password"
                            type="password"
                            variant="filled"
                            fullWidth
                            required
                            value={formValue().password}
                            onChange={(e) => {
                                const value = e.target.value;
                                setFormValue((prev) => {
                                    return {
                                        ...prev,
                                        password: value
                                    }
                                });
                            }}
                        />
                        <FormControlLabel
                            label="PTZ Auto-tracking"
                            control={
                                <SwitchButton
                                    value={formValue()?.autotracking ?? false}
                                    onChange={(e) => {
                                        const value = e.target.checked;
                                        setFormValue((prev) => {
                                            return {
                                                ...prev,
                                                autotracking: value
                                            }
                                        });
                                    }}
                                />}
                        />
                    </DialogContent>
                    {
                        formErr() != null &&
                        <DialogContentText>
                            <Alert severity="error">{formErr()}</Alert>
                        </DialogContentText>
                    }
                    <DialogActions>
                        <Button onClick={props.onClose}>Cancel</Button>
                        <Button onClick={handleAdd} disabled={waitingAdd()}>Submit</Button>
                    </DialogActions>
                </DialogContent>
            </Dialog >
        </Match>
    </Switch>
}