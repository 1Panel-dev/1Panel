import type { AI } from '@/api/interface/ai';

export const getGPUEnvironmentState = (info?: AI.Info, failed = false) => {
    const checks = (info?.environment?.checks || []).filter(
        (check) => check.item === 'driver' || check.vendor === 'nvidia',
    );
    const issueCount = checks.filter((check) => check.status !== 'passed').length + (info?.warnings?.length || 0);
    const status = failed
        ? 'loadFailed'
        : checks.length > 0 && issueCount === 0
          ? 'passed'
          : issueCount > 0
            ? 'issues'
            : info?.environment?.hardwareStatus === 'not_detected'
              ? 'notDetected'
              : 'unknown';
    return { checks, issueCount, status };
};
