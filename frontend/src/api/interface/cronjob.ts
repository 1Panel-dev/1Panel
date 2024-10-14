import { ReqPage } from '.';

export namespace Cronjob {
    export interface CronjobInfo {
        id: number;
        name: string;
        type: string;
        spec: string;
        specObjs: Array<SpecObj>;

        script: string;
        isCustom: boolean;
        command: string;
        inContainer: boolean;
        containerName: string;
        appID: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        sourceDir: string;

        backupAccounts: string;
        defaultDownload: string;
        backupAccountList: Array<string>;
        appIdList: Array<string>;
        websiteList: Array<string>;
        dbNameList: Array<string>;
        retainCopies: number;
        status: string;
        secret: string;
        hasAlert: boolean;
        alertCount: number;
        alertTitle: string;
    }
    export interface CronjobCreate {
        name: string;
        type: string;
        spec: string;
        specObjs: Array<SpecObj>;

        script: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        sourceDir: string;

        backupAccounts: string;
        defaultDownload: string;
        retainCopies: number;
        secret: string;
    }
    export interface SpecObj {
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;
        second: number;
    }
    export interface CronjobUpdate {
        id: number;
        spec: string;

        script: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        sourceDir: string;

        backupAccounts: string;
        defaultDownload: string;
        retainCopies: number;
        secret: string;
    }
    export interface CronjobDelete {
        ids: Array<number>;
        cleanData: boolean;
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
        file: string;
        startTime: string;
        records: string;
        status: string;
        message: string;
        targetPath: string;
        interval: number;
    }
}
