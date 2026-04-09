type XpackSettingModule = {
    searchXSetting?: () => Promise<any>;
    updateXSettingByKey?: (key: string, value: string) => Promise<any>;
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
    const xpackModule = findModule(xpackModules, '/api/modules/setting.ts');
    if (xpackModule) {
        return xpackModule;
    }
    const xpackEEModules = import.meta.glob('@/xpack-ee/api/modules/setting.ts', { eager: true }) as Record<
        string,
        XpackSettingModule
    >;
    return findModule(xpackEEModules, '/api/modules/setting.ts');
}

export async function searchXpackSetting() {
    const module = getXpackSettingModule();
    if (!module?.searchXSetting) {
        return null;
    }
    return module.searchXSetting();
}

export async function updateXpackSettingByKey(key: string, value: string) {
    const module = getXpackSettingModule();
    if (!module?.updateXSettingByKey) {
        return null;
    }
    return module.updateXSettingByKey(key, value);
}

export const searchExtensionSetting = searchXpackSetting;
export const updateExtensionSettingByKey = updateXpackSettingByKey;
