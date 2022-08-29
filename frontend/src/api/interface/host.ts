import { CommonModel, ReqPage } from '.';

export namespace Host {
    export interface Host extends CommonModel {
        name: string;
        addr: string;
        port: number;
        user: string;
        authMode: string;
        description: string;
    }
    export interface HostOperate {
        id: number;
        name: string;
        addr: string;
        port: number;
        user: string;
        authMode: string;
        privateKey: string;
        password: string;

        description: string;
    }
    export interface ReqSearchWithPage extends ReqPage {
        info?: string;
    }
}
