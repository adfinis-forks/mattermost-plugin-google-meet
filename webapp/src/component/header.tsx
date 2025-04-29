import React from 'react';
import {FormattedMessage} from 'react-intl';

// export const HeaderMessage = () => <FormattedMessage id='mattermost_meet_plugin.header.label'/>;

export default class HeaderMessage extends React.PureComponent {
    render() {
        return (
            <FormattedMessage id='mattermost_meet_plugin.header.label'/>
        );
    }
}
