import { ref } from 'vue';
import { AI } from '@/api/interface/ai';
import { checkAgentPlugin, installAgentPlugin } from '@/api/modules/ai';
import { newUUID } from '@/utils/util';

export const useAgentPluginChannel = (pluginType: AI.AgentPluginInstallReq['type']) => {
    const agentId = ref(0);
    const installed = ref(false);
    const installing = ref(false);
    const taskLogRef = ref();

    const checkPluginStatus = async () => {
        if (!agentId.value) {
            return false;
        }
        const res = await checkAgentPlugin({
            agentId: agentId.value,
            type: pluginType,
        });
        installed.value = Boolean(res.data?.installed);
        return installed.value;
    };

    const loadPlugin = async (id: number) => {
        agentId.value = id;
        return await checkPluginStatus();
    };

    const installPlugin = async () => {
        if (!agentId.value) {
            return;
        }
        const taskID = newUUID();
        installing.value = true;
        try {
            await installAgentPlugin({
                agentId: agentId.value,
                type: pluginType,
                taskID,
            });
            taskLogRef.value?.openWithTaskID(taskID);
        } finally {
            installing.value = false;
        }
    };

    return {
        agentId,
        installed,
        installing,
        taskLogRef,
        checkPluginStatus,
        loadPlugin,
        installPlugin,
    };
};
