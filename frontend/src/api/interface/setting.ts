import { DateTimeFormats } from '@intlify/core-base';

export namespace Setting {
    export interface AgentSettingInfo {
        dockerSockPath: string;
        systemVersion: string;
        systemIP: string;

        localTime: string;
        timeZone: string;
        ntpSite: string;

        defaultNetwork: string;
        defaultIO: string;
        lastCleanTime: string;
        lastCleanSize: string;
        lastCleanData: string;

        monitorStatus: string;
        monitorInterval: string;
        monitorStoreDays: string;

        appStoreVersion: string;
        appStoreLastModified: string;
        appStoreSyncStatus: string;

        fileRecycleBin: string;
        localSSHConnShow: string;
        firewallPortWhiteList: string;
    }
    export interface SettingInfo {
        systemVersion: string;
        upgradeBackupCopies: string;
        developerMode: string;

        sessionTimeout: number;
        expirationDays: number;

        panelName: string;
        edition: string;
        theme: string;
        menuTabs: string;
        menuAccordion: string;
        language: string;
        docSource: string;

        serverPort: number;
        ipv6: string;
        bindAddress: string;
        ssl: string;
        sslType: string;
        allowIPs: string;
        bindDomain: string;
        passkeyTrustedProxies: string;
        securityEntrance: string;
        dashboardMemoVisible: string;
        dashboardSimpleNodeVisible: string;
        complexityVerification: string;

        messageType: string;
        emailVars: string;
        weChatVars: string;
        dingVars: string;
        snapshotIgnore: string;
        hideMenu: string;
        noAuthSetting: string;

        proxyUrl: string;
        proxyType: string;
        proxyPort: string;
        proxyUser: string;
        proxyPasswd: string;
        proxyPasswdKeep: string;

        opsReportExportFormat: string;
        opsReportSchedule: string;
        opsReportSavePath: string;
        opsReportThreshold: string;
    }
    export interface SettingBaseInfo {
        systemVersion: string;
        developerMode: string;
        upgradeBackupCopies: string;

        port: string;
        ipv6: string;
        bindAddress: string;
        panelName: string;
        edition: string;
        theme: string;
        menuTabs: string;
        menuAccordion: string;
        language: string;
        hideMenu: string;
        docSource: string;

        serverPort: string;
        securityEntrance: string;
        complexityVerification: string;
        noAuthSetting: string;
        proxyType: string;

        scriptSync: string;
        dashboardMemoVisible: string;
        dashboardSimpleNodeVisible: string;
    }
    export interface TerminalInfo {
        lineHeight: string;
        letterSpacing: string;
        fontSize: string;
        fontFamily: string;
        backgroundColor: string;
        foregroundColor: string;
        cursorBlink: string;
        cursorStyle: string;
        scrollback: string;
        scrollSensitivity: string;
    }
    export interface TerminalAIInfo {
        aiStatus: string;
        aiAccountId: string;
        aiPrefix: string;
        aiRiskCommands: string;
        aiRiskCommandsDefault?: string;
    }

    export interface FileManageAIInfo {
        aiStatus: string;
        aiAccountId: string;
    }

    export interface FileHistoryInfo {
        enable: string;
        maxPerPath: number;
        diskQuotaMB: number;
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
        withDockerRestart: boolean;
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
    export interface PortUpdate {
        serverPort: number;
    }
    export interface PasskeyRegisterRequest {
        name: string;
    }
    export interface PasskeyBeginResponse {
        sessionId: string;
        publicKey: Record<string, any>;
    }
    export interface PasskeyInfo {
        id: string;
        name: string;
        createdAt: string;
        lastUsedAt: string;
    }
    export interface CommonDescription {
        id: string;
        type: string;
        detailType: string;
        isPinned: boolean;
        description: string;
    }

    export interface SnapshotCreate {
        id: number;
        sourceAccountIDs: string;
        downloadAccountID: string;
        description: string;
        secret: string;
        timeout: number;

        appData: Array<DataTree>;
        panelData: Array<DataTree>;
        backupData: Array<DataTree>;

        withMonitorData: boolean;
        withLoginLog: boolean;
        withOperationLog: boolean;
    }
    export interface SnapshotImport {
        backupAccountID: number;
        names: Array<string>;
        description: string;
    }
    export interface SnapshotRecover {
        id: number;
        taskID: string;
        isNew: boolean;
        reDownload: boolean;
        secret: string;
    }
    export interface SnapshotInfo {
        id: number;
        name: string;
        sourceAccounts: Array<string>;
        downloadAccount: string;
        description: string;
        status: string;
        message: string;
        createdAt: DateTimeFormats;
        version: string;
        secret: string;
        timeout: number;

        taskID: string;
        taskRecoverID: string;
        taskRollbackID: string;

        interruptStep: string;
        recoverStatus: string;
        recoverMessage: string;
        rollbackStatus: string;
        rollbackMessage: string;
    }
    export interface SnapshotData {
        appData: Array<DataTree>;
        panelData: Array<DataTree>;
        backupData: Array<DataTree>;

        withMonitorData: boolean;
        withLoginLog: boolean;
        withOperationLog: boolean;
    }
    export interface DataTree {
        id: string;
        label: string;
        key: string;
        name: string;
        size: number;
        isShow: boolean;
        isDisable: boolean;

        path: string;

        Children: Array<DataTree>;
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
        status: string;
        message: string;
        smsUsed: number;
        smsTotal: number;
    }
    export interface LicenseOptions {
        id: number;
        licenseName: string;
        totalFreeCount: number;
        availableXpackCount: number;
        availableFreeCount: number;
    }
    export interface LicenseStatus {
        productPro: string;
        status: string;
        smsTotal: number;
        smsUsed: number;
    }
    export interface LicenseEE {
        deviceID: string;
        corporation: string;
        isv: string;
        expired: string;
        product: string;
        edition: string;
        licenseVersion: string;
        count: number;
        serialNo: string;
        remark: string;
        ext: string;

        status: string;
        message: string;
    }
    export interface NodeItem {
        id: number;
        groupID?: number;
        groupBelong?: string;
        addr: string;
        status: string;
        version: string;
        isXpack: boolean;
        isBound: boolean;
        isFavorite?: boolean;
        name: string;
    }
    export interface SimpleNodeItem {
        id: number;
        name: string;
        addr: string;
        description: string;
        systemVersion: string;
        securityEntrance: string;
        cpuUsedPercent: number;
        cpuTotal: number;
        memoryTotal: number;
        memoryUsedPercent: number;
    }
    export interface ReleasesNotes {
        Version: string;
        CreatedAt: string;
        Content: string;
        NewCount: number;
        OptimizationCount: number;
        FixCount: number;
    }

    export interface LicenseBind {
        nodeID: number;
        licenseID: number;
        syncList: string;
        withDockerRestart: boolean;
    }
    export interface LicenseUnbind {
        id: number;
        force: boolean;
        withDockerRestart: boolean;
    }

    export interface SmsInfo {
        licenseName: string;
        smsUsed: number;
        smsTotal: number;
    }

    export interface NodeAppItem {
        name: string;
        updateCount: number;
    }
}
