import {combineReducers} from 'redux';

import type {Post} from 'mattermost-redux/types/posts';

import ActionTypes from '../action_types';

function openMeeting(state: Post | null = null, action: {type: string; data: {post: Post | null; jwt: string | null}}) {
    switch (action.type) {
    case ActionTypes.OPEN_MEETING:
        return action.data.post;
    default:
        return state;
    }
}

function config(state: object = {}, action: {type: string; data: object}) {
    switch (action.type) {
    case ActionTypes.CONFIG_RECEIVED:
        return action.data;
    default:
        return state;
    }
}

export default combineReducers({
    openMeeting,
    config,
});
