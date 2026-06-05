import { CommonModel, ReqPage } from '.';
export namespace File {
    export interface File extends CommonModel {
        path: string;
        name: string;
        user: string;
        group: string;
        uid: number;
        gid: number;
        content: string;
        size: number;
        isDir: boolean;
        isSymlink: boolean;
        isHidden: boolean;
        linkPath: string;
        type: string;
        updateTime: string;
        modTime: string;
        mode: number;
        mimeType: string;
        dirSize: number;
        items: File[];
        extension: string;
        itemTotal: number;
        favoriteID: number;
        shareCode: string;
        remark?: string;
    }

    export interface ReqFile extends ReqPage {
        path: string;
        search?: string;
        expand: boolean;
        dir?: boolean;
        showHidden?: boolean;
        containSub?: boolean;
        sortBy?: string;
        sortOrder?: string;
        isDetail?: boolean;
    }

    export interface FileAISearchReq {
        path: string;
        query: string;
        responseLanguage?: string;
        containSub?: boolean;
        maxItems?: number;
        matchCase?: boolean;
        wholeWord?: boolean;
        useRegex?: boolean;
        extensions?: string[];
        minSize?: number;
        maxSize?: number;
        modifiedAfter?: string;
        modifiedBefore?: string;
        maxScanFiles?: number;
        maxFileBytes?: number;
        maxHitsPerFile?: number;
        maxTotalHits?: number;
        contentHitsPromptMaxBytes?: number;
        llmMaxOutputTokens?: number;
    }

    export interface FileAIContentHit {
        path: string;
        line: number;
        text: string;
    }

    export interface FileAISearchResult {
        mode?: 'ai' | 'grep';
        summary: string;
        hits: FileAIContentHit[];
        contentScannedFiles: number;
        contentHitsTruncated: boolean;
        truncated: boolean;
        preFiltered: boolean;
        itemCount: number;
        promptTokens: number;
        completionTokens: number;
        totalTokens: number;
        duration: string;
    }

    export interface ReqNodeFile extends ReqFile {
        node: string;
    }

    export interface PreviewContentReq {
        path: string;
        isDetail?: boolean;
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
        mode?: number;
        isLink?: boolean;
        isSymlink?: boolean;
        linkPath?: boolean;
        sub?: boolean;
        name?: string;
    }

    export interface FileDelete {
        path: string;
        isDir: boolean;
        forceDelete: boolean;
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
        secret: string;
        taskID?: string;
    }

    export interface FileCompressStopReq {
        taskID: string;
    }

    export interface FileDeCompress {
        path: string;
        dst: string;
        type: string;
        secret: string;
        taskID?: string;
    }

    export interface FileDeCompressStopReq {
        taskID: string;
    }

    export interface FileEdit {
        path: string;
        content: string;
    }

    export interface FileHistorySearchReq {
        page: number;
        pageSize: number;
        path: string;
        scope: 'current' | 'all';
        operation?: string;
    }

    export interface FileHistoryInfo {
        id: number;
        fileId?: string;
        path: string;
        currentPath?: string;
        previousId?: number;
        sourcePath?: string;
        targetPath?: string;
        fileName: string;
        extension: string;
        fileMode?: string;
        operation: string;
        deleted?: boolean;
        contentSize: number;
        contentSHA: string;
        storagePath?: string;
        content?: string;
        currentContent?: string;
        createdAt: string;
        updatedAt: string;
    }

    export interface FileHistoryDeleteReq {
        ids: number[];
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
        ignoreCertificate?: boolean;
        useProxy?: boolean;
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

    export interface DepthDirSizeRes {
        size: number;
        path: string;
    }

    export interface FilePath {
        path: string;
    }

    export interface ExistFileInfo {
        name: string;
        path: string;
        size: number;
        uploadSize: number;
        modTime: string;
        isDir: boolean;
    }

    export interface RecycleBin {
        sourcePath: string;
        name: string;
        isDir: boolean;
        size: number;
        deleteTime: string;
        rName: string;
        from: string;
    }

    export interface RecycleBinReduce {
        rName: string;
        from: string;
        name: string;
    }

    export interface FileReadByLine {
        id?: number;
        type: string;
        name?: string;
        page: number;
        pageSize: number;
        taskID?: string;
        taskType?: string;
        taskOperate?: string;
        resourceID?: number;
    }

    export interface Favorite extends CommonModel {
        path: string;
        isDir: boolean;
        isTxt: boolean;
        name: string;
    }

    export interface FileRole {
        paths: string[];
        mode: number;
        user: string;
        group: string;
        sub: boolean;
    }

    export interface FileRemarkUpdate {
        path: string;
        remark: string;
    }

    export interface FileRemarksRes {
        remarks: Record<string, string>;
    }

    export interface UserGroupResponse {
        users: UserInfo[];
        groups: string[];
    }
    export interface UserInfo {
        username: string;
        group: string;
    }

    export interface ConvertFile {
        type: string;
        path: string;
        extension: string;
        inputFile: string;
        outputFormat: string;
    }

    export interface ConvertFileRequest {
        files: ConvertFile[];
        outputPath: string;
        deleteSource: boolean;
        taskID: string;
    }

    export interface ConvertLogResponse {
        date: string;
        type: string;
        log: string;
        status: string;
        message: string;
    }

    export interface FileShareCreate {
        path: string;
        expireMinutes: number;
        password?: string;
    }

    export interface FileShareCheck {
        code: string;
        password?: string;
        operateNode: string;
    }

    export interface FileShareInfo {
        code: string;
        path: string;
        fileName: string;
        expiresAt: number;
        permanent: boolean;
        hasPassword: boolean;
        password?: string;
    }

    export interface FileSharePublicInfo {
        fileName: string;
        expiresAt: number;
        permanent: boolean;
        hasPassword: boolean;
    }
}
