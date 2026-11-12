import type { AI } from '@/api/interface/ai';

export const getGPUEnvironmentState = (info?: AI.Info, failed = false) => {
    const checks = (info?.environment?.checks || []).filter(
        (check) => check.item === 'driver' || check.vendor === 'nvidia',
    );
    const deviceWarnings = [...(info?.gpu || []), ...(info?.npu || []), ...(info?.xpu || [])].flatMap(
        (device) => device.warnings || [],
    );
    const vendorWarnings = Object.values(info?.vendorWarnings || {}).flatMap((warnings) => warnings || []);
    const monitorIssueCount =
        checks.filter(
            (check) =>
                check.item === 'driver' &&
                check.status !== 'passed' &&
                !(
                    info?.vendorWarnings?.[check.vendor]?.length ||
                    (check.vendor === 'intel'
                        ? info?.xpu
                        : check.vendor === 'ascend'
                          ? info?.npu
                          : info?.gpu?.filter((device) => device.type === check.vendor)
                    )?.some((device) => device.warnings?.length)
                ),
        ).length +
        (info?.warnings?.length || 0) +
        vendorWarnings.length +
        deviceWarnings.length;
    const issueCount =
        monitorIssueCount + checks.filter((check) => check.item !== 'driver' && check.status !== 'passed').length;
    const status = failed
        ? 'loadFailed'
        : checks.length > 0 && issueCount === 0
          ? 'passed'
          : issueCount > 0
            ? 'issues'
            : info?.environment?.hardwareStatus === 'not_detected'
              ? 'notDetected'
              : 'unknown';
    return { checks, issueCount, monitorIssueCount, status };
};
