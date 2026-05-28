export const resolveRuntimeAppResource = (isOffline: boolean, customAppStatus?: string) => {
    return isOffline || customAppStatus?.toLowerCase() === 'enable' ? 'custom' : 'remote';
};
