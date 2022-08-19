// import { File } from '@/api/interface/file';
import files from '@/api/interface/files.json';

export const GetFilesList = () => {
    // return http.post<Login.ResLogin>(`/auth/login`, params);
    return files;
};
