export namespace Setting {
    export interface SettingInfo {
        userName: string;
        password: string;
        email: string;

        sessionTimeout: number;
        localTime: string;

        panelName: string;
        theme: string;
        language: string;

        serverPort: number;
        securityEntrance: string;
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
    export interface PasswordUpdate {
        oldPassword: string;
        newPassword: string;
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
