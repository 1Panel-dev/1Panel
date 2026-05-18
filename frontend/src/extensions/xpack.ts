type XpackSettingModule = {
    searchXSetting?: () => Promise<any>;
    getXProxyDocker?: () => Promise<any>;
    updateXSettingByKey?: (key: string, value: string) => Promise<any>;
};

type EnterpriseUserModule = {
    getUserInfo?: (currentNode: string) => Promise<any>;
};

type EnterpriseSkillsHubModule = {
    listPublishedSkillHub?: (params: { agentType?: string; keyword?: string }) => Promise<any>;
};

function findModule<T>(modules: Record<string, T>, suffix: string): T | null {
    for (const path in modules) {
        if (path.endsWith(suffix)) {
            return modules[path];
        }
    }
    return null;
}

function getXpackSettingModule(): XpackSettingModule | null {
    const xpackModules = import.meta.glob('@/xpack/api/modules/setting.ts', { eager: true }) as Record<
        string,
        XpackSettingModule
    >;
    return findModule(xpackModules, '/api/modules/setting.ts');
}

export async function searchXpackSetting() {
    const module = getXpackSettingModule();
    if (!module?.searchXSetting) {
        return null;
    }
    return module.searchXSetting();
}

export async function getXpackProxyDocker() {
    const module = getXpackSettingModule();
    if (!module?.getXProxyDocker) {
        return null;
    }
    return module.getXProxyDocker();
}

export async function updateXpackSettingByKey(key: string, value: string) {
    const module = getXpackSettingModule();
    if (!module?.updateXSettingByKey) {
        return null;
    }
    return module.updateXSettingByKey(key, value);
}

function getEnterpriseUserModule(): EnterpriseUserModule | null {
    const enterpriseModules = import.meta.glob('@/enterprise/api/modules/user.ts', { eager: true }) as Record<
        string,
        EnterpriseUserModule
    >;
    return findModule(enterpriseModules, '/api/modules/user.ts');
}

export async function getEnterpriseUserInfo(currentNode: string) {
    const module = getEnterpriseUserModule();
    if (!module?.getUserInfo) {
        return null;
    }
    return module.getUserInfo(currentNode);
}

function getEnterpriseSkillsHubModule(): EnterpriseSkillsHubModule | null {
    const enterpriseModules = import.meta.glob('@/enterprise/api/modules/skills-hub.ts', { eager: true }) as Record<
        string,
        EnterpriseSkillsHubModule
    >;
    return findModule(enterpriseModules, '/api/modules/skills-hub.ts');
}

export async function listEnterprisePublishedSkillHub(params: { agentType?: string; keyword?: string }) {
    const module = getEnterpriseSkillsHubModule();
    if (!module?.listPublishedSkillHub) {
        return null;
    }
    return module.listPublishedSkillHub(params);
}

export const searchExtensionSetting = searchXpackSetting;
export const getExtensionProxyDocker = getXpackProxyDocker;
export const updateExtensionSettingByKey = updateXpackSettingByKey;
