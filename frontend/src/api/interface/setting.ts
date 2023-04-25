import { DateTimeFormats } from '@intlify/core-base';

export namespace Setting {
    export interface SettingInfo {
        userName: string;
        password: string;
        email: string;
        systemVersion: string;

        sessionTimeout: number;
        localTime: string;

        panelName: string;
        theme: string;
        language: string;

        serverPort: number;
        securityEntranceStatus: string;
        ssl: string;
        sslType: string;
        securityEntrance: string;
        expirationDays: number;
        expirationTime: string;
        complexityVerification: string;
        mfaStatus: string;
        mfaSecret: string;

        monitorStatus: string;
        monitorStoreDays: number;

        messageType: string;
        emailVars: string;
        weChatVars: string;
        dingVars: string;
    }
    export interface SettingUpdate {
        key: string;
        value: string;
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
    export interface MFAInfo {
        secret: string;
        qrImage: string;
    }
    export interface MFABind {
        secret: string;
        code: string;
    }
    export interface SnapshotCreate {
        from: string;
        description: string;
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
    }
    export interface SnapshotInfo {
        id: number;
        name: string;
        from: string;
        description: string;
        status: string;
        message: string;
        createdAt: DateTimeFormats;
        version: string;
        interruptStep: string;
        recoverStatus: string;
        recoverMessage: string;
        lastRecoveredAt: string;
        rollbackStatus: string;
        rollbackMessage: string;
        lastRollbackedAt: string;
    }
    export interface UpgradeInfo {
        newVersion: string;
        latestVersion: string;
        releaseNote: string;
    }
}
