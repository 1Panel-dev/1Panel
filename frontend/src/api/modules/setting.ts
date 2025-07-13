import http from '@/api';
import { deepCopy } from '@/utils/util';
import { Base64 } from 'js-base64';
import { ResPage, SearchWithPage, DescriptionUpdate, ReqPage } from '../interface';
import { Setting } from '../interface/setting';
import { TimeoutEnum } from '@/enums/http-enum';
import { App } from '../interface/app';

// license
export const uploadLicense = (oldLicense: string, params: FormData) => {
    if (oldLicense === '') {
        return http.upload('/core/licenses/upload', params);
    }
    return http.upload('/core/licenses/update', params);
};
export const searchLicense = (params: ReqPage) => {
    return http.post<ResPage<Setting.License>>('/core/licenses/search', params);
};
export const deleteLicense = (params: { ids: number }) => {
    return http.post('/core/licenses/del', params);
};
export const getLicenseStatus = () => {
    return http.get<Setting.LicenseStatus>(`/core/licenses/status`);
};
export const getMasterLicenseStatus = () => {
    return http.get<Setting.LicenseStatus>(`/core/licenses/master/status`);
};
export const syncLicense = (id: number) => {
    return http.post(`/core/licenses/sync`, { id: id });
};
export const bindLicense = (params: Setting.LicenseBind) => {
    return http.post(`/core/licenses/bind`, params, TimeoutEnum.T_60S);
};
export const unbindLicense = (params: Setting.LicenseUnbind) => {
    return http.post(`/core/licenses/unbind`, params, TimeoutEnum.T_60S);
};
export const changeBind = (id: number, nodeIDs: Array<number>) => {
    return http.post(`/core/licenses/bind/free`, { licenseID: id, nodeIDs: nodeIDs }, TimeoutEnum.T_60S);
};
export const loadLicenseOptions = () => {
    return http.get<Array<Setting.LicenseOptions>>(`/core/licenses/options`);
};
export const listNodeOptions = (type: string) => {
    return http.post<Array<Setting.NodeItem>>(`/core/nodes/list`, { type: type });
};

export const listAllNodes = () => {
    return http.get<Array<Setting.NodeItem>>(`/core/nodes/all`);
};

export const getLicenseSmsInfo = () => {
    return http.get<Setting.LicenseStatus>(`/core/licenses/sms/info`);
};

// agent
export const loadBaseDir = () => {
    return http.get<string>(`/settings/basedir`);
};
export const loadDaemonJsonPath = () => {
    return http.get<string>(`/settings/daemonjson`, {});
};
export const updateAgentSetting = (param: Setting.SettingUpdate) => {
    return http.post(`/settings/update`, param);
};
export const getAgentSettingInfo = () => {
    return http.post<Setting.SettingInfo>(`/settings/search`);
};
export const getAgentSettingByKey = (key: string) => {
    return http.get<string>(`/settings/get/${key}`);
};

// core
export const getSettingInfo = () => {
    return http.post<Setting.SettingInfo>(`/core/settings/search`);
};
export const getTerminalInfo = () => {
    return http.post<Setting.TerminalInfo>(`/core/settings/terminal/search`);
};
export const UpdateTerminalInfo = (param: Setting.TerminalInfo) => {
    return http.post(`/core/settings/terminal/update`, param);
};
export const getSystemAvailable = () => {
    return http.get(`/core/settings/search/available`);
};
export const updateSetting = (param: Setting.SettingUpdate) => {
    return http.post(`/core/settings/update`, param);
};
export const updateMenu = (param: Setting.SettingUpdate) => {
    return http.post(`/core/settings/menu/update`, param);
};
export const updateProxy = (params: Setting.ProxyUpdate) => {
    let request = deepCopy(params) as Setting.ProxyUpdate;
    if (request.proxyPasswd) {
        request.proxyPasswd = Base64.encode(request.proxyPasswd);
    }
    request.proxyType = request.proxyType === 'close' ? '' : request.proxyType;
    return http.post(`/core/settings/proxy/update`, request);
};
export const updatePassword = (param: Setting.PasswordUpdate) => {
    return http.post(`/core/settings/password/update`, param);
};
export const loadInterfaceAddr = () => {
    return http.get(`/core/settings/interface`);
};
export const updateBindInfo = (ipv6: string, bindAddress: string) => {
    return http.post(`/core/settings/bind/update`, { ipv6: ipv6, bindAddress: bindAddress });
};
export const updatePort = (param: Setting.PortUpdate) => {
    return http.post(`/core/settings/port/update`, param);
};
export const updateSSL = (param: Setting.SSLUpdate) => {
    return http.post(`/core/settings/ssl/update`, param);
};
export const loadSSLInfo = () => {
    return http.get<Setting.SSLInfo>(`/core/settings/ssl/info`);
};
export const downloadSSL = () => {
    return http.download<any>(`/core/settings/ssl/download`);
};
export const handleExpired = (param: Setting.PasswordUpdate) => {
    return http.post(`/core/settings/expired/handle`, param);
};
export const loadMFA = (param: Setting.MFARequest) => {
    return http.post<Setting.MFAInfo>(`/core/settings/mfa`, param);
};
export const bindMFA = (param: Setting.MFABind) => {
    return http.post(`/core/settings/mfa/bind`, param);
};
export const getAppStoreConfig = (node?: string) => {
    const params = node ? `?operateNode=${node}` : '';
    return http.get<App.AppStoreConfig>(`/core/settings/apps/store/config${params}`);
};
export const updateAppStoreConfig = (req: App.AppStoreConfigUpdate) => {
    return http.post(`/core/settings/apps/store/update`, req);
};

// snapshot
export const loadSnapshotInfo = () => {
    return http.get<Setting.SnapshotData>(`/settings/snapshot/load`);
};
export const snapshotCreate = (param: Setting.SnapshotCreate) => {
    return http.post(`/settings/snapshot`, param);
};
export const snapshotRecreate = (id: number) => {
    return http.post(`/settings/snapshot/recreate`, { id: id });
};
export const snapshotImport = (param: Setting.SnapshotImport) => {
    return http.post(`/settings/snapshot/import`, param);
};
export const updateSnapshotDescription = (param: DescriptionUpdate) => {
    return http.post(`/settings/snapshot/description/update`, param);
};
export const snapshotDelete = (param: { ids: number[]; deleteWithFile: boolean }) => {
    return http.post(`/settings/snapshot/del`, param);
};
export const snapshotRecover = (param: Setting.SnapshotRecover) => {
    return http.post(`/settings/snapshot/recover`, param);
};
export const snapshotRollback = (param: Setting.SnapshotRecover) => {
    return http.post(`/settings/snapshot/rollback`, param);
};
export const searchSnapshotPage = (param: SearchWithPage) => {
    return http.post<ResPage<Setting.SnapshotInfo>>(`/settings/snapshot/search`, param);
};

// upgrade
export const loadUpgradeInfo = () => {
    return http.get<Setting.UpgradeInfo>(`/core/settings/upgrade`);
};
export const loadReleaseNotes = (version: string) => {
    return http.post<string>(`/core/settings/upgrade/notes`, { version: version });
};
export const upgrade = (version: string) => {
    return http.post(`/core/settings/upgrade`, { version: version });
};

// api config
export const generateApiKey = () => {
    return http.post<string>(`/core/settings/api/config/generate/key`);
};
export const updateApiConfig = (param: Setting.ApiConfig) => {
    return http.post(`/core/settings/api/config/update`, param);
};
