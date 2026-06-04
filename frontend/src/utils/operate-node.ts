let operateNodeOverride = '';

export const setOperateNodeOverride = (node?: string) => {
    operateNodeOverride = node || '';
};

export const getOperateNodeOverride = () => {
    return operateNodeOverride;
};

export const clearOperateNodeOverride = (node?: string) => {
    if (!node || operateNodeOverride === node) {
        operateNodeOverride = '';
    }
};
