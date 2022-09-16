export namespace Backup {
    export interface BackupInfo {
        id: number;
        name: string;
        type: string;
        bucket: string;
        vars: string;
        varsJson: object;
    }
    export interface BackupOperate {
        id: number;
        name: string;
        type: string;
        bucket: string;
        credential: string;
        vars: string;
    }
    export interface ForBucket {
        type: string;
        credential: string;
        vars: string;
    }
}
