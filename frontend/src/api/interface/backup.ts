import { ReqPage } from '.';

export namespace Backup {
    export interface SearchWithType extends ReqPage {
        type: string;
        name: string;
    }
    export interface BackupOption {
        id: number;
        name: string;
        type: string;
    }
    export interface BackupInfo {
        id: number;
        name: string;
        type: string;
        isPublic: boolean;
        accessKey: string;
        bucket: string;
        credential: string;
        rememberAuth: boolean;
        backupPath: string;
        bucketInput: boolean;
        vars: string;
        varsJson: object;
        createdAt: Date;
    }
    export interface ClientInfo {
        client_id: string;
        client_secret: string;
        redirect_uri: string;
    }
    export interface BackupOperate {
        id: number;
        type: string;
        name: string;
        isPublic: boolean;
        accessKey: string;
        bucket: string;
        credential: string;
        backupPath: string;
        vars: string;
    }
    export interface RecordDownload {
        downloadAccountID: number;
        fileDir: string;
        fileName: string;
    }
    export interface RecordInfo {
        id: number;
        createdAt: Date;
        accountType: string;
        accountName: string;
        downloadAccountID: number;
        fileDir: string;
        fileName: string;
        size: number;
    }
    export interface ForBucket {
        type: string;
        isPublic: boolean;
        accessKey: string;
        credential: string;
        vars: string;
    }
    export interface SearchBackupRecord extends ReqPage {
        type: string;
        name: string;
        detailName: string;
    }
    export interface SearchForSize extends ReqPage {
        type: string;
        name: string;
        detailName: string;
        info: string;
        cronjobID: number;
    }
    export interface RecordFileSize extends ReqPage {
        id: number;
        size: number;
    }
    export interface SearchBackupRecordByCronjob extends ReqPage {
        cronjobID: number;
    }
    export interface Backup {
        type: string;
        name: string;
        detailName: string;
        secret: string;
        taskID: string;
    }
    export interface Recover {
        downloadAccountID: number;
        type: string;
        name: string;
        detailName: string;
        file: string;
        secret: string;
        taskID: string;
    }
}
