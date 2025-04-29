import manifest from '../manifest';

const {id: pluginId} = manifest;

export default {
    OPEN_MEETING: pluginId + '_open_meeting',
    CONFIG_RECEIVED: pluginId + '_config_received',
};
