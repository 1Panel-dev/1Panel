export namespace Backup {
    export interface BackupInfo {
        id: number;
        type: string;
        bucket: string;
        vars: string;
        varsJson: object;
    }
    export interface BackupOperate {
        id: number;
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
