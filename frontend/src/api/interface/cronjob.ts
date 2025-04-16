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
        sourceAccountIDs: string;
        downloadAccountID: number;
        sourceAccountItems: Array<number>;

        retainCopies: number;
        retryTimes: number;
        timeout: number;
        timeoutItem: number;
        timeoutUint: string;
        status: string;
        secret: string;
        hasAlert: boolean;
        alertCount: number;
        alertTitle: string;
    }
    export interface Item {
        val: string;
    }
    export interface CronjobOperate {
        id: number;
        name: string;
        type: string;
        specCustom: boolean;
        spec: string;
        specs: Array<string>;
        specObjs: Array<SpecObj>;

        appID: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        isDir: boolean;
        sourceDir: string;

        //shell
        executor: string;
        scriptMode: string;
        script: string;
        command: string;
        containerName: string;
        user: string;

        sourceAccountIDs: string;
        downloadAccountID: number;
        retainCopies: number;
        retryTimes: number;
        timeout: number;
        secret: string;

        alertCount: number;
        alertTitle: string;
    }
    export interface SpecObj {
        specType: string;
        week: number;
        day: number;
        hour: number;
        minute: number;
        second: number;
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

    export interface ScriptInfo {
        id: number;
        name: string;
        script: string;
        groups: string;
        groupList: Array<number>;
        groupBelong: Array<string>;
        description: string;
        createdAt: Date;
    }
    export interface ScriptOperate {
        id: number;
        name: string;
        script: string;
        groups: string;
        description: string;
    }
}
