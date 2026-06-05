import i18n from '@/lang';

const taskTextMap: Record<string, string> = {
    BatchInstallAgent: 'aiTools.agents.batchInstallAgent',
    DispatchAgentInstallTasks: 'aiTools.agents.dispatchAgentInstallTasks',
    BatchUpgradeAgent: 'aiTools.agents.batchUpgradeAgent',
    DispatchAgentUpgradeTasks: 'aiTools.agents.dispatchAgentUpgradeTasks',
    BatchInstallAgentSkill: 'aiTools.agents.batchInstallAgentSkill',
    DispatchAgentSkillInstallTasks: 'aiTools.agents.dispatchAgentSkillInstallTasks',
};

export const translateTaskText = (value?: string) => {
    if (!value) {
        return value || '';
    }
    return Object.entries(taskTextMap).reduce((text, [key, localeKey]) => {
        return text.replaceAll(key, i18n.global.t(localeKey));
    }, value);
};
