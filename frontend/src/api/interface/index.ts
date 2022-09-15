export interface Result {
    code: number;
    message: string;
}

export interface ResultData<T> {
    code: number;
    message: string;
    data: T;
}

export interface ResPage<T> {
    items: T[];
    total: number;
}

export interface ReqPage {
    page: number;
    pageSize: number;
}
export interface CommonModel {
    id: number;
    CreatedAt?: string;
    UpdatedAt?: string;
}

// * 登录模块
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
    export interface ResCaptcha {
        imagePath: string;
        captchaID: string;
        captchaLength: number;
    }
    export interface ResAuthButtons {
        [propName: string]: any;
    }
}
// * 文件上传模块
export namespace Upload {
    export interface ResFileUrl {
        fileUrl: string;
    }
}
