import http from '@/api';
import { File } from '../interface/file';

export const GetNginx = () => {
    return http.get<File.File>(`/nginx`);
};
