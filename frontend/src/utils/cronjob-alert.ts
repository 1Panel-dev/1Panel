export type CronjobAlertTriggerMode = 'failed' | 'success' | 'both';

export const cronjobAlertTypes = [
    'shell',
    'app',
    'website',
    'database',
    'directory',
    'log',
    'snapshot',
    'curl',
    'cutWebsiteLog',
    'clean',
    'ntp',
    'syncIpGroup',
    'cleanLog',
];

export const cronjobAlertModes = [
    { value: 'failed', label: 'xpack.alert.alertTriggerFailed' },
    { value: 'success', label: 'xpack.alert.alertTriggerSuccess' },
    { value: 'both', label: 'xpack.alert.alertTriggerBoth' },
] as const;

export const normalizeCronjobAlertMode = (value: unknown): CronjobAlertTriggerMode => {
    return value === 'success' || value === 'both' ? value : 'failed';
};

const parseAdvancedParams = (value?: string): Record<string, unknown> => {
    try {
        const params = JSON.parse(value || '{}');
        return params && typeof params === 'object' && !Array.isArray(params) ? params : {};
    } catch {
        return {};
    }
};

export const getCronjobAlertMode = (advancedParams?: string): CronjobAlertTriggerMode => {
    return normalizeCronjobAlertMode(parseAdvancedParams(advancedParams).alertTriggerMode);
};

export const setCronjobAlertMode = (advancedParams: string | undefined, mode: CronjobAlertTriggerMode): string => {
    return JSON.stringify({ ...parseAdvancedParams(advancedParams), alertTriggerMode: mode });
};

export const getCronjobAlertModeLabel = (mode: unknown): string => {
    return cronjobAlertModes.find((item) => item.value === normalizeCronjobAlertMode(mode))!.label;
};

export const getCronjobAlertTitle = (
    row: { type: string; subType?: string; taskName?: string; title: string },
    renderTitle: (type: string, taskName: string) => string,
): string => {
    const type = row.type === 'cronJob' ? row.subType : row.type;
    if (type && cronjobAlertTypes.includes(type) && typeof row.taskName === 'string' && row.taskName.length > 0) {
        return renderTitle(type, row.taskName);
    }
    return row.title;
};

export const getCronjobAlertResult = (params: unknown): 'success' | 'failed' => {
    if (typeof params === 'string') {
        try {
            params = JSON.parse(params);
        } catch {
            return 'failed';
        }
    }
    if (!Array.isArray(params)) return 'failed';
    const result = params.find((item) => item?.index === 'result');
    if (result) return result.value === 'success' ? 'success' : 'failed';
    const numberedResult = params.find((item) => item?.index === '2' && item?.key === 'result');
    return numberedResult?.value === '成功' ? 'success' : 'failed';
};
