import { ReqPage } from '.';

export namespace Cronjob {
    export interface Search extends ReqPage {
        info: string;
        groupIDs: Array<number>;
        orderBy?: string;
        order?: string;
    }
    export interface CronjobInfo {
        id: number;
        name: string;
        type: string;
        groupID: number;
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
        scriptID: number;
        appID: string;
        website: string;
        exclusionRules: string;
        ignoreFiles: Array<string>;
        dbType: string;
        dbName: string;
        url: string;
        isDir: boolean;
        files: Array<Item>;
        sourceDir: string;
        snapshotRule: snapshotRule;
        ignoreAppIDs: Array<Number>;
        withImage: boolean;

        websiteList: Array<string>;
        appIdList: Array<string>;
        dbNameList: Array<string>;

        sourceAccounts: Array<string>;
        downloadAccount: string;
        sourceAccountIDs: string;
        downloadAccountID: number;
        sourceAccountItems: Array<number>;

        retainCopies: number;
        ignoreErr: boolean;
        retryTimes: number;
        timeout: number;
        timeoutItem: number;
        timeoutUnit: string;
        status: string;
        secret: string;
        hasAlert: boolean;
        alertCount: number;
        alertTitle: string;
        alertMethod: string;
        alertMethodItems: Array<string>;
    }
    export interface Item {
        val: string;
    }
    export interface CronjobOperate {
        id: number;
        name: string;
        groupID: number;
        type: string;
        specCustom: boolean;
        spec: string;
        specs: Array<string>;
        specObjs: Array<SpecObj>;

        scriptID: number;
        appID: string;
        website: string;
        exclusionRules: string;
        dbType: string;
        dbName: string;
        url: string;
        isDir: boolean;
        sourceDir: string;
        snapshotRule: snapshotRule;

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
        ignoreErr: boolean;
        secret: string;

        alertCount: number;
        alertTitle: string;
        alertMethod: string;
    }
    export interface CronjobTrans {
        name: string;
        type: string;
        specCustom: boolean;
        spec: string;
        group: string;

        executor: string;
        scriptMode: string;
        script: string;
        command: string;
        containerName: string;
        user: string;
        url: string;

        scriptName: string;
        apps: Array<TransHelper>;
        websites: Array<string>;
        dbType: string;
        dbNames: Array<TransHelper>;

        exclusionRules: string;

        isDir: boolean;
        sourceDir: string;

        retainCopies: number;
        retryTimes: number;
        timeout: number;
        ignoreErr: boolean;
        snapshotRule: string;
        secret: string;

        sourceAccounts: Array<string>;
        downloadAccount: string;

        alertCount: number;
    }
    export interface TransHelper {
        name: string;
        detailName: string;
    }
    export interface snapshotTransHelper {
        withImage: boolean;
        ignoreApps: Array<TransHelper>;
    }
    export interface snapshotRule {
        withImage: boolean;
        ignoreAppIDs: Array<Number>;
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
        cleanRemoteData: boolean;
    }
    export interface UpdateStatus {
        id: number;
        status: string;
    }
    export interface Download {
        recordID: number;
        backupAccountID: number;
    }
    export interface ScriptOptions {
        id: number;
        name: string;
        script: string;
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
        isInteractive: boolean;
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
