import http from '@/api';
import { UpdateByFile } from '../interface';
import { Toolbox } from '../interface/toolbox';

// fail2ban
export const getFail2banBase = () => {
    return http.get<Toolbox.Fail2banBaseInfo>(`/toolbox/fail2ban/base`);
};
export const getFail2banConf = () => {
    return http.get<string>(`/toolbox/fail2ban/load/conf`);
};

export const searchFail2ban = (param: Toolbox.Fail2banSearch) => {
    return http.post<Array<string>>(`/toolbox/fail2ban/search`, param);
};

export const operateFail2ban = (operate: string) => {
    return http.post(`/toolbox/fail2ban/operate`, { operation: operate });
};

export const operatorFail2banSSHD = (param: Toolbox.Fail2banSet) => {
    return http.post(`/toolbox/fail2ban/operate/sshd`, param);
};

export const updateFail2ban = (param: Toolbox.Fail2banUpdate) => {
    return http.post(`/toolbox/fail2ban/update`, param);
};

export const updateFail2banByFile = (param: UpdateByFile) => {
    return http.post(`/toolbox/fail2ban/update/byconf`, param);
};
