export namespace VM {
    export interface Search {
        info?: string;
    }

    export interface TaskResp {
        taskID: string;
    }

    export interface Environment {
        arch: string;
        archSupported: boolean;
        ready: boolean;
        helperExist: boolean;
        helperMessage: string;
        kvmExist: boolean;
        kvmMessage: string;
        libvirtExist: boolean;
        libvirtActive: boolean;
        libvirtMessage: string;
        storageExist: boolean;
        storageMessage: string;
    }

    export interface HealthCheckItem {
        level: 'PASS' | 'INFO' | 'WARN' | 'FAIL';
        check: string;
        reason?: string;
        params?: Record<string, string>;
        detail?: string;
        fixCommand?: string;
    }

    export interface HealthCheckResult {
        pass: number;
        warn: number;
        fail: number;
        items: HealthCheckItem[];
    }

    export interface Disk {
        capacity: number;
        format: string;
        readOnly: boolean;
        shareable: boolean;
        resizeBlocked: string;
        device: string;
        type: string;
        source: string;
        target: string;
        bus: string;
    }

    export interface NetInterface {
        type: string;
        source: string;
        model: string;
        mac: string;
        ip: string;
    }

    export interface Graphics {
        type: string;
        port: string;
        listen: string;
        url?: string;
    }

    export interface Stats {
        name: string;
        status: string;
        timestamp: number;
        cpuTime: number;
        vcpu: number;
        memory: MemoryStats;
        disk: DiskStats;
        network: NetworkStats;
    }

    export interface MemoryStats {
        availableKiB: number;
        unusedKiB: number;
        usableKiB: number;
        rssKiB: number;
        actualKiB: number;
    }

    export interface DiskStats {
        readBytes: number;
        writeBytes: number;
        readReqs: number;
        writeReqs: number;
    }

    export interface NetworkStats {
        rxBytes: number;
        txBytes: number;
        rxPackets: number;
        txPackets: number;
    }

    export interface Item {
        id: number;
        name: string;
        uuid: string;
        cpu: number;
        /** Memory is measured in MiB. */
        memory: number;
        diskSize: number;
        diskBus: 'sata' | 'virtio' | string;
        diskPath: string;
        storagePool: string;
        isoPaths: string[];
        imagePath: string;
        network: string;
        bootOrder: string;
        autoStart: boolean;
        status: string;
        disks: Disk[];
        interfaces: NetInterface[];
        graphics?: Graphics;
        taskID: string;
        createTaskID: string;
        description: string;
    }

    export interface Create {
        taskID?: string;
        name: string;
        cpu: number;
        /** Memory is measured in MiB. */
        memory: number;
        diskSize: number;
        osType?: 'windows' | 'linux' | 'other';
        diskBus?: 'sata' | 'virtio';
        templateID?: number;
        diskPath?: string;
        storagePool?: string;
        isoPaths: string[];
        imagePath?: string;
        network: string;
        bootOrder?: string;
        startAfterCreate: boolean;
        autoStart: boolean;
        description?: string;
    }

    export interface TemplateItem {
        id: number;
        name: string;
        sourceVM: string;
        cpu: number;
        /** Memory is measured in MiB. */
        memory: number;
        diskSize: number;
        diskBus: 'sata' | 'virtio' | string;
        diskPath: string;
        storagePool: string;
        network: string;
        bootOrder: string;
        isExist: boolean;
        size: number;
        status: 'uploading' | 'downloading' | 'ready' | 'failed' | string;
        error: string;
        tempFileID: string;
        tempFileType: 'upload' | 'download' | string;
        createdAt: string;
        description: string;
    }

    export interface TemplateCreate {
        taskID?: string;
        name: string;
        templateName: string;
        storagePool?: string;
        description?: string;
    }

    export interface TemplateDelete {
        taskID?: string;
        id: number;
        deleteDisk: boolean;
    }

    export interface Update {
        taskID?: string;
        name: string;
        cpu: number;
        /** Memory is measured in MiB. */
        memory: number;
        isoPaths: string[];
        network: string;
        bootOrder?: string;
        autoStart: boolean;
        description?: string;
    }

