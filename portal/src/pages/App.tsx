import { A } from "@solidjs/router";
import { CameraFront, EventAvailable, GroupOutlined, PeopleAlt } from "@suid/icons-material";
import MenuIcon from "@suid/icons-material/Menu";
import SourceIcon from "@suid/icons-material/Source";
import {
    AppBar,
    Box,
    Drawer,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Toolbar,
    Typography
} from "@suid/material";
import { ParentComponent } from "solid-js";
import { createMutable } from "solid-js/store";

export const App: ParentComponent = (props) => {
    const state = createMutable({
        drawer: false,
        currentPageId: 'cameras'
    })

    const toggleDrawer = (open: boolean) => (event: MouseEvent | KeyboardEvent) => {
        if (event.type == "keydown") {
            const kbEvent = event as KeyboardEvent;
            if (kbEvent.key === "Tab" || kbEvent.key === "Shift") {
                return;
            }
        }
        state.drawer = open;
    }

    const setCurrentPage = (page: string) => (event: MouseEvent) => {
        state.currentPageId = page;
    }

    const itemList = [
        {
            id: 'cameras',
            name: "Cameras",
            title: "Portal - Cameras",
            icon: <CameraFront />,
            route: "/"
        },
        {
            id: 'people',
            name: "People",
            title: "Portal - People",
            icon: <PeopleAlt />,
            route: "/people",
        },
        {
            id: 'groups',
            name: "Groups",
            title: "Portal - Groups",
            icon: <GroupOutlined />,
            route: "/groups",
        },
        {
            id: 'events',
            name: "Events",
            title: "Portal - Events",
            icon: <EventAvailable />,
            route: "/events"
        },
    ]

    const items = () => <Box sx={{ width: 250 }} role="presentation" onClick={toggleDrawer(false)} onKeyDown={toggleDrawer(true)}>
        <List>
            {itemList.map((item, _) => <ListItem disablePadding={true}>
                <ListItemButton selected={state.currentPageId === item.id} onClick={setCurrentPage(item.id)} component={A} href={item.route}>
                    <ListItemIcon>{item.icon}</ListItemIcon>
                    <ListItemText primary={item.name}></ListItemText>
                </ListItemButton>
            </ListItem>)}
        </List>
    </Box>

    return <div class="flex flex-col h-full">
        <div class="flex">
            <Box sx={{ flexGrow: 1 }}>
                <AppBar position="static">
                    <Toolbar>
                        <IconButton
                            size="large" edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }} onClick={toggleDrawer(!state.drawer)}>
                            <MenuIcon />
                        </IconButton>
                        <Typography variant="h6" component={"div"} sx={{ flexGrow: 1 }}>
                            Portal
                        </Typography>
                        <IconButton color="inherit" aria-label="github-ref" size="small" >
                            <SourceIcon />
                        </IconButton>
                    </Toolbar>
                </AppBar>
            </Box>
            <Drawer anchor="left" open={state.drawer} onClose={toggleDrawer(false)}>{items()}</Drawer>
        </div>
        <main class="flex-1">
            {props.children}
        </main>
    </div>
}