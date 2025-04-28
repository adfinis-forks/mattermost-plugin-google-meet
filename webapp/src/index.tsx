// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import manifest from './manifest';

import {MattermostGoogleMeetPlugin} from '@/plugin/mattermost_google_meet_plugin';

window.registerPlugin(manifest.id, new MattermostGoogleMeetPlugin());
