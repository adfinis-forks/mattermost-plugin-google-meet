// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import type {Store, Action} from 'redux';

import {getCurrentUserLocale} from 'mattermost-redux/selectors/entities/i18n';
import type {GlobalState} from 'mattermost-redux/types/store';

import {startCall} from './start_call';

import type {PluginRegistry, Plugin} from '../types/mattermost-webapp';

import {HeaderMessage} from '@/component/header';
import {I18nProvider} from '@/component/i18n_provider';
import Icon from '@/component/icon';
import {PostTypeGoogleMeet} from '@/component/post_type_google_meet';
import {GOOGLE_MEET_MESSAGE} from '@/plugin/constant';
import {getTranslations} from '@/plugin/translation';

export class MattermostGoogleMeetPlugin implements Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        const locale = getCurrentUserLocale(store.getState());

        registry.registerTranslations((locale) => getTranslations(locale));
        registry.registerChannelHeaderButtonAction(
            <Icon/>,
            (channel) => startCall(channel)(store.dispatch, store.getState),
            <I18nProvider currentLocale={locale}><HeaderMessage/></I18nProvider>,
        );

        registry.registerPostTypeComponent(
            GOOGLE_MEET_MESSAGE,
            (props) => (<I18nProvider currentLocale={locale}>
                <PostTypeGoogleMeet
                    theme={props.theme}
                    post={props.post}
                />
            </I18nProvider>),
        );

        // Maybe in future
        // registry.registerSlashCommandWillBePostedHook(
        //     (message, args) => {
        //         if (message.startsWith('/meet')) {
        //             this.startCall(args.channel_id)(store.dispatch, store.getState)
        //             return {error: {message: 'rejected'}};
        //         }
        //     }
        // );
    }

    public async uninitialize() {}
}
