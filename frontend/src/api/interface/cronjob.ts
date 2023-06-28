import { ReqPage } from '.';

export namespace Cronjob {
    export interface CronjobInfo {
        id: number;
        name: string;
        type: string;
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;
        second: number;

        script: string;
        inContainer: boolean;
        containerName: string;
        website: string;
        exclusionRules: string;
        dbName: string;
        url: string;
        sourceDir: string;
        keepLocal: boolean;
        targetDirID: number;
        targetDir: string;
        retainCopies: number;
        status: string;
    }
    export interface CronjobCreate {
        name: string;
        type: string;
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;
        second: number;

        script: string;
        website: string;
        exclusionRules: string;
        dbName: string;
        url: string;
        sourceDir: string;
        keepLocal: boolean;
        targetDirID: number;
        retainCopies: number;
    }
    export interface CronjobUpdate {
        id: number;
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;
        second: number;

        script: string;
        website: string;
        exclusionRules: string;
        dbName: string;
        url: string;
        sourceDir: string;
        keepLocal: boolean;
        targetDirID: number;
        retainCopies: number;
    }
    export interface UpdateStatus {
        id: number;
        status: string;
    }
    export interface Download {
        recordID: number;
        backupAccountID: number;
    }
    export interface SearchRecord extends ReqPage {
        cronjobID: number;
        startTime: Date;
        endTime: Date;
        status: string;
    }
    export interface Record {
        id: number;
        startTime: Date;
        endTime: Date;
        records: string;
        status: string;
        message: string;
        targetPath: string;
        interval: number;
    }
}
