import { ReqPage, CommonModel } from '.';

export namespace App {
    export interface App extends CommonModel {
        name: string;
        icon: string;
        key: string;
        tags: Tag[];
        shortDescZh: string;
        shortDescEn: string;
        description: string;
        author: string;
        source: string;
        type: string;
        status: string;
        limit: number;
        website: string;
        github: string;
        readme: string;
    }

    interface Locale {
        zh: string;
        en: string;
        'zh-Hant': string;
        ja: string;
        ms: string;
        'pt-br': string;
        ru: string;
        ko: string;
        tr: string;
    }

    export interface AppDTO extends App {
        versions: string[];
        installed: boolean;
        architectures: string;
    }

    export interface Tag {
        key: string;
        name: string;
        sort: number;
    }

    export interface AppResPage {
        total: number;
        items: App.AppDTO[];
    }

    export interface AppUpdateRes {
        version: string;
        canUpdate: boolean;
    }

    export interface AppDetail extends CommonModel {
        appId: number;
        icon: string;
        version: string;
        readme: string;
        params: AppParams;
        dockerCompose: string;
        image: string;
        hostMode?: boolean;
        memoryRequired: number;
        architectures: string;
        gpuSupport: boolean;
    }

    export interface AppReq extends ReqPage {
        name?: string;
        tags?: string[];
        type?: string;
        recommend?: boolean;
        resource?: string;
        showCurrentArch?: boolean;
    }

    export interface AppParams {
        formFields: FromField[];
    }

    export interface FromField {
        type: string;
        labelZh: string;
        labelEn: string;
        required: boolean;
        default: any;
        envKey: string;
        key?: string;
        values?: ServiceParam[];
        child?: FromFieldChild;
        params?: FromParam[];
        multiple?: boolean;
        allowCreate?: boolean;
        label: Locale;
        description: Locale;
    }

    export interface FromFieldChild extends FromField {
        services: App.AppService[];
    }

    export interface FromParam {
        type: string;
        key: string;
        value: string;
        envKey: string;
    }

    export interface ServiceParam {
        label: '';
        value: '';
        from?: '';
    }

    export interface AppInstall {
        appDetailId: number;
        params: any;
        taskID: string;
    }

    export interface AppInstallSearch extends ReqPage {
        name?: string;
        tags?: string[];
        update?: boolean;
        unused?: boolean;
        sync?: boolean;
    }
    export interface ChangePort {
        key: string;
        name: string;
        port: number;
    }

    export interface AppInstalled extends CommonModel {
        name: string;
        appID: number;
        appDetailId: string;
        env: string;
        status: string;
        description: string;
        message: string;
        icon: string;
        canUpdate: boolean;
        path: string;
        httpPort?: number;
        httpsPort?: number;
        favorite: boolean;
        app: App;
        webUI: string;
    }

    export interface AppInstalledInfo {
        id: number;
        name: string;
        version: string;
        status: string;
        message: string;
        httpPort: number;
        container: string;
        env: { [key: string]: string };
    }

    export interface AppInstallDto {
        id: number;
        name: string;
        appID: number;
        appDetailID: number;
        version: string;
        status: string;
        message: string;
        httpPort: number;
        httpsPort: number;
        path: string;
        canUpdate: boolean;
        icon: string;
        appName: string;
        ready: number;
        total: number;
        appKey: string;
        appType: string;
        appStatus: string;
    }

    export interface AppInstalledInfo {
        id: number;
        key: string;
        name: string;
    }

    export interface CheckInstalled {
        name: string;
        version: string;
        isExist: boolean;
        app: string;
        status: string;
        createdAt: string;
        lastBackupAt: string;
        appInstallId: number;
        containerName: string;
        installPath: string;
        httpPort: number;
        httpsPort: number;
    }

    export interface DatabaseConnInfo {
        status: string;
        username: string;
        password: string;
        privilege: boolean;
        containerName: string;
        serviceName: string;
        systemIP: string;
        port: number;
    }
    export interface AppInstallResource {
        type: string;
        name: string;
    }

    export interface AppInstalledOp {
        installId: number;
        operate: string;
        backupId?: number;
        detailId?: number;
        forceDelete?: boolean;
        deleteBackup?: boolean;
        deleteImage?: boolean;
        taskID?: string;
    }

    export interface AppInstalledSearch extends ReqPage {
        type: string;
        unused?: boolean;
        all?: boolean;
    }

    export interface AppService {
        label: string;
        value: string;
        config?: Object;
        from?: string;
        status: string;
    }

    export interface VersionDetail {
        version: string;
        detailId: number;
    }

    export interface InstallParams {
        labelZh: string;
        labelEn: string;
        value: any;
        edit: boolean;
        key: string;
        rule: string;
        type: string;
        values?: any;
        showValue?: string;
        required?: boolean;
        multiple?: boolean;
        label: Locale;
    }

    export interface AppConfig {
        params: InstallParams[];
        cpuQuota: number;
        memoryLimit: number;
        memoryUnit: string;
        containerName: string;
        allowPort: boolean;
        dockerCompose: string;
        hostMode?: boolean;
        type: string;
        webUI: string;
        specifyIP: string;
        restartPolicy: string;
    }

    export interface IgnoredApp {
        name: string;
        detailID: number;
        version: string;
        scope: string;
    }

    export interface AppUpdateVersionReq {
        appInstallID: number;
        updateVersion?: string;
    }

    export interface AppIgnoreReq {
        appID: number;
        appDetailID: number;
        scope: string;
    }

    export interface CancelAppIgnore {
        id: number;
    }

    export interface AppStoreSync {
        taskID: string;
    }

    export interface AppConfigUpdate {
        installID: number;
        webUI: string;
    }

    export interface AppStoreConfig {
        uninstallDeleteImage: string;
        uninstallDeleteBackup: string;
        upgradeBackup: string;
    }

    export interface AppStoreConfigUpdate {
        scope: string;
        status: string;
    }

    export interface CustomAppStoreConfig {
        status: string;
        imagePrefix: string;
    }
}
