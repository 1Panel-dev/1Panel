import { CommonModel, ReqPage } from '.';

export namespace User {
    export interface User extends CommonModel {
        name: string;
        email: string;
        password: string;
    }
    export interface UserCreate {
        username: string;
        email: string;
    }

    export interface ReqGetUserParams extends ReqPage {
        info?: string;
        email?: string;
    }
}
