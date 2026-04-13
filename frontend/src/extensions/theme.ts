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
    const modules = import.meta.glob('@/xpack/styles/index.scss');
    const loader = findModule(modules, '/styles/index.scss');
    loader?.();
}

export function setXpackPrimaryColor(color: string) {
    const modules = import.meta.glob('@/xpack/utils/theme/tool.ts', { eager: true }) as Record<
        string,
        XpackThemeModule
    >;
    const module = findModule(modules, '/utils/theme/tool.ts');
    return module?.setPrimaryColor?.(color);
}
