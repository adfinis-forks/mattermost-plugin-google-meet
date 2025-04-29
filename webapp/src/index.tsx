import manifest from './manifest';

import {MattermostGoogleMeetPlugin} from '@/plugin/mattermost_google_meet_plugin';

window.registerPlugin(manifest.id, new MattermostGoogleMeetPlugin());
