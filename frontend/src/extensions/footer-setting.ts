import type { FooterNavigationSetting } from '@/components/footer-navigation/model';

type FooterSettingResult = {
    data: FooterNavigationSetting;
};

type EnterpriseFooterSettingModule = {
    getFooterSetting?: (silent?: boolean) => Promise<FooterSettingResult>;
};

type EnterpriseFooterSettingModuleLoader = () => Promise<EnterpriseFooterSettingModule>;

const enterpriseModules = import.meta.glob<EnterpriseFooterSettingModule>('@/enterprise/api/modules/footer-setting.ts');

function getEnterpriseFooterSettingModuleLoader(): EnterpriseFooterSettingModuleLoader | null {
    for (const path in enterpriseModules) {
        if (path.endsWith('/api/modules/footer-setting.ts')) {
            return enterpriseModules[path];
        }
    }
    return null;
}

export async function getEnterpriseFooterSetting(silent = false) {
    const loader = getEnterpriseFooterSettingModuleLoader();
    if (!loader) {
        return null;
    }
    const module = await loader();
    if (!module?.getFooterSetting) {
        return null;
    }
    return module.getFooterSetting(silent);
}
