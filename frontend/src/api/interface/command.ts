import { ReqPage } from '.';

export namespace Command {
    export interface CommandInfo {
        id: number;
        name: string;
        groupID: number;
        command: string;
    }
    export interface CommandOperate {
        id: number;
        name: string;
        groupID: number;
        command: string;
    }
    export interface CommandSearch extends ReqPage {
        info: string;
    }
}
