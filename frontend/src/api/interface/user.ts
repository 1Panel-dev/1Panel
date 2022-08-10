import { CommonModel, ReqPage } from '.';

export namespace User {
    export interface User extends CommonModel {
        username: string;
        email: string;
    }
    export interface UserCreate {
        username: string;
        email: string;
    }

    export interface ReqGetUserParams extends ReqPage {
        username?: string;
        email?: string;
    }
}
