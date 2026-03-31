import { ref } from 'vue';
import { AI } from '@/api/interface/ai';
import { checkAgentPlugin, installAgentPlugin, uninstallAgentPlugin, upgradeAgentPlugin } from '@/api/modules/ai';
import { newUUID } from '@/utils/util';

export const useAgentPluginChannel = (pluginType: AI.AgentPluginInstallReq['type']) => {
    const agentId = ref(0);
    const loading = ref(false);
    const installed = ref(false);
    const installing = ref(false);
    const upgrading = ref(false);
    const uninstalling = ref(false);
    const currentVersion = ref('');
    const latestVersion = ref('');
    const upgradable = ref(false);
    let loadVersion = 0;

    const applyPluginStatus = (status?: AI.AgentPluginStatus) => {
        installed.value = Boolean(status?.installed);
        currentVersion.value = status?.currentVersion || '';
        latestVersion.value = status?.latestVersion || '';
        upgradable.value = Boolean(status?.upgradable);
    };

    const checkPluginStatus = async (checkLatest = false) => {
        if (!agentId.value) {
            return false;
        }
        const res = await checkAgentPlugin({
            agentId: agentId.value,
            type: pluginType,
            checkLatest,
        });
        applyPluginStatus(res.data);
        return Boolean(res.data?.installed);
    };

    const checkPluginLatestVersion = async (version: number) => {
        if (!agentId.value) {
            return;
        }
        try {
            const res = await checkAgentPlugin({
                agentId: agentId.value,
                type: pluginType,
                checkLatest: true,
            });
            if (version !== loadVersion) {
                return;
            }
            applyPluginStatus(res.data);
        } catch {}
    };

    const loadPlugin = async (id: number) => {
        loadVersion += 1;
        const version = loadVersion;
        agentId.value = id;
        loading.value = true;
        try {
            latestVersion.value = '';
            upgradable.value = false;
            const isInstalled = await checkPluginStatus();
            if (isInstalled) {
                void checkPluginLatestVersion(version);
            }
            return isInstalled;
        } finally {
            loading.value = false;
        }
    };

    const installPlugin = async () => {
        if (!agentId.value) {
            return '';
        }
        const taskID = newUUID();
        installing.value = true;
        try {
            await installAgentPlugin({
                agentId: agentId.value,
                type: pluginType,
                taskID,
            });
            return taskID;
        } finally {
            installing.value = false;
        }
    };

    const upgradePlugin = async () => {
        if (!agentId.value) {
            return '';
        }
        const taskID = newUUID();
        upgrading.value = true;
        try {
            await upgradeAgentPlugin({
                agentId: agentId.value,
                type: pluginType,
                taskID,
            });
            return taskID;
        } finally {
            upgrading.value = false;
        }
    };

    const uninstallPlugin = async () => {
        if (!agentId.value) {
            return '';
        }
        const taskID = newUUID();
        uninstalling.value = true;
        try {
            await uninstallAgentPlugin({
                agentId: agentId.value,
                type: pluginType,
                taskID,
            });
            return taskID;
        } finally {
            uninstalling.value = false;
        }
    };

    return {
        agentId,
        loading,
        installed,
        installing,
        upgrading,
        uninstalling,
        currentVersion,
        latestVersion,
        upgradable,
        checkPluginStatus,
        loadPlugin,
        installPlugin,
        upgradePlugin,
        uninstallPlugin,
    };
};
