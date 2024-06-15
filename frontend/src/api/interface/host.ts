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
        status: string;
        strategy: string;
        info: string;
        type: string;
    }
    export interface RuleInfo extends ReqPage {
        family: string;
        address: string;
        destination: string;
        port: string;
        srcPort: string;
        destPort: string;
        protocol: string;
        strategy: string;

        usedStatus: string;
        description: string;

        [key: string]: any;
    }
    export interface UpdateDescription {
        address: string;
        port: string;
        protocol: string;
        strategy: string;
        description: string;
    }
    export interface RulePort {
        operation: string;
        address: string;
        port: string;
        source: string;
        protocol: string;
        strategy: string;
        description: string;
    }
    export interface RuleForward {
        operation: string;
        protocol: string;
        port: string;
        targetIP: string;
        targetPort: string;
    }
    export interface RuleIP {
        operation: string;
        address: string;
        strategy: string;
        description: string;
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

    export interface MonitorData {
        param: string;
        date: Array<Date>;
        value: Array<any>;
    }
    export interface MonitorSearch {
        param: string;
        info: string;
        startTime: Date;
        endTime: Date;
    }

    export interface SSHInfo {
        autoStart: boolean;
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
    export interface SSHUpdate {
        key: string;
        oldValue: string;
        newValue: string;
    }
    export interface SSHGenerate {
        encryptionMode: string;
        password: string;
    }
    export interface searchSSHLog extends ReqPage {
        info: string;
        status: string;
    }
    export interface analysisSSHLog extends ReqPage {
        orderBy: string;
    }
    export interface logAnalysisRes {
        total: number;
        items: Array<logAnalysis>;
        successfulCount: number;
        failedCount: number;
    }
    export interface sshLog {
        logs: Array<sshHistory>;
        successfulCount: number;
        failedCount: number;
    }
    export interface logAnalysis {
        address: string;
        area: string;
        successfulCount: number;
        failedCount: number;
        status: string;
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
