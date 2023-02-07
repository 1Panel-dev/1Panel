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
export interface SearchWithPage {
    info: string;
    page: number;
    pageSize: number;
}
export interface CommonModel {
    id: number;
    CreatedAt?: string;
    UpdatedAt?: string;
}

// * 文件上传模块
export namespace Upload {
    export interface ResFileUrl {
        fileUrl: string;
    }
}
