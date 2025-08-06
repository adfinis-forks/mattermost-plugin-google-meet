import React from 'react';
import type {Store, Action} from 'redux';

import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {getCurrentUserLocale} from 'mattermost-redux/selectors/entities/i18n';
import type {GlobalState} from 'mattermost-redux/types/store';

import {loadConfig, startMeeting} from './actions';
import {GOOGLE_MEET_MESSAGE} from './constant';
import {getTranslations} from './translation';

import Client from '../client';
import HeaderMessage from '../component/header';
import {I18nProvider} from '../component/i18n_provider';
import Icon from '../component/icon';
import {PostTypeGoogleMeet} from '../component/post_type_google_meet';
import manifest from '../manifest';
import reducer from '../reducers';
import type {PluginRegistry, Plugin} from '../types/mattermost-webapp';

export class MattermostGoogleMeetPlugin implements Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/

        registry.registerReducer(reducer);

        const locale = getCurrentUserLocale(store.getState());
        registry.registerTranslations((locale) => getTranslations(locale));
        registry.registerChannelHeaderButtonAction(
            <Icon/>,
            (channel) => {
                store.dispatch(startMeeting(channel.id) as any);
            },
            <I18nProvider currentLocale={locale}><HeaderMessage/></I18nProvider>,
        );

        registry.registerAppBarComponent(
            getPluginURL(store.getState()) + '/public/app-bar-icon.png',
            (channel) => {
                store.dispatch(startMeeting(channel.id) as any);
            },
            <I18nProvider currentLocale={locale}><HeaderMessage/></I18nProvider>,
        );

        Client.setServerRoute(getServerRoute(store.getState()));
        registry.registerPostTypeComponent(
            GOOGLE_MEET_MESSAGE,
            (props) => (<I18nProvider currentLocale={locale}>
                <PostTypeGoogleMeet
                    post={props.post}
                    theme={props.theme}
                />
            </I18nProvider>),
        );
        registry.registerWebSocketEventHandler('custom_gmeet_config_update', () => store.dispatch(loadConfig() as any));
        store.dispatch(loadConfig() as any);
    }

    public async uninitialize() {}
}

function getServerRoute(state: GlobalState) {
    const config = getConfig(state);
    let basePath = '';
    if (config && config.SiteURL) {
        basePath = config.SiteURL;
        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }
    return basePath;
}

function getPluginURL(state: GlobalState) {
    const siteURL = getServerRoute(state);
    return siteURL + '/plugins/' + manifest.id;
}
