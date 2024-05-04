import createDebounce from "@solid-primitives/debounce";
import { createFileUploader } from "@solid-primitives/upload";
import { Add, Delete, Face, History, MoreVert, Refresh, UploadFile } from "@suid/icons-material";
import { Alert, Box, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, IconButton, Link, ListItemIcon, ListItemText, Menu, MenuItem, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from "@suid/material";
import { Component, Match, Show, Switch, createResource, createSignal } from "solid-js";
import { AddDetectablePerson, addDetectablePerson, deletePerson, getPersonHistory } from "../../clients/backend/people";
import { FrontPagePersonInfo, getListPeople } from "../../helper/helper";

export const PeoplePage: Component = () => {
    const [peopleIds, setPeopleIds] = createSignal<string[]>([]);
    const [menuAnchorEl, setMenuAnchorEl] = createSignal<HTMLElement | null>(null);
    const [menuDialogCurrentItem, setMenuDialogCurrentItem] = createSignal<FrontPagePersonInfo | null>(null);
    const [menuCurrentlySelectedId, setMenuCurrentlySelectedId] = createSignal<string>("");

    const [deleteDialogStatus, setDeleteDialogStatus] = createSignal<boolean>(false);
    const [imageDialogStatus, setImageDialogStatus] = createSignal<boolean>(false);
    const [historyDialogStatus, setHistoryDialogStatus] = createSignal<boolean>(false);

    const openMenu = () => Boolean(menuAnchorEl());
    const closeMenu = () => setMenuAnchorEl(null);

    const [addPersonDialogStatus, setAddPersonDialog] = createSignal(false);

    const filterDebouncer = createDebounce((m: string) => {
        if (m === "") {
            setPeopleIds([]);
        } else {
            setPeopleIds(m.split(","));
        }
    }, 350);

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
            id: 'face',
            name: 'Face',
            icon: <Face />,
            onClick: () => {
                setImageDialogStatus(true);
            }
        },
        {
            id: 'history',
            name: "History",
            icon: <History />,
            onClick: () => {
                setHistoryDialogStatus(true);
            }
        }
    ]

    const [people, { refetch: fetchPeople }] = createResource(peopleIds, getListPeople);
    return <div class="w-full h-full p-8">
        <Paper class="h-max w-full">
            <div class="flex flex-row justify-between items-center p-4">
                <Typography variant="h6" component="h1">People</Typography>
                <div class="flex flex-row items-end gap-2">
                    <TextField id="person-id-filter" label="Person ID" variant="standard" size="small" color="primary" margin="none" sx={{ marginRight: '1rem' }} onChange={(e) => {
                        const value = e.target.value;
                        filterDebouncer(value);
                    }} />
                    <Button variant="contained" size="medium" color="primary" startIcon={<Add />} onClick={() => {
                        setAddPersonDialog(true);
                    }}>
                        Add Person
                    </Button>
                    <Button variant="contained" size="medium" color="primary" startIcon={<Refresh />} onClick={() => {
                        fetchPeople();
                    }}>
                        Refresh
                    </Button>
                </div>
            </div>
            <Show when={people.loading}>
                <div class="flex flex-row justify-center align-middle w-full">
                    <CircularProgress color="primary" sx={{ margin: '2rem' }} />
                </div>
            </Show>
            <Switch>
                <Match when={people.error}>
                    <p>Error when retrieving data: {people.error}</p>
                </Match>
                <Match when={people() && !people.loading}>
                    <Box sx={{
                        display: "flex",
                        flexWrap: "wrap",
                    }}>
                        <TableContainer>
                            <Table aria-label="person-list">
                                <TableHead>
                                    <TableRow>
                                        <TableCell sx={{ width: '20%' }}>Name</TableCell>
                                        <TableCell align="left">ID</TableCell>
                                        <TableCell align="left">Age</TableCell>
                                        <TableCell align="center"></TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {people()?.items.map((person, _) => {
                                        return <>
                                            <TableRow sx={{ "&:last-child td, &:last-child th": { border: 0 } }}>
                                                <TableCell align="left">
                                                    <Link variant="body2" underline="none">
                                                        {person.person.name}
                                                    </Link>
                                                </TableCell>
                                                <TableCell align="left">
                                                    {person.person.personId}
                                                </TableCell>
                                                <TableCell align="left">{person.person.age}</TableCell>
                                                <TableCell align="center">
                                                    <IconButton
                                                        aria-label="more"
                                                        id="long-button"
                                                        aria-controls={openMenu() ? 'long-menu' : undefined}
                                                        aria-expanded={openMenu() ? 'true' : undefined}
                                                        aria-haspopup="true"
                                                        onClick={(e) => {
                                                            setMenuDialogCurrentItem(person);
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
                addPersonDialogStatus() &&
                <AddPersonDialog
                    open={addPersonDialogStatus()}
                    onClose={() => {
                        setAddPersonDialog(false);
                        closeMenu();
                        fetchPeople();
                    }}
                />
            }
            {
                deleteDialogStatus() &&
                <DeletePersonConfirmDialog
                    personId={menuDialogCurrentItem()!.person.personId}
                    open={deleteDialogStatus()}
                    onClose={() => {
                        setDeleteDialogStatus(false);
                        closeMenu();
                        fetchPeople();
                    }}
                />
            }
            {
                imageDialogStatus() &&
                <PersonImageDialog
                    item={menuDialogCurrentItem()!}
                    open={imageDialogStatus()}
                    onClose={() => {
                        setImageDialogStatus(false);
                        closeMenu();
                    }}
                />
            }
            {
                historyDialogStatus() &&
                <PersonHistoryDialog
                    item={menuDialogCurrentItem()!}
                    open={historyDialogStatus()}
                    onClose={() => {
                        setHistoryDialogStatus(false);
                        closeMenu();
                    }}
                />
            }
        </Paper>
    </div>
}

interface AddPersonDialogProps {
    open: boolean;
    onClose: () => void;
}

const AddPersonDialog = (props: AddPersonDialogProps) => {
    const [waitingAdd, setWaitingAdd] = createSignal<boolean>(false);
    const [formErr, setFormErr] = createSignal<string | null>(null);

    const { files, selectFiles } = createFileUploader();
    const [data, setData] = createSignal<AddDetectablePerson>({
        name: "",
        age: "",
        base64Image: "",
    });
    const handleAdd = async () => {
        setWaitingAdd(true);
        try {
            await addDetectablePerson(data());
        } catch (e: any) {
            setFormErr(e.message);
            setWaitingAdd(false);
            return;
        }
        setWaitingAdd(false);
        props.onClose();
    }
    const getBase64StringFromDataURL = (dataURL: string) =>
        dataURL.replace('data:', '').replace(/^.+,/, '');

    return <Dialog open={props.open} onClose={props.onClose}>
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
                <TextField
                    margin="dense"
                    id="age"
                    label="Age"
                    type="number"
                    variant="filled"
                    value={data().age}
                    onChange={(e) => {
                        const value = e.target.value;
                        setData((prev) => {
                            return {
                                ...prev,
                                age: value
                            }
                        });
                    }}
                    fullWidth
                />
                <div class="flex flex-row justify-start items-center gap-4">
                    <Button variant="contained" size="medium" color="primary" sx={{ marginTop: '0.75rem' }} startIcon={<UploadFile />} onClick={() => {
                        selectFiles(async (files) => {
                            if (files.length > 1) {
                                setFormErr("Too many files selected");
                                return;
                            }
                            const f = files[0];
                            const typ = f.file.type;
                            if (typ !== "image/jpeg") {
                                setFormErr("Image is not JPEG");
                                return;
                            }
                            var reader = new FileReader();
                            reader.onloadend = function () {
                                const res = reader.result;
                                const encoded = getBase64StringFromDataURL(res as string);
                                setData((prev: AddDetectablePerson) => {
                                    return {
                                        ...prev,
                                        base64Image: encoded,
                                    }
                                });
                            }
                            reader.readAsDataURL(f.file);

                        });
                    }}>
                        Upload Face Image
                    </Button>
                    <Typography variant="body2">
                        {files().length > 0 ? files()[0].name : ""}
                    </Typography>
                </div>
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
}


interface DeletePersonConfirmDialogProps {
    personId: string;
    open: boolean;
    onClose: () => void;
}

const DeletePersonConfirmDialog = (props: DeletePersonConfirmDialogProps) => {
    const [waitingDelete, setWaitingDelete] = createSignal<boolean>(false);
    const handleDelete = async () => {
        setWaitingDelete(true);
        await deletePerson(props.personId);
        setWaitingDelete(false);
        props.onClose();
    }

    return <Dialog onClose={props.onClose} open={props.open || waitingDelete()}>
        <DialogTitle id="alert-dialog-title">
            Delete camera?
        </DialogTitle>
        <DialogContent>
            <DialogContentText id="alert-dialog-description">
                {`Requested to delete person with ID: ${props.personId}, proceed?`}
            </DialogContentText>
        </DialogContent>
        <DialogActions>
            <Button onClick={props.onClose}>Disagree</Button>
            <Button onClick={handleDelete} disabled={waitingDelete()}>Agree</Button>
        </DialogActions>
    </Dialog>
}

interface PersonImageDialogProps {
    item: FrontPagePersonInfo;
    onClose: () => void;
    open: boolean;
}

const PersonImageDialog = (props: PersonImageDialogProps) => {
    console.log(props.item);
    return <Dialog onClose={props.onClose} open={props.open}>
        <DialogTitle>Person Information</DialogTitle>
        <DialogContent>
            <DialogContentText>{props.item.person.name}</DialogContentText>
            <DialogContent>
                <img src={props.item.image.presignedUrl} alt="face" />
            </DialogContent>
        </DialogContent>
    </Dialog>
}


interface PersonHistoryDialogProps {
    item: FrontPagePersonInfo;
    onClose: () => void;
    open: boolean;
}

const PersonHistoryDialog = (props: PersonHistoryDialogProps) => {
    const [history, refetch] = createResource([props.item.person.personId], getPersonHistory);
    return <Dialog onClose={props.onClose} open={props.open}>
        <DialogTitle>Person History</DialogTitle>
        <DialogContent>
            <DialogContentText>History</DialogContentText>
            <DialogContent>
                <TableContainer component={Paper}>
                    <Table aria-label="history-table">
                        <TableHead>
                            <TableRow>
                                <TableCell>History ID</TableCell>
                                <TableCell align="right">Timestamp</TableCell>
                                <TableCell align="right">Event ID</TableCell>
                                <TableCell align="right">Person ID</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {history()?.map((item: any) => (
                                <TableRow>
                                    <TableCell component="th" scope="row">
                                        {item.historyId}
                                    </TableCell>
                                    <TableCell align="right">{item.timestamp}</TableCell>
                                    <TableCell align="right">{item.eventId}</TableCell>
                                    <TableCell align="right">{item.personId}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
            </DialogContent>
        </DialogContent>
    </Dialog>
}