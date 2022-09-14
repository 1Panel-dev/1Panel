export namespace Setting {
    export interface SettingInfo {
        userName: string;
        password: string;
        email: string;

        sessionTimeout: string;
        localTime: string;

        panelName: string;
        theme: string;
        language: string;

        serverPort: string;
        securityEntrance: string;
        passwordTimeOut: string;
        complexityVerification: string;
        mfaStatus: string;
        mfaSecret: string;

        monitorStatus: string;
        monitorStoreDays: string;

        messageType: string;
        emailVars: string;
        weChatVars: string;
        dingVars: string;
    }
    export interface SettingUpdate {
        key: string;
        value: string;
    }
    export interface PasswordUpdate {
        oldPassword: string;
        newPassword: string;
        retryPassword: string;
    }
    export interface MFAInfo {
        secret: string;
        qrImage: string;
    }
    export interface MFABind {
        secret: string;
        code: string;
    }
}
