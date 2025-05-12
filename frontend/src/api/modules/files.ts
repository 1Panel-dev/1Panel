import { File } from '@/api/interface/file';
import http from '@/api';
import { AxiosRequestConfig } from 'axios';
import { ResPage } from '../interface';
import { TimeoutEnum } from '@/enums/http-enum';
import { ReqPage } from '@/api/interface';
import { Dashboard } from '@/api/interface/dashboard';

export const getFilesList = (params: File.ReqFile) => {
    return http.post<File.File>('files/search', params, TimeoutEnum.T_5M);
};

export const getFilesListByNode = (params: File.ReqNodeFile) => {
    return http.post<File.File>('files/search?operateNode=' + params.node, params, TimeoutEnum.T_5M);
};

export const getUploadList = (params: File.SearchUploadInfo) => {
    return http.post<ResPage<File.UploadInfo>>('files/upload/search', params);
};

export const getFilesTree = (params: File.ReqFile) => {
    return http.post<File.FileTree[]>('files/tree', params);
};

export const createFile = (form: File.FileCreate) => {
    return http.post<File.File>('files', form);
};

export const deleteFile = (form: File.FileDelete) => {
    return http.post<File.File>('files/del', form);
};

export const deleteFileByNode = (form: File.FileDelete, node: string) => {
    return http.post<File.File>('files/del?operateNode=' + node, form);
};

export const batchDeleteFile = (form: File.FileBatchDelete) => {
    return http.post('files/batch/del', form);
};

export const changeFileMode = (form: File.FileCreate) => {
    return http.post<File.File>('files/mode', form);
};

export const compressFile = (form: File.FileCompress) => {
    return http.post<File.File>('files/compress', form, TimeoutEnum.T_10M);
};

export const deCompressFile = (form: File.FileDeCompress) => {
    return http.post<File.File>('files/decompress', form, TimeoutEnum.T_10M);
};

export const getFileContent = (params: File.ReqFile) => {
    return http.post<File.File>('files/content', params);
};

export const saveFileContent = (params: File.FileEdit) => {
    return http.post<File.File>('files/save', params);
};

export const checkFile = (path: string, withInit: boolean) => {
    return http.post<boolean>('files/check', { path: path, withInit: withInit });
};

export const uploadFileData = (params: FormData, config: AxiosRequestConfig) => {
    return http.upload<File.File>('files/upload', params, config);
};

export const batchCheckFiles = (paths: string[]) => {
    return http.post<File.ExistFileInfo[]>('files/batch/check', { paths: paths }, TimeoutEnum.T_5M);
};

export const chunkUploadFileData = (params: FormData, config: AxiosRequestConfig) => {
    return http.upload<File.File>('files/chunkupload', params, config);
};

export const renameRile = (params: File.FileRename) => {
    return http.post<File.File>('files/rename', params);
};

export const changeOwner = (params: File.FileOwner) => {
    return http.post<File.File>('files/owner', params);
};

export const wgetFile = (params: File.FileWget) => {
    return http.post<File.FileWgetRes>('files/wget', params);
};

export const moveFile = (params: File.FileMove) => {
    return http.post<File.File>('files/move', params, TimeoutEnum.T_5M);
};

export const downloadFile = (params: File.FileDownload) => {
    return http.download<BlobPart>('files/download', params, { responseType: 'blob', timeout: TimeoutEnum.T_40S });
};

export const computeDirSize = (params: File.DirSizeReq) => {
    return http.post<File.DirSizeRes>('files/size', params, TimeoutEnum.T_5M);
};

export const fileWgetKeys = () => {
    return http.get<File.FileKeys>('files/wget/process/keys');
};

export const getRecycleList = (params: ReqPage) => {
    return http.post<ResPage<File.RecycleBin>>('files/recycle/search', params);
};

export const reduceFile = (params: File.RecycleBinReduce) => {
    return http.post<any>('files/recycle/reduce', params);
};

export const clearRecycle = () => {
    return http.post<any>('files/recycle/clear');
};

export const searchFavorite = (params: ReqPage) => {
    return http.post<ResPage<File.Favorite>>('files/favorite/search', params);
};

export const addFavorite = (path: string) => {
    return http.post<any>('files/favorite', { path: path });
};

export const readByLine = (req: File.FileReadByLine) => {
    return http.post<any>('files/read', req);
};

export const removeFavorite = (id: number) => {
    return http.post<any>('files/favorite/del', { id: id });
};

export const batchChangeRole = (params: File.FileRole) => {
    return http.post<any>('files/batch/role', params);
};

export const getRecycleStatus = () => {
    return http.get<string>('files/recycle/status');
};

export const getRecycleStatusByNode = (node: string) => {
    return http.get<string>('files/recycle/status?operateNode=' + node);
};

export const getPathByType = (pathType: string) => {
    return http.get<string>(`files/path/${pathType}`);
};

export const searchHostMount = () => {
    return http.post<Dashboard.DiskInfo[]>(`/files/mount`);
};

export const searchUserGroup = () => {
    return http.post<File.UserGroupResponse>(`/files/user/group`);
};
