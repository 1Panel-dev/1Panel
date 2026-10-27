<template>
    <el-form-item v-if="showSync">
        <el-checkbox v-model="form.syncModelAccount">{{ $t('aiTools.model.syncModelAccount') }}</el-checkbox>
    </el-form-item>
    <template v-if="!showSync || form.syncModelAccount">
        <span class="input-help model-account-helper">{{ $t('aiTools.model.modelAccountAddressHelper') }}</span>
        <el-form-item
            :label="$t('aiTools.agents.allowedOrigins')"
            prop="modelAccountBaseURLType"
            :rules="Rules.requiredSelect"
        >
            <el-radio-group v-model="form.modelAccountBaseURLType">
                <el-radio-button
                    v-for="item in addressTypes"
                    :key="item.value"
                    :value="item.value"
                    :disabled="hostMode && item.value === 'container'"
                >
                    {{ item.label }}
                </el-radio-button>
            </el-radio-group>
        </el-form-item>
        <el-form-item
            v-if="form.modelAccountBaseURLType === 'custom'"
            label="Base URL"
            prop="modelAccountBaseURL"
            :rules="customURLRules"
        >
            <el-input v-model.trim="form.modelAccountBaseURL" placeholder="http://192.168.1.100:8000/v1" />
        </el-form-item>
        <el-form-item v-else label="Base URL">
            <el-input :model-value="baseURL" readonly />
        </el-form-item>
    </template>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';

const props = defineProps<{
    form: {
        name: string;
        syncModelAccount?: boolean;
        containerName: string;
        modelAccountBaseURLType: string;
        modelAccountBaseURL: string;
    };
    hostPort: number;
    containerPort: string;
    hostMode?: boolean;
    showSync?: boolean;
}>();
const systemIP = ref('');
const addressTypes = computed(() => [
    { label: i18n.global.t('aiTools.model.containerAddress'), value: 'container' },
    { label: '127.0.0.1', value: 'localhost' },
    { label: i18n.global.t('setting.systemIP'), value: 'systemIP' },
    { label: i18n.global.t('aiTools.model.customAddress'), value: 'custom' },
]);
const customURLRules = [
    Rules.requiredInput,
    { pattern: /^https?:\/\/.+/i, message: i18n.global.t('aiTools.agents.allowedOriginsInvalid'), trigger: 'blur' },
];
const baseURL = computed(() => {
    if (props.form.modelAccountBaseURLType === 'custom') return props.form.modelAccountBaseURL;
    if (props.form.modelAccountBaseURLType === 'localhost') return `http://127.0.0.1:${props.hostPort}/v1`;
    if (props.form.modelAccountBaseURLType === 'systemIP')
        return `http://${systemIP.value || '127.0.0.1'}:${props.hostPort}/v1`;
    return `http://${props.form.containerName || props.form.name}:${props.containerPort}/v1`;
});
onMounted(async () => {
    try {
        const { data } = await getAgentSettingInfo();
        systemIP.value = data.systemIP || '';
    } catch {
        systemIP.value = '';
    }
});
defineExpose({ baseURL });
</script>

<style scoped>
.model-account-helper {
    display: block;
    margin-bottom: 12px;
}
</style>
