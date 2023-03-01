import { CommonModel, ReqPage } from '.';

export namespace Host {
    export interface HostTree {
        id: number;
        label: string;
        children: Array<TreeNode>;
    }
    export interface TreeNode {
        id: number;
        label: string;
    }
    export interface Host extends CommonModel {
        name: string;
        groupBelong: string;
        addr: string;
        port: number;
        user: string;
        authMode: string;
        description: string;
    }
    export interface HostOperate {
        id: number;
        name: string;
        groupBelong: string;
        addr: string;
        port: number;
        user: string;
        authMode: string;
        privateKey: string;
        password: string;

        description: string;
    }
    export interface HostConnTest {
        addr: string;
        port: number;
        user: string;
        authMode: string;
        privateKey: string;
        password: string;
    }
    export interface GroupChange {
        id: number;
        group: string;
    }
    export interface ReqSearch {
        info?: string;
    }
    export interface SearchWithPage extends ReqPage {
        group: string;
        info?: string;
    }
}
