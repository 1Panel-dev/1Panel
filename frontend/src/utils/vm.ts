export const isVMResourceUnavailable = (status?: string) =>
    ['creating', 'failed', 'deleted'].includes(status?.trim().toLowerCase() || '');

export const isVMMemoryAllocated = (status: string) =>
    ['running', 'blocked', 'paused', 'shutdown', 'pmsuspended'].includes(status);
