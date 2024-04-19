import { Box, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography } from "@suid/material";
import { Component, For, Match, Show, Switch, createResource, createSignal } from "solid-js";
import { getCameras } from "../../clients/backend/client";

export const CamerasPage: Component = () => {
    const [cameraIds, setCamerasIds] = createSignal([]);
    const [cameras] = createResource(cameraIds, getCameras);

    return <div class="w-full h-full">
        <Show when={cameras.loading}>
            <p>Loading...</p>
        </Show>
        <Switch>
            <Match when={cameras.error}>
                <p>Error loading cameras: {cameras.error}</p>
            </Match>
            <Match when={cameras()}>
                <Box sx={{
                    display: "flex",
                    flexWrap: "wrap",
                }}>
                    <Paper class="h-full w-full m-8">
                        <Typography variant="h6" component="h1" marginLeft={2} marginTop={2}>Cameras</Typography>
                        <TableContainer>
                            <Table aria-label="camera-list">
                                <TableHead>
                                    <TableRow>
                                        <TableCell >Name</TableCell>
                                        <TableCell align="left">ID</TableCell>
                                        <TableCell align="left">Enabled</TableCell>
                                        <TableCell align="left">IP Address</TableCell>
                                        <TableCell align="center">Details</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    <For each={cameras()}>{camera => (
                                        <TableRow sx={{ "&:last-child td, &:last-child th": { border: 0 } }}>

                                            <TableCell>{camera.name}</TableCell>
                                            <TableCell align="left">
                                                {camera.cameraId}
                                            </TableCell>
                                            <TableCell align="left">{camera.enabled === true ? "Yes" : "No"}</TableCell>
                                            <TableCell align="left">{camera.ip}</TableCell>
                                            <TableCell align="center"></TableCell>
                                        </TableRow>
                                    )}
                                    </For>
                                </TableBody>
                            </Table>
                        </TableContainer>
                    </Paper>

                </Box>
            </Match>
        </Switch>
    </div>
}