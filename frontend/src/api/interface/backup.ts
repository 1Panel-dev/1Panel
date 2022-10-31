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
    export interface RecordDownload {
        source: string;
        fileDir: string;
        fileName: string;
    }
    export interface RecordInfo {
        id: number;
        createdAt: Date;
        source: string;
        backupType: string;
        fileDir: string;
        fileName: string;
    }
    export interface ForBucket {
        type: string;
        credential: string;
        vars: string;
    }
}
