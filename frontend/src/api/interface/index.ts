// * 请求响应参数(不包含data)
export interface Result {
    code: string;
    msg: string;
}

// * 请求响应参数(包含data)
export interface ResultData<T = any> extends Result {
    data?: T;
}

// * 分页响应参数
export interface ResPage<T> {
    items: T[];
    total: number;
    code: number;
    msg?: '';
}

// * 分页请求参数
export interface ReqPage {
    currentPage: number;
    pageSize: number;
}
export interface CommonModel {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
}

// * 登录模块
export namespace Login {
    export interface ReqLoginForm {
        username: string;
        password: string;
    }
    export interface ResLogin {
        access_token: string;
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
