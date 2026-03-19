import {Client4, ClientError} from '@mattermost/client';

import manifest from '../manifest';

export default class Client {
    private serverUrl: string | undefined;
    private url: string | undefined;
    private client: Client4;

    constructor() {
        this.client = new Client4();
    }

    setServerRoute(url: string) {
        this.serverUrl = url;
        this.url = url + '/plugins/' + manifest.id;
    }

    startMeeting = async (
        channelId: string,
        personal: boolean = false,
        topic: string = '',
        meetingId: string = '',
    ) => {
        return this.doPost(`${this.url}/api/v1/meetings`, {
            channel_id: channelId,
            personal,
            topic,
            meeting_id: meetingId,
        });
    };

    loadConfig = async () => {
        return this.doPost(`${this.url}/api/v1/config`, {});
    };

    doPost = async (
        url: string,
        body: unknown,
        headers: Record<string, string> = {},
    ) => {
        const options = {
            method: 'post',
            body: JSON.stringify(body),
            headers,
        };

        const response = await fetch(url, this.client.getOptions(options));

        if (response.ok) {
            return response.json();
        }

        const text = await response.text();

        throw new ClientError(this.client.getUrl(), {
            message: text || '',
            status_code: response.status,
            url,
        });
    };
}
