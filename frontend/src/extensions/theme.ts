type XpackThemeModule = {
    setPrimaryColor?: (color: string) => any;
};

function findModule<T>(modules: Record<string, T>, suffix: string): T | null {
    for (const path in modules) {
        if (path.endsWith(suffix)) {
            return modules[path];
        }
    }
    return null;
}

export function loadXpackStyles() {
    const xpackModules = import.meta.glob('@/xpack/styles/index.scss');
    const xpackLoader = findModule(xpackModules, '/styles/index.scss');
    xpackLoader?.();

    const enterpriseModules = import.meta.glob('@/enterprise/styles/index.scss');
    const enterpriseLoader = findModule(enterpriseModules, '/styles/index.scss');
    enterpriseLoader?.();
}

export function setXpackPrimaryColor(color: string) {
    const xpackModules = import.meta.glob('@/xpack/utils/theme/tool.ts', { eager: true }) as Record<
        string,
        XpackThemeModule
    >;
    const xpackModule = findModule(xpackModules, '/utils/theme/tool.ts');
    xpackModule?.setPrimaryColor?.(color);

    const enterpriseModules = import.meta.glob('@/enterprise/utils/theme/tool.ts', { eager: true }) as Record<
        string,
        XpackThemeModule
    >;
    const enterpriseModule = findModule(enterpriseModules, '/utils/theme/tool.ts');
    enterpriseModule?.setPrimaryColor?.(color);
}

export const loadExtensionStyles = loadXpackStyles;
export const setExtensionPrimaryColor = setXpackPrimaryColor;
