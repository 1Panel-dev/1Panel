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
        secret: string;
        code: string;
        authMethod: string;
    }
    export interface ResLogin {
        name: string;
        token: string;
        mfaStatus: string;
        mfaSecret: string;
    }
    export interface InitUser {
        name: string;
        password: string;
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
