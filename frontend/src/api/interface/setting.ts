import { DateTimeFormats } from '@intlify/core-base';

export namespace Setting {
    export interface SettingInfo {
        userName: string;
        password: string;
        email: string;
        systemIP: string;
        systemVersion: string;
        dockerSockPath: string;
        developerMode: string;

        sessionTimeout: number;
        localTime: string;
        timeZone: string;
        ntpSite: string;

        panelName: string;
        theme: string;
        menuTabs: string;
        language: string;
        defaultNetwork: string;
        lastCleanTime: string;
        lastCleanSize: string;
        lastCleanData: string;

        serverPort: number;
        ipv6: string;
        bindAddress: string;
        ssl: string;
        sslType: string;
        allowIPs: string;
        bindDomain: string;
        securityEntrance: string;
        expirationDays: number;
        expirationTime: string;
        complexityVerification: string;
        mfaStatus: string;
        mfaSecret: string;
        mfaInterval: string;

        monitorStatus: string;
        monitorInterval: number;
        monitorStoreDays: number;

        messageType: string;
        emailVars: string;
        weChatVars: string;
        dingVars: string;
        snapshotIgnore: string;
        xpackHideMenu: string;
        noAuthSetting: string;

        proxyUrl: string;
        proxyType: string;
        proxyPort: string;
        proxyUser: string;
        proxyPasswd: string;
        proxyPasswdKeep: string;

        apiInterfaceStatus: string;
        apiKey: string;
        ipWhiteList: string;
        apiKeyValidityTime: number;
    }
    export interface SettingUpdate {
        key: string;
        value: string;
    }
    export interface ProxyUpdate {
        proxyUrl: string;
        proxyType: string;
        proxyPort: string;
        proxyUser: string;
        proxyPasswd: string;
        proxyPasswdKeep: string;
    }
    export interface SSLUpdate {
        ssl: string;
        domain: string;
        sslType: string;
        cert: string;
        key: string;
        sslID: number;
    }
    export interface SSLInfo {
        domain: string;
        timeout: string;
        rootPath: string;
        cert: string;
        key: string;
        sslID: number;
    }
    export interface PasswordUpdate {
        oldPassword: string;
        newPassword: string;
    }
    export interface PortUpdate {
        serverPort: number;
    }
    export interface MFARequest {
        title: string;
        interval: number;
    }
    export interface MFAInfo {
        secret: string;
        qrImage: string;
    }
    export interface MFABind {
        secret: string;
        code: string;
        interval: string;
    }

    export interface SnapshotCreate {
        id: number;
        from: string;
        fromAccounts: Array<string>;
        defaultDownload: string;
        description: string;
        secret: string;
    }
    export interface SnapshotImport {
        from: string;
        names: Array<string>;
        description: string;
    }
    export interface SnapshotRecover {
        id: number;
        isNew: boolean;
        reDownload: boolean;
        secret: string;
    }
    export interface SnapshotInfo {
        id: number;
        name: string;
        from: string;
        defaultDownload: string;
        description: string;
        status: string;
        message: string;
        created_at: DateTimeFormats;
        version: string;
        interruptStep: string;
        recoverStatus: string;
        recoverMessage: string;
        lastRecoveredAt: string;
        rollbackStatus: string;
        rollbackMessage: string;
        lastRollbackedAt: string;
        secret: string;
    }
    export interface SnapshotFile {
        id: number;
        name: string;
        size: number;
    }
    export interface SnapshotStatus {
        panel: string;
        panelInfo: string;
        daemonJson: string;
        appData: string;
        panelData: string;
        backupData: string;

        compress: string;
        size: string;
        upload: string;
    }
    export interface UpgradeInfo {
        testVersion: string;
        newVersion: string;
        latestVersion: string;
        releaseNote: string;
    }

    export interface License {
        licenseName: string;
        assigneeName: string;
        productPro: string;
        versionConstraint: string;
        trial: boolean;
        offline: boolean;
        status: string;
        message: string;
        smsUsed: number;
        smsTotal: number;
    }
    export interface LicenseStatus {
        productPro: string;
        trial: boolean;
        status: string;
    }
    export interface ApiConfig {
        apiInterfaceStatus: string;
        apiKey: string;
        ipWhiteList: string;
        apiKeyValidityTime: number;
    }
}