    export interface DiskResize {
        taskID?: string;
        name: string;
        target: string;
        sizeGiB: number;
    }

    export interface Rename {
        taskID?: string;
        name: string;
        newName: string;
    }

    export interface Delete {
        taskID?: string;
        id: number;
        force: boolean;
    }

    export interface SnapshotItem {
        name: string;
        description: string;
        createdAt: number;
        state: string;
        parent: string;
        path: string;
        size: number;
        current: boolean;
        children: number;
    }

    export interface SnapshotCreate {
        taskID?: string;
        name: string;
        snapshotName: string;
        description?: string;
    }

    export interface SnapshotOperate {
        taskID?: string;
        name: string;
        snapshotName: string;
    }

    export interface IsoItem {
        id: number;
        type: string;
        name: string;
        path: string;
        isExist: boolean;
        size: number;
        status: string;
        error: string;
        tempFileID: string;
        tempFileType: string;
        createdAt: string;
        description: string;
    }

    export interface IsoCreate {
        type: string;
        name: string;
        path: string;
        description?: string;
    }

    export interface IsoUpdate {
        id: number;
        type: string;
        name: string;
        path: string;
        description?: string;
    }

    export interface IsoDelete {
        id: number;
        deleteFile?: boolean;
    }

    export interface IsoImportStart extends IsoCreate {
        status: 'uploading' | 'downloading';
        tempFileID?: string;
        tempFileType?: 'upload' | 'download';
    }

    export interface IsoImportUpdate {
        id: number;
        status: 'uploading' | 'downloading' | 'ready' | 'failed';
        error?: string;
        tempFileID?: string;
        tempFileType?: 'upload' | 'download';
    }

    export interface ResourceDelete {
        id: number;
        force?: boolean;
    }

    export interface NetworkItem {
        id: number;
        name: string;
        type: 'nat' | 'bridge' | 'macvtap' | string;
        bridgeName: string;
        sourceInterface: string;
        cidr: string;
        gateway: string;
        dhcpStart: string;
        dhcpEnd: string;
        autoStart: boolean;
        description: string;
        createdAt: string;
        status: string;
        error?: string;
        vmCount: number;
        vms: ResourceVM[];
        registered: boolean;
        usedIPs: NetworkIPUsage[];
    }

    export interface ResourceVM {
        id: number;
        name: string;
    }

    export interface NetworkIPUsage {
        ip: string;
        vmName: string;
        mac: string;
        source: string;
    }

    export interface NetworkInterfaceItem {
        name: string;
        addresses: string[];
    }

    export interface NetworkBridgeItem {
        name: string;
        recommended: boolean;
        description: string;
    }

    export interface NetworkCreate {
        name: string;
        type: 'nat' | 'bridge' | 'macvtap';
        bridgeName?: string;
        sourceInterface?: string;
        cidr?: string;
        gateway?: string;
        dhcpStart?: string;
        dhcpEnd?: string;
        autoStart: boolean;
        description?: string;
    }

    export interface NetworkUpdate {
        id: number;
        autoStart?: boolean;
        status?: string;
        description?: string;
    }

    export interface StorageItem {
        id: number;
        name: string;
        type: 'dir' | string;
        path: string;
        autoStart: boolean;
        description: string;
        createdAt: string;
        status: string;
        error?: string;
        capacity: number;
        available: number;
        usedPercent?: number;
        vmCount: number;
        vms: ResourceVM[];
    }

    export interface StorageCreate {
        name: string;
        type: 'dir';
        path: string;
        autoStart: boolean;
        description?: string;
    }

    export interface StorageUpdate {
        id: number;
        name?: string;
        path?: string;
        autoStart?: boolean;
        status?: string;
        description?: string;
    }

    export interface OrphanItem {
        type: 'orphan_domain_disk' | 'orphan_snapshot' | 'partial_file';
        path: string;
        size: number;
        modifiedAt: number;
    }
}
