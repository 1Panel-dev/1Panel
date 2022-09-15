export namespace Setting {
    export interface SettingInfo {
        userName: string;
        password: string;
        email: string;

        sessionTimeout: string;

        panelName: string;
        theme: string;
        language: string;

        serverPort: string;
        securityEntrance: string;
        complexityVerification: string;
        mfaStatus: string;

        monitorStatus: string;
        monitorStoreDays: string;

        messageType: string;
        emailVars: string;
        weChatVars: string;
        dingVars: string;
    }
}
