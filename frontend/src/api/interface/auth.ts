export namespace Login {
    export interface ReqLoginForm {
        name: string;
        password: string;
        captcha: string;
        captchaID: string;
        authMethod: string;
    }
    export interface MFALoginForm {
        name: string;
        password: string;
        code: string;
        authMethod: string;
    }
    export interface ResLogin {
        name: string;
        token: string;
        mfaStatus: string;
    }
    export interface PasskeyBeginResponse {
        sessionId: string;
        publicKey: Record<string, any>;
    }
    export interface ResCaptcha {
        imagePath: string;
        captchaID: string;
        captchaLength: number;
    }
    export interface ResAuthButtons {
        [propName: string]: any;
    }

    export interface LoginSetting {
        isDemo: boolean;
        isIntl: boolean;
        isFxplay: boolean;
        language: string;
        menuTabs: string;
        panelName: string;
        theme: string;
        isOffLine: boolean;
        needCaptcha: boolean;
        passkeySetting: boolean;
    }
}
