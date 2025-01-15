export function setPrimaryColor(color: string) {
    let setPrimaryColor: (arg0: string) => any;
    const xpackModules = import.meta.glob('../xpack/utils/theme/tool.ts', { eager: true });
    if (xpackModules['../xpack/utils/theme/tool.ts']) {
        setPrimaryColor = xpackModules['../xpack/utils/theme/tool.ts']['setPrimaryColor'] || {};
        return setPrimaryColor(color);
    }
}
