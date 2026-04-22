<template>
    <OpenclawSkills v-if="agentType === 'openclaw'" ref="openclawRef" :app-version="appVersion" />
    <HermesSkills v-else-if="agentType === 'hermes-agent'" ref="hermesRef" />
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { AI } from '@/api/interface/ai';
import OpenclawSkills from './skills/openclaw.vue';
import HermesSkills from './skills/hermes.vue';

const props = defineProps<{
    appVersion: string;
    agentType: AI.AgentType;
}>();

const openclawRef = ref<InstanceType<typeof OpenclawSkills>>();
const hermesRef = ref<InstanceType<typeof HermesSkills>>();

const load = async (id: number) => {
    if (props.agentType === 'openclaw') {
        await openclawRef.value?.load(id);
        return;
    }
    await hermesRef.value?.load(id);
};

defineExpose({
    load,
});
</script>
