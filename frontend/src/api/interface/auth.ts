export namespace Login {
    export interface ReqLoginForm {
        name: string;
        password: string;
        ignoreCaptcha: boolean;
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
    export interface ResCaptcha {
        imagePath: string;
        captchaID: string;
        captchaLength: number;
    }
    export interface ResAuthButtons {
        [propName: string]: any;
    }
}
