import createDebounce from "@solid-primitives/debounce";
import { createFileUploader } from "@solid-primitives/upload";
import { Add, MoreVert, Refresh, UploadFile } from "@suid/icons-material";
import { Alert, Box, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, IconButton, Link, Menu, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Typography } from "@suid/material";
import { Component, Match, Show, Switch, createResource, createSignal } from "solid-js";
import { AddDetectablePerson, addDetectablePerson } from "../../clients/backend/people";
import { FrontPagePersonInfo, getListPeople } from "../../helper/helper";

export const PeoplePage: Component = () => {
    const [peopleIds, setPeopleIds] = createSignal<string[]>([]);
    const [menuAnchorEl, setMenuAnchorEl] = createSignal<HTMLElement | null>(null);
    const [menuDialogCurrentItem, setMenuDialogCurrentItem] = createSignal<FrontPagePersonInfo | null>(null);

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
    const [people, { refetch: fetchPeople }] = createResource(peopleIds, getListPeople);
    return <div class="w-full h-full p-8">
        <Paper class="h-max w-full">
            <div class="flex flex-row justify-between items-center p-4">
                <Typography variant="h6" component="h1">People</Typography>
                <div class="flex flex-row items-end gap-2">
                    <TextField id="person-id-filter" label="Camera ID" variant="standard" size="small" color="primary" margin="none" sx={{ marginRight: '1rem' }} onChange={(e) => {
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
                                        <TableCell align="center">Status</TableCell>
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
                                                    <Link variant="body2" underline="none" href={`/person/${person.person.personId}`}>
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
                                                        aria-controls={open() ? 'long-menu' : undefined}
                                                        aria-expanded={open() ? 'true' : undefined}
                                                        aria-haspopup="true"
                                                        onClick={(e) => {

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
                                                            closeMenu();
                                                        }}
                                                        PaperProps={{
                                                            elevation: 2,
                                                        }}
                                                    >

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