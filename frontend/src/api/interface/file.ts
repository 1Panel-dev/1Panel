import { CommonModel } from '.';
export namespace File {
    export interface File extends CommonModel {
        path: string;
        name: string;
        user: string;
        group: string;
        content: string;
        size: number;
        isDir: boolean;
        isSymlink: boolean;
        type: string;
        updateTime: string;
        modTime: string;
        mode: number;
        items: File[];
    }

    export interface ReqFile {
        path: string;
        search?: string;
        expand: boolean;
    }

    export interface FileTree {
        id: string;
        name: string;
        isDir: Boolean;
        path: string;
        children?: FileTree[];
    }
}
