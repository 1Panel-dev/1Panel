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
        isLocal: boolean;
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
        isLocal: boolean;
        addr: string;
        port: number;
        user: string;
        authMode: string;
        privateKey: string;
        passPhrase: string;
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
        isExist: boolean;
        isActive: boolean;
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
        interface: string;
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

    export interface MonitorSetting {
        defaultNetwork: string;
        monitorStatus: string;
        monitorStoreDays: string;
        monitorInterval: string;
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
        isActive: boolean;
        message: string;
        port: string;
        listenAddress: string;
        passwordAuthentication: string;
        pubkeyAuthentication: string;
        encryptionMode: string;
        primaryKey: string;
        permitRootLogin: string;
        useDNS: string;
        currentUser: string;
    }
    export interface SSHUpdate {
        key: string;
        oldValue: string;
        newValue: string;
    }
    export interface RootCert {
        name: string;
        mode: string;
        encryptionMode: string;
        passPhrase: string;
        privateKey: string;
        publicKey: string;
        description: string;
    }
    export interface RootCertInfo {
        id: number;
        createAt: Date;
        name: string;
        mode: string;
        encryptionMode: string;
        passPhrase: string;
        description: string;
        publicKey: string;
        privateKey: string;
    }
    export interface searchSSHLog extends ReqPage {
        info: string;
        status: string;
    }
    export interface analysisSSHLog extends ReqPage {
        orderBy: string;
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

    export interface DiskBasicInfo {
        device: string;
        size: string;
        model: string;
        diskType: string;
        isRemovable: boolean;
        isSystem: boolean;
        filesystem: string;
        used: string;
        avail: string;
        usePercent: number;
        mountPoint: string;
        isMounted: boolean;
        serial: string;
    }

    export interface DiskInfo extends DiskBasicInfo {
        partitions?: DiskBasicInfo[];
    }

    export interface CompleteDiskInfo {
        disks: DiskInfo[];
        unpartitionedDisks: DiskBasicInfo[];
        systemDisk?: DiskInfo;
        totalDisks: number;
        totalCapacity: number;
    }

    export interface DiskPartition {
        device: string;
        filesystem: string;
        label: string;
        autoMount: boolean;
        mountPoint: string;
    }

    export interface DiskMount {
        device: string;
        mountPoint: string;
        filesystem?: string;
    }

    export interface DiskUmount {
        mountPoint: string;
    }

    export interface ComponentInfo {
        exists: boolean;
        version: string;
        path: string;
        error: string;
    }
}
