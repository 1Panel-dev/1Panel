export interface Result {
    code: number;
    errorCode?: string;
    message: string;
}

export interface ResultData<T> {
    code: number;
    errorCode?: string;
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
    excludeAppStore?: boolean;
    orderBy?: string;
    order?: string;
    name?: string;
    type?: string;
}
export interface CommonModel {
    id: number;
    createdAt?: string;
    updatedAt?: string;
}
export interface DescriptionUpdate {
    id: number;
    description: string;
}
export interface UpdateByFile {
    file: string;
}
