import { File } from '@/api/interface/file';
import http from '@/api';
import { AxiosRequestConfig } from 'axios';
import { ResPage } from '../interface';

export const GetFilesList = (params: File.ReqFile) => {
    return http.post<File.File>('files/search', params, 200000);
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

export const LoadFile = (form: File.FilePath) => {
    return http.post<string>('files/loadfile', form);
};

export const CompressFile = (form: File.FileCompress) => {
    return http.post<File.File>('files/compress', form);
};

export const DeCompressFile = (form: File.FileDeCompress) => {
    return http.post<File.File>('files/decompress', form);
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
    return http.post<File.File>('files/move', params);
};

export const DownloadFile = (params: File.FileDownload) => {
    return http.download<BlobPart>('files/download', params, { responseType: 'blob', timeout: 20000 });
};

export const DownloadByPath = (path: string) => {
    return http.download<BlobPart>('files/download/bypath', { path: path }, { responseType: 'blob', timeout: 40000 });
};

export const ComputeDirSize = (params: File.DirSizeReq) => {
    return http.post<File.DirSizeRes>('files/size', params);
};

export const FileKeys = () => {
    return http.get<File.FileKeys>('files/keys');
};
