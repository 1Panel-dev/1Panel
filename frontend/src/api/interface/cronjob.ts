import { ReqPage } from '.';

export namespace Cronjob {
    export interface CronjobInfo {
        id: number;
        name: string;
        type: string;
        specCustom: boolean;
        spec: string;
        specs: Array<string>;
        specObjs: Array<SpecObj>;

        executor: string;
        isExecutorCustom: boolean;
        script: string;
        scriptMode: string;
        isCustom: boolean;
        command: string;
        inContainer: boolean;
        containerName: string;
        user: string;
        appID: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        isDir: boolean;
        files: Array<Item>;
        sourceDir: string;

        sourceAccounts: Array<string>;
        downloadAccount: string;
        retainCopies: number;
        status: string;
        secret: string;
    }
    export interface Item {
        val: string;
    }
    export interface CronjobCreate {
        name: string;
        type: string;
        specCustom: boolean;
        spec: string;
        specObjs: Array<SpecObj>;

        script: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        sourceDir: string;

        sourceAccountIDs: string;
        downloadAccountID: number;
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
        specCustom: boolean;
        spec: string;

        script: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        sourceDir: string;

        sourceAccountIDs: string;
        downloadAccountID: number;
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
        taskID: string;
        file: string;
        startTime: string;
        records: string;
        status: string;
        message: string;
        targetPath: string;
        interval: number;
    }
}
