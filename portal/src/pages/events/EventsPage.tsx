import createDebounce from "@solid-primitives/debounce";
import { Face, MoreVert, Refresh } from "@suid/icons-material";
import { Box, Button, CircularProgress, Dialog, DialogContent, DialogContentText, DialogTitle, IconButton, Link, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from "@suid/material";
import { Component, Match, Show, Switch, createResource, createSignal } from "solid-js";
import { SummarizedEvent, getListEvents } from "../../helper/helper";

export const EventsPage: Component = () => {
    const [eventIds, setEventIds] = createSignal<string[]>([]);
    const [menuAnchorEl, setMenuAnchorEl] = createSignal<HTMLElement | null>(null);
    const [menuDialogCurrentItem, setMenuDialogCurrentItem] = createSignal<SummarizedEvent | null>(null);
    const [menuCurrentlySelectedId, setMenuCurrentlySelectedId] = createSignal<string>("");

    const [imageDialogStatus, setImageDialogStatus] = createSignal<boolean>(false);

    const openMenu = () => Boolean(menuAnchorEl());
    const closeMenu = () => setMenuAnchorEl(null);

    const filterDebouncer = createDebounce((m: string) => {
        if (m === "") {
            setEventIds([]);
        } else {
            setEventIds(m.split(","));
        }
    }, 350);

    const moreMenuItems = [
        {
            id: 'snapshot',
            name: 'View Image',
            shouldAppear: (item: SummarizedEvent) => true,
            icon: < Face />,
            onClick: () => {
                setImageDialogStatus(true);
            }
        },
    ]

    const [events, { refetch: fetchEvents }] = createResource(eventIds, (ids: string[]) => getListEvents(ids, 99));
    return <div class="w-full h-full p-8">
        <Paper class="h-max w-full">
            <div class="flex flex-row justify-between items-center p-4">
                <Typography variant="h6" component="h1">People</Typography>
                <div class="flex flex-row items-end gap-2">
                    <TextField id="person-id-filter" label="Person ID" variant="standard" size="small" color="primary" margin="none" sx={{ marginRight: '1rem' }} onChange={(e) => {
                        const value = e.target.value;
                        filterDebouncer(value);
                    }} />
                    <Button variant="contained" size="medium" color="primary" startIcon={<Refresh />} onClick={() => {
                        fetchEvents();
                    }}>
                        Refresh
                    </Button>
                </div>
            </div>
            <Show when={events.loading}>
                <div class="flex flex-row justify-center align-middle w-full">
                    <CircularProgress color="primary" sx={{ margin: '2rem' }} />
                </div>
            </Show>
            <Switch>
                <Match when={events.error}>
                    <p>Error when retrieving data: {events.error}</p>
                </Match>
                <Match when={events() && !events.loading}>
                    <Box sx={{
                        display: "flex",
                        flexWrap: "wrap",
                    }}>
                        <TableContainer>
                            <Table aria-label="person-list">
                                <TableHead>
                                    <TableRow>
                                        <TableCell sx={{ width: '20%' }}>ID</TableCell>
                                        <TableCell align="left">Start Time</TableCell>
                                        <TableCell align="left">End Time</TableCell>
                                        <TableCell align="center">Is Ongoing?</TableCell>
                                        <TableCell align="center">Camera</TableCell>
                                        <TableCell align="center">Label</TableCell>
                                        <TableCell align="center">Score</TableCell>
                                        <TableCell align="left">Person</TableCell>
                                        <TableCell></TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {events()?.map((event, _) => {
                                        return <>
                                            <TableRow sx={{ "&:last-child td, &:last-child th": { border: 0 } }}>
                                                <TableCell align="left">
                                                    <Link variant="body2" underline="none">
                                                        {event.event.eventId}
                                                    </Link>
                                                </TableCell>
                                                <TableCell align="left">
                                                    {event.event.startTime}
                                                </TableCell>
                                                <TableCell align="left">{event.event.endTime}</TableCell>
                                                <TableCell align="center">{event.event.endTime ? "No" : "Yes"}</TableCell>
                                                <TableCell align="center">{event.event.CameraName}</TableCell>
                                                <TableCell align="center">{event.event.label}</TableCell>
                                                <TableCell align="center">{event.event.score}</TableCell>
                                                <TableCell align="left">
                                                    {event.person?.name ?? "N/A"}
                                                </TableCell>
                                                <TableCell align="center">
                                                    <IconButton
                                                        aria-label="more"
                                                        id="long-button"
                                                        aria-controls={openMenu() ? 'long-menu' : undefined}
                                                        aria-expanded={openMenu() ? 'true' : undefined}
                                                        aria-haspopup="true"
                                                        onClick={(e) => {
                                                            setMenuDialogCurrentItem(event);
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
                                                            setMenuDialogCurrentItem(null);
                                                            setMenuCurrentlySelectedId("");
                                                            closeMenu();
                                                        }}
                                                        PaperProps={{
                                                            elevation: 2,
                                                        }}
                                                    >
                                                        {moreMenuItems.map((item, _) => {
                                                            if (item.shouldAppear(event) === false) {
                                                                return <></>;
                                                            }
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
                imageDialogStatus() &&
                <EventSnapshotDialog
                    item={menuDialogCurrentItem()!}
                    open={imageDialogStatus()}
                    onClose={() => {
                        setImageDialogStatus(false);
                        closeMenu();
                    }}
                />
            }
        </Paper>
    </div>
}


interface EventSnapshotDialogProps {
    item: SummarizedEvent;
    onClose: () => void;
    open: boolean;
}

const EventSnapshotDialog = (props: EventSnapshotDialogProps) => {
    return <Dialog onClose={props.onClose} open={props.open}>
        <DialogTitle>Event Information</DialogTitle>
        <DialogContent>
            {
                props.item.person !== undefined ? <>
                    <DialogContentText>
                        Name: {props.item.person?.name}</DialogContentText>
                    <DialogContentText>Age: {props.item.person?.age}</DialogContentText>
                </> : <></>
            }
            <DialogContent>
                <div class="flex flex-col gap-2">
                    <Typography variant="h6" component="h1">Snapshot</Typography>
                    <img src={props.item.presignedUrl} alt="snapshot" />
                </div>
            </DialogContent>
        </DialogContent>
    </Dialog>
}
