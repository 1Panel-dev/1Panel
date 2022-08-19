import { CommonModel } from '.';
export namespace File {
    export interface File extends CommonModel {
        name: string;
        mode: number;
        user: string;
        group: string;
        updateDate: string;
        isDir: boolean;
        isLink: boolean;
        path: string;
        size: number;
        accessTime: string;
        changeTime: string;
    }
}
