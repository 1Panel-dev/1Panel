import { File } from '@/api/interface/file';
import http from '@/api';
import { AxiosRequestConfig } from 'axios';
import { ResPage } from '../interface';
import { TimeoutEnum } from '@/enums/http-enum';
import { ReqPage } from '@/api/interface';

export const GetFilesList = (params: File.ReqFile) => {
    return http.post<File.File>('files/search', params, TimeoutEnum.T_5M);
};

export const GetUploadList = (params: File.SearchUploadInfo) => {
    return http.post<ResPage<File.UploadInfo>>('files/upload/search', params);
};

export const GetFilesTree = (params: File.ReqFile) => {
    return http.post<File.FileTree[]>('files/tree', params);
};

export const CreateFile = (form: File.FileCreate) => {
    return http.post<File.File>('files', form);
};

export const DeleteFile = (form: File.FileDelete) => {
    return http.post<File.File>('files/del', form);
};

export const BatchDeleteFile = (form: File.FileBatchDelete) => {
    return http.post('files/batch/del', form);
};

export const ChangeFileMode = (form: File.FileCreate) => {
    return http.post<File.File>('files/mode', form);
};

export const CompressFile = (form: File.FileCompress) => {
    return http.post<File.File>('files/compress', form, TimeoutEnum.T_10M);
};

export const DeCompressFile = (form: File.FileDeCompress) => {
    return http.post<File.File>('files/decompress', form, TimeoutEnum.T_10M);
};

export const GetFileContent = (params: File.ReqFile) => {
    return http.post<File.File>('files/content', params);
};

export const SaveFileContent = (params: File.FileEdit) => {
    return http.post<File.File>('files/save', params);
};

export const CheckFile = (path: string) => {
    return http.post<boolean>('files/check', { path: path });
};

export const UploadFileData = (params: FormData, config: AxiosRequestConfig) => {
    return http.upload<File.File>('files/upload', params, config);
};

export const ChunkUploadFileData = (params: FormData, config: AxiosRequestConfig) => {
    return http.upload<File.File>('files/chunkupload', params, config);
};

export const RenameRile = (params: File.FileRename) => {
    return http.post<File.File>('files/rename', params);
};

export const ChangeOwner = (params: File.FileOwner) => {
    return http.post<File.File>('files/owner', params);
};

export const WgetFile = (params: File.FileWget) => {
    return http.post<File.FileWgetRes>('files/wget', params);
};

export const MoveFile = (params: File.FileMove) => {
    return http.post<File.File>('files/move', params, TimeoutEnum.T_5M);
};

export const DownloadFile = (params: File.FileDownload) => {
    return http.download<BlobPart>('files/download', params, { responseType: 'blob', timeout: TimeoutEnum.T_40S });
};

export const ComputeDirSize = (params: File.DirSizeReq) => {
    return http.post<File.DirSizeRes>('files/size', params, TimeoutEnum.T_5M);
};

export const FileKeys = () => {
    return http.get<File.FileKeys>('files/keys');
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

export const SearchFavorite = (params: ReqPage) => {
    return http.post<ResPage<File.Favorite>>('files/favorite/search', params);
};

export const AddFavorite = (path: string) => {
    return http.post<any>('files/favorite', { path: path });
};

export const ReadByLine = (req: File.FileReadByLine) => {
    return http.post<any>('files/read', req);
};

export const RemoveFavorite = (id: number) => {
    return http.post<any>('files/favorite/del', { id: id });
};

export const BatchChangeRole = (params: File.FileRole) => {
    return http.post<any>('files/batch/role', params);
};

export const GetRecycleStatus = () => {
    return http.get<string>('files/recycle/status');
};
