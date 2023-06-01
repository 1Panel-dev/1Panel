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
        groupID: number;
        groupBelong: string;
        addr: string;
        port: number;
        user: string;
        authMode: string;
        password: string;
        privateKey: string;
        passPhrase: string;
        rememberPassword: boolean;
        description: string;
    }
    export interface HostOperate {
        id: number;
        name: string;
        groupID: number;
        addr: string;
        port: number;
        user: string;
        authMode: string;
        password: string;
        privateKey: string;
        passPhrase: string;
        rememberPassword: boolean;

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
        groupID: number;
    }
    export interface ReqSearch {
        info?: string;
    }
    export interface SearchWithPage extends ReqPage {
        groupID: number;
        info?: string;
    }

    export interface FirewallBase {
        name: string;
        status: string;
        version: string;
        pingStatus: string;
    }
    export interface RuleSearch extends ReqPage {
        info: string;
        type: string;
    }
    export interface RuleInfo extends ReqPage {
        family: string;
        address: string;
        port: string;
        protocol: string;
        strategy: string;
        appName: string;
        isUsed: boolean;
    }
    export interface RulePort {
        operation: string;
        address: string;
        port: string;
        source: string;
        protocol: string;
        strategy: string;
    }
    export interface RuleIP {
        operation: string;
        address: string;
        strategy: string;
    }
    export interface UpdatePortRule {
        oldRule: RulePort;
        newRule: RulePort;
    }
    export interface UpdateAddrRule {
        oldRule: RuleIP;
        newRule: RuleIP;
    }
    export interface BatchRule {
        type: string;
        rules: Array<RulePort>;
    }

    export interface SSHInfo {
        status: string;
        message: string;
        port: string;
        listenAddress: string;
        passwordAuthentication: string;
        pubkeyAuthentication: string;
        encryptionMode: string;
        primaryKey: string;
        permitRootLogin: string;
        useDNS: string;
    }
    export interface SSHGenerate {
        encryptionMode: string;
        password: string;
    }
    export interface searchSSHLog extends ReqPage {
        info: string;
        status: string;
    }
    export interface sshLog {
        logs: Array<sshHistory>;
        successfulCount: number;
        failedCount: number;
    }
    export interface sshHistory {
        date: Date;
        area: string;
        user: string;
        authMode: string;
        address: string;
        port: string;
        status: string;
        message: string;
    }
}
