import { File } from '@/api/interface/file';
import http from '@/api';
import { ResultData } from '@/api/interface';

export const GetFilesList = (params: File.ReqFile) => {
    return http.post<ResultData<File.File>>('files/search', params);
};
