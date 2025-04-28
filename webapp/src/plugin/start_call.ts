// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Dispatch} from 'react';
import type {Action} from 'redux';

import {createPost} from 'mattermost-redux/actions/posts';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/common';
import type {Channel} from 'mattermost-redux/types/channels';
import type {Post} from 'mattermost-redux/types/posts';
import type {GlobalState} from 'mattermost-redux/types/store';

import {GOOGLE_MEET_MESSAGE} from '@/plugin/constant';

export const startCall = (channel: Channel) => {
    return async (dispatch: Dispatch<Action<Record<string, unknown>>>, getState: () => GlobalState) => {
        const state = getState();

        const team = state.entities.teams.teams[state.entities.teams.currentTeamId];
        const callName = `${team.name}-${channel.name}`; // = uuidv4();
        const trimmedCallName = callName.substring(0, 60);
        const url = `http://g.co/meet/${trimmedCallName}`;

        // Open a window?
        // window.open(url);

        const post: Post = {
            create_at: Date.now(),
            user_id: getCurrentUserId(state),
            channel_id: channel.id,
            message: `I have started a meeting: [${url}](${url})`,
            type: GOOGLE_MEET_MESSAGE as any,
            props: {
                call_name: trimmedCallName,
                meeting_link: url,
            },
        } as any;

        // Based on https://zenn.dev/kaakaa/articles/qiita-20201220-fd10c58b00c43ae3cc3c
        return dispatch(createPost(post as any) as any);
    };
};
