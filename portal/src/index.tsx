import { Route, Router } from '@solidjs/router';
import { render } from 'solid-js/web';
import './index.css';
import { App } from './pages/App';
import { CamerasPage } from './pages/cameras/CamerasPage';
import { CameraViewerPage } from './pages/cameras/ViewerPage';
import { EventsPage } from './pages/events/EventsPage';
import { GroupsPage } from './pages/groups/GroupsPage';
import { PeoplePage } from './pages/people/PeoplePage';
import { TranscoderPage } from './pages/transcoders/Transcoder';

const root = document.getElementById('root');

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
  throw new Error(
    'Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?',
  );
}

render(() => <Router root={App}>
  <Route path="/" component={CamerasPage} />
  <Route path="/people" component={PeoplePage} />
  <Route path="/groups" component={GroupsPage} />
  <Route path="/events" component={EventsPage} />
  <Route path="/devices" component={TranscoderPage} />
  <Route path="/cameras/:cameraId" component={CameraViewerPage} />
</Router>, root!);
