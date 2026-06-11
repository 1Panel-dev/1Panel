import http from '@/api';
import { deepCopy } from '@/utils/misc';
import { encodeBase64Fields } from '@/utils/base64';
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
export const updateNodeFavorite = (id: number, isFavorite: boolean) => {
    return http.post(`/core/xpack/nodes/favorite`, { id, isFavorite });
};
export const listAllSimpleNodes = () => {
    return http.get<Array<Setting.SimpleNodeItem>>(`/core/nodes/simple/all`);
};
export const getLicenseSmsInfo = () => {
    return http.get<Setting.SmsInfo>(`/core/licenses/sms/info`);
};
export const listAppNodes = () => {
    return http.get<Array<Setting.NodeAppItem>>(`/core/xpack/nodes/apps/update`, {}, { timeout: TimeoutEnum.T_60S });
};

export const uploadEnterpriseLicense = (params: FormData) => {
    return http.upload('/core/enterprise/licenses/upload', params);
};
export const getEnterpriseLicense = () => {
    return http.get<Setting.LicenseEE>(`/core/enterprise/licenses/info`);
};
export const getEnterpriseLicenseStatus = () => {
    return http.get<Setting.LicenseStatus>(`/core/enterprise/licenses/status`);
};

// agent
export const loadBaseDir = (node?: string) => {
    const query = node ? `?operateNode=${node}` : '';
    return http.get<string>(`/settings/basedir${query}`);
};
export const loadWebsiteDir = () => {
    return http.get<string>(`/settings/website/dir`);
};
export const loadDaemonJsonPath = () => {
    return http.get<string>(`/settings/daemonjson`, {});
};
export const updateAgentSetting = (param: Setting.SettingUpdate) => {
    return http.post(`/settings/update`, param);
};
export const getAgentSettingInfo = (currentNode?: string) => {
    return http.post<Setting.AgentSettingInfo>(
        `/settings/search`,
        {},
        undefined,
        currentNode ? { CurrentNode: currentNode } : undefined,
    );
};
export const getAgentTerminalAIInfo = () => {
    return http.post<Setting.TerminalAIInfo>(`/settings/terminal/ai/search`);
};
export const updateAgentTerminalAIInfo = (param: Setting.TerminalAIInfo) => {
    return http.post(`/settings/terminal/ai/update`, param);
};
export const getAgentFileManageAIInfo = () => {
    return http.post<Setting.FileManageAIInfo>(`/settings/files/ai/search`);
};
export const updateAgentFileManageAIInfo = (param: Setting.FileManageAIInfo) => {
    return http.post(`/settings/files/ai/update`, param);
};
export const getAgentFileHistoryInfo = () => {
    return http.post<Setting.FileHistoryInfo>(`/settings/file-history/search`);
};
export const updateAgentFileHistoryInfo = (param: Setting.FileHistoryInfo) => {
    return http.post(`/settings/file-history/update`, param);
};
export const updateCommonDescription = (param: Setting.CommonDescription) => {
    return http.post(`/settings/description/save`, param);
};

// core
export const getSettingInfo = () => {
    return http.post<Setting.SettingInfo>(`/core/settings/search`);
};
export const getSettingBaseInfo = () => {
    return http.post<Setting.SettingBaseInfo>(`/core/settings/search/base`);
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
export const defaultMenu = () => {
    return http.post(`/core/settings/menu/default`);
};
export const updateProxy = (params: Setting.ProxyUpdate) => {
    let request = deepCopy(params) as Setting.ProxyUpdate;
    encodeBase64Fields(request, ['proxyPasswd']);
    request.proxyType = request.proxyType === 'close' ? '' : request.proxyType;
    return http.post(`/core/settings/proxy/update`, request);
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
export const getAppStoreConfig = (node?: string) => {
    const params = node ? `?operateNode=${node}` : '';
    return http.get<App.AppStoreConfig>(`/core/settings/apps/store/config${params}`);
};
export const updateAppStoreConfig = (req: App.AppStoreConfigUpdate) => {
    return http.post(`/core/settings/apps/store/update`, req);
};

// snapshot
export const loadSnapshotInfo = () => {
    return http.get<Setting.SnapshotData>(`/settings/snapshot/load`, {}, { timeout: TimeoutEnum.T_60S });
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
export const listReleases = () => {
    return http.get<Array<Setting.ReleasesNotes>>(`/core/settings/upgrade/releases`);
};
export const upgrade = (version: string) => {
    return http.post(`/core/settings/upgrade`, { version: version });
};

// memo
export const getMemo = () => {
    return http.get<string>(`/core/settings/memo`);
};
export const updateMemo = (content: string) => {
    return http.post(`/core/settings/memo`, { content });
};
