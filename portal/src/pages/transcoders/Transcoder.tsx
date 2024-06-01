import createDebounce from "@solid-primitives/debounce";
import { MoreVert, Refresh, Update } from "@suid/icons-material";
import { Alert, Box, Button, Checkbox, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormControl, FormControlLabel, IconButton, InputLabel, Link, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Select, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from "@suid/material";
import { Component, Match, Show, Switch, createResource, createSignal } from "solid-js";
import { UpdateTranscoder, updateTranscoder } from "../../clients/backend/transcoders";
import { TranscoderInfo, getListTranscoders } from "../../helper/helper";

export const TranscoderPage: Component = () => {
    const [transcoderIds, setTranscoderIds] = createSignal<string[]>([]);
    const [menuAnchorEl, setMenuAnchorEl] = createSignal<HTMLElement | null>(null);
    const [menuDialogCurrentItem, setMenuDialogCurrentItem] = createSignal<TranscoderInfo | null>(null);
    const [menuCurrentlySelectedId, setMenuCurrentlySelectedId] = createSignal<string>("");
    const [updateDeviceDialogVisible, setUpdateDeviceDialogVisible] = createSignal<boolean>(false);

    const openMenu = () => Boolean(menuAnchorEl());
    const closeMenu = () => setMenuAnchorEl(null);

    const filterDebouncer = createDebounce((m: string) => {
        if (m === "") {
            setTranscoderIds([]);
        } else {
            setTranscoderIds(m.split(","));
        }
    }, 350);

    const moreMenuItems = [
        {
            id: 'update',
            name: 'Update',
            icon: <Update />,
            onClick: () => {
                setUpdateDeviceDialogVisible(true);
            }
        }
    ]

    const [transcoders, { refetch: fetchTranscoders }] = createResource(transcoderIds, getListTranscoders);
    return <div class="w-full h-full p-8">
        <Paper class="h-max w-full">
            <div class="flex flex-row justify-between items-center p-4">
                <Typography variant="h6" component="h1">People</Typography>
                <div class="flex flex-row items-end gap-2">
                    <TextField id="person-id-filter" label="Device ID" variant="standard" size="small" color="primary" margin="none" sx={{ marginRight: '1rem' }} onChange={(e) => {
                        const value = e.target.value;
                        filterDebouncer(value);
                    }} />
                    <Button variant="contained" size="medium" color="primary" startIcon={<Refresh />} onClick={() => {
                        fetchTranscoders();
                    }}>
                        Refresh
                    </Button>
                </div>
            </div>
            <Show when={transcoders.loading}>
                <div class="flex flex-row justify-center align-middle w-full">
                    <CircularProgress color="primary" sx={{ margin: '2rem' }} />
                </div>
            </Show>
            <Switch>
                <Match when={transcoders.error}>
                    <p>Error when retrieving data: {transcoders.error}</p>
                </Match>
                <Match when={transcoders() && !transcoders.loading}>
                    <Box sx={{
                        display: "flex",
                        flexWrap: "wrap",
                    }}>
                        <TableContainer>
                            <Table aria-label="person-list">
                                <TableHead>
                                    <TableRow>
                                        <TableCell sx={{ width: '20%' }}>ID</TableCell>
                                        <TableCell align="left">Name</TableCell>
                                        <TableCell align="center">Edge TPU</TableCell>
                                        <TableCell align="center">Hardware Accel.</TableCell>
                                        <TableCell align="center">Log Level</TableCell>
                                        <TableCell align="center">Actions</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {transcoders()?.items.map((t, _) => {
                                        return <>
                                            <TableRow sx={{ "&:last-child td, &:last-child th": { border: 0 } }}>
                                                <TableCell align="left">
                                                    <Link variant="body2" underline="none">
                                                        {t.transcoder.deviceId}
                                                    </Link>
                                                </TableCell>
                                                <TableCell align="left">
                                                    {t.transcoder.name}
                                                </TableCell>
                                                <TableCell align="left">{t.integration.withEdgeTpu ? "Yes" : "No"}</TableCell>
                                                <TableCell align="center">{t.integration.hardwareAccelerationType !== "cpu" ? "Yes" : "No"}</TableCell>
                                                <TableCell align="center">{t.integration.logLevel}</TableCell>
                                                <TableCell align="center">
                                                    <IconButton
                                                        aria-label="update"
                                                        id="long-button"
                                                        aria-controls={openMenu() ? 'long-menu' : undefined}
                                                        aria-expanded={openMenu() ? 'true' : undefined}
                                                        aria-haspopup="true"
                                                        onClick={(e) => {
                                                            setMenuDialogCurrentItem(t);
                                                            setMenuAnchorEl(e.currentTarget);
                                                        }}
                                                    >
                                                        <MoreVert />
                                                    </IconButton>
                                                    <Menu
                                                        id="long-menu"
                                                        anchorEl={menuAnchorEl()}
                                                        MenuListProps={{ "aria-labelledby": "long-button" }}
                                                        open={openMenu()}
                                                        onClose={() => {
                                                            setMenuCurrentlySelectedId("");
                                                            setMenuDialogCurrentItem(null);
                                                            closeMenu();
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
                updateDeviceDialogVisible() &&
                <UpdateDeviceDialog
                    open={updateDeviceDialogVisible()}
                    deviceId={menuDialogCurrentItem()!.transcoder.deviceId}
                    existingDevice={menuDialogCurrentItem()!}
                    onClose={() => {
                        setUpdateDeviceDialogVisible(false);
                        closeMenu();
                        fetchTranscoders();
                    }}
                />
            }

        </Paper>
    </div>
}

interface UpdateDeviceDialogProps {
    deviceId: string;
    existingDevice: TranscoderInfo;
    open: boolean;
    onClose: () => void;
}

const UpdateDeviceDialog = (props: UpdateDeviceDialogProps) => {
    const [waitingUpdate, setWaitingUpdate] = createSignal<boolean>(false);
    const [formErr, setFormErr] = createSignal<string | null>(null);

    const [data, setData] = createSignal<UpdateTranscoder>({
        id: props.deviceId,
        name: props.existingDevice.transcoder.name,
        logLevel: props.existingDevice.integration.logLevel,
        hardwareAccelerationType: props.existingDevice.integration.hardwareAccelerationType,
        edgeTpuEnabled: props.existingDevice.integration.withEdgeTpu,
    });

    const handleUpdate = async () => {
        setWaitingUpdate(true);
        try {
            await updateTranscoder(data());
        } catch (e: any) {
            setFormErr(e.message);
            setWaitingUpdate(false);
            return;
        }
        setWaitingUpdate(false);
        props.onClose();
    }

    return <Dialog open={props.open} onClose={props.onClose}>
        <DialogTitle>Update Device</DialogTitle>
        <DialogContent>
            <DialogContentText>
                Update registered Local Transcoder Device settings
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
                    value={data().name}
                    onChange={(e) => {
                        const value = e.target.value;
                        setData((prev) => {
                            return {
                                ...prev,
                                name: value
                            }
                        });
                    }}
                />
                <FormControl
                    variant="filled"
                    margin="dense"
                    sx={{
                        marginTop: '8px',
                        width: '100%',
                    }}>
                    <InputLabel id="select-filled-label">
                        Hardware Acceleration Types
                    </InputLabel>
                    <Select
                        id="demo-simple-select-filled"
                        labelId="select-filled-label"
                        fullWidth
                        variant="filled"
                        value={data().hardwareAccelerationType}
                        onChange={(e) => {
                            const value = e.target.value;
                            setData((prev) => {
                                return {
                                    ...prev,
                                    hardwareAccelerationType: value
                                }
                            });
                        }}
                    >
                        <MenuItem value={"cpu"}>
                            CPU
                        </MenuItem>
                        <MenuItem value={"vaapi"}>
                            VA_API
                        </MenuItem>
                        <MenuItem value={"quicksync"}>
                            Intel Quicksync
                        </MenuItem>
                    </Select>
                </FormControl>
                <FormControl
                    variant="filled"
                    margin="dense"
                    sx={{
                        marginTop: '8px',
                        width: '100%',
                    }}>
                    <InputLabel id="select-filled-label">
                        Log Level
                    </InputLabel>
                    <Select
                        id="demo-simple-select-filled"
                        labelId="select-filled-label"
                        fullWidth
                        variant="filled"
                        value={data().logLevel}
                        onChange={(e) => {
                            const value = e.target.value;
                            setData((prev) => {
                                return {
                                    ...prev,
                                    logLevel: value
                                }
                            });
                        }}
                    >
                        <MenuItem value={"info"}>
                            Info
                        </MenuItem>
                        <MenuItem value={"debug"}>
                            Debug
                        </MenuItem>
                        <MenuItem value={"error"}>
                            Error
                        </MenuItem>
                    </Select>
                </FormControl>
                <FormControlLabel
                    label="Edge TPU Enabled"
                    control={
                        <Checkbox
                            checked={data().edgeTpuEnabled} onChange={(e) => {
                                const value = e.target.checked;
                                setData((prev) => {
                                    return {
                                        ...prev,
                                        edgeTpuEnabled: value
                                    }
                                });
                            }}
                        />
                    }
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
                <Button onClick={handleUpdate} disabled={waitingUpdate()}>Submit</Button>
            </DialogActions>
        </DialogContent>
    </Dialog >
}
