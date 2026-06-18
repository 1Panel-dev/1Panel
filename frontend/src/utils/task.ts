import i18n from '@/lang';

const agentActionKeyMap: Record<string, string> = {
    Install: 'commons.button.install',
    Upgrade: 'commons.button.upgrade',
    Start: 'commons.operate.start',
    Stop: 'commons.operate.stop',
    Restart: 'commons.operate.restart',
    Delete: 'commons.operate.delete',
};

const getActionText = (action: string) => i18n.global.t(agentActionKeyMap[action]);

const formatBatchAgentTask = (action: string) => {
    return i18n.global.t('aiTools.agents.batchAgentTask', [getActionText(action)]);
};

const formatDispatchAgentTask = (action: string) => {
    return i18n.global.t('aiTools.agents.dispatchAgentTask', [getActionText(action)]);
};

const formatSkillInstallTask = () => {
    return `${'Skill'} ${i18n.global.t('commons.button.install')}`;
};

const taskTextMap: Record<string, () => string> = {
    BatchInstallAgent: () => formatBatchAgentTask('Install'),
    DispatchAgentInstallTasks: () => formatDispatchAgentTask('Install'),
    BatchUpgradeAgent: () => formatBatchAgentTask('Upgrade'),
    DispatchAgentUpgradeTasks: () => formatDispatchAgentTask('Upgrade'),
    BatchInstallAgentSkill: () =>
        i18n.global.t('aiTools.agents.batchAction', [i18n.global.t('aiTools.agents.installSkillAction')]),
    DispatchAgentSkillInstallTasks: () => i18n.global.t('aiTools.agents.dispatchTask', [formatSkillInstallTask()]),
    BatchStartAgent: () => formatBatchAgentTask('Start'),
    DispatchAgentStartTasks: () => formatDispatchAgentTask('Start'),
    BatchStopAgent: () => formatBatchAgentTask('Stop'),
    DispatchAgentStopTasks: () => formatDispatchAgentTask('Stop'),
    BatchRestartAgent: () => formatBatchAgentTask('Restart'),
    DispatchAgentRestartTasks: () => formatDispatchAgentTask('Restart'),
    BatchDeleteAgent: () => formatBatchAgentTask('Delete'),
    DispatchAgentDeleteTasks: () => formatDispatchAgentTask('Delete'),
};

export const translateTaskText = (value?: string) => {
    if (!value) {
        return value || '';
    }
    return Object.entries(taskTextMap)
        .sort(([prevKey], [nextKey]) => nextKey.length - prevKey.length)
        .reduce((text, [key, resolver]) => {
            return text.replaceAll(key, resolver());
        }, value);
};
