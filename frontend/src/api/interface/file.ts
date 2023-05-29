import { CommonModel, ReqPage } from '.';
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
        linkPath: boolean;
        type: string;
        updateTime: string;
        modTime: string;
        mode: number;
        mimeType: string;
        dirSize: number;
        items: File[];
        extension: string;
        itemTotal: number;
    }

    export interface ReqFile extends ReqPage {
        path: string;
        search?: string;
        expand: boolean;
        dir?: boolean;
        showHidden?: boolean;
        containSub?: boolean;
    }

    export interface SearchUploadInfo extends ReqPage {
        path: string;
    }
    export interface UploadInfo {
        name: string;
        size: number;
        createdAt: string;
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
        isLink?: boolean;
        isSymlink?: boolean;
        linkPath?: boolean;
        sub?: boolean;
        name?: string;
    }

    export interface FileDelete {
        path: string;
        isDir: boolean;
    }

    export interface FileBatchDelete {
        isDir: boolean;
        paths: Array<string>;
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

    export interface FileEdit {
        path: string;
        content: string;
    }

    export interface FileRename {
        oldName: string;
        newName: string;
    }

    export interface FileOwner {
        path: string;
        user: string;
        group: string;
        sub: boolean;
    }

    export interface FileWget {
        path: string;
        name: string;
        url: string;
    }

    export interface FileWgetRes {
        key: string;
    }

    export interface FileKeys {
        keys: string[];
    }

    export interface FileMove {
        oldPaths: string[];
        newPath: string;
        type: string;
    }

    export interface FileDownload {
        paths: string[];
        name: string;
        url: string;
    }

    export interface FileChunkDownload {
        name: string;
        path: string;
    }

    export interface DirSizeReq {
        path: string;
    }

    export interface DirSizeRes {
        size: number;
    }

    export interface FilePath {
        path: string;
    }
}
