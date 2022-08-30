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
    return http.post<File.FileCreate>('files/mode', form);
};

export const CompressFile = (form: File.FileCompress) => {
    return http.post<File.FileCompress>('files/compress', form);
};
