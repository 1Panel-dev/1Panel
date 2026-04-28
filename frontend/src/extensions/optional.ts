type VueComponent = { template?: string } | Record<string, any>;

const EmptyComponent: VueComponent = { template: '<div></div>' };
const extensionViewModules = {
    ...import.meta.glob('@/xpack/views/**/*.vue'),
    ...import.meta.glob('@/enterprise/views/**/*.vue'),
};

function findLoader(suffix: string) {
    for (const path in extensionViewModules) {
        if (path.endsWith(suffix)) {
            return extensionViewModules[path];
        }
    }
    return null;
}

export async function loadOptionalComponent(key: string): Promise<VueComponent> {
    const loader = findLoader(key);
    if (!loader) {
        return EmptyComponent;
    }
    return ((await loader()) as { default?: VueComponent }).default || EmptyComponent;
}

export const loadExtensionComponent = loadOptionalComponent;
