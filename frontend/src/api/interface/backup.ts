import { ReqPage } from '.';

export namespace Backup {
    export interface SearchWithType extends ReqPage {
        type: string;
        name: string;
    }
    export interface BackupInfo {
        id: number;
        name: string;
        type: string;
        accessKey: string;
        bucket: string;
        credential: string;
        rememberAuth: boolean;
        backupPath: string;
        vars: string;
        varsJson: object;
        createdAt: Date;
    }
    export interface OneDriveInfo {
        client_id: string;
        client_secret: string;
        redirect_uri: string;
    }
    export interface BackupOperate {
        id: number;
        type: string;
        accessKey: string;
        bucket: string;
        credential: string;
        backupPath: string;
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
        accessKey: string;
        credential: string;
        vars: string;
    }
    export interface SearchBackupRecord extends ReqPage {
        type: string;
        name: string;
        detailName: string;
    }
    export interface SearchBackupRecordByCronjob extends ReqPage {
        cronjobID: number;
    }
    export interface Backup {
        type: string;
        name: string;
        detailName: string;
        secret: string;
    }
    export interface Recover {
        source: string;
        type: string;
        name: string;
        detailName: string;
        file: string;
        secret: string;
    }
}
