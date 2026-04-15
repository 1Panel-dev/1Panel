type VueComponent = { template?: string } | Record<string, any>;

const EmptyComponent: VueComponent = { template: '<div></div>' };
const xpackViewModules = import.meta.glob('@/xpack/views/**/*.vue');

function findLoader(suffix: string) {
    for (const path in xpackViewModules) {
        if (path.endsWith(suffix)) {
            return xpackViewModules[path];
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
