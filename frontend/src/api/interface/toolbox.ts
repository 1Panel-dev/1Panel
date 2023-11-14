import { ReqPage } from '.';

export namespace Toolbox {
    export interface Fail2banBaseInfo {
        isEnable: boolean;
        isActive: boolean;
        version: string;

        port: number;
        maxRetry: number;
        banTime: string;
        findTime: string;
        banAction: string;
        logPath: string;
    }

    export interface Fail2banSearch extends ReqPage {
        status: string;
    }

    export interface Fail2banUpdate {
        key: string;
        value: string;
    }

    export interface Fail2banSet {
        ips: Array<string>;
        operate: string;
    }
}
