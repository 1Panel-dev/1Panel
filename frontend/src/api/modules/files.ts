import { File } from '@/api/interface/file';
import http from '@/api';

export const GetFilesList = (params: File.ReqFile) => {
    return http.post<File.File>('files/search', params);
};
