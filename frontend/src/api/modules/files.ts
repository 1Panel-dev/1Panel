import { File } from '@/api/interface/file';
import http from '@/api';

export const GetFilesList = (params: File.ReqFile) => {
    return http.post<File.File>('files/search', params);
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

export const ChangeFileMode = (form: File.FileCreate) => {
    return http.post<File.File>('files/mode', form);
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

export const UploadFileData = (params: FormData) => {
    return http.post<File.File>('files/upload', params);
};

export const RenameRile = (params: File.FileRename) => {
    return http.post<File.File>('files/rename', params);
};

export const WgetFile = (params: File.FileWget) => {
    return http.post<File.File>('files/wget', params);
};

export const MoveFile = (params: File.FileMove) => {
    return http.post<File.File>('files/move', params);
};

export const DownloadFile = (params: File.FileDownload) => {
    return http.download<BlobPart>('files/download', params, { responseType: 'blob' });
};

export const ComputeDirSize = (params: File.DirSizeReq) => {
    return http.post<File.DirSizeRes>('files/size', params);
};
