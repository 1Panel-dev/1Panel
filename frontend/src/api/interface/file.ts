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
        mimeType: string;
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
        isDir: boolean;
        path: string;
        children?: FileTree[];
    }

    export interface FileCreate {
        path: string;
        isDir: boolean;
        mode: number;
    }

    export interface FileDelete {
        path: string;
        isDir: boolean;
    }

    export interface FileCompress {
        files: string[];
        type: string;
        dst: string;
        name: string;
        replace: boolean;
    }

    export interface FileDeCompress {
        path: string;
        dst: string;
        type: string;
    }
}
