<template>
    <DrawerPro
        v-model="visible"
        :header="$t(enginesOnly ? 'aiTools.model.addInstance' : 'aiTools.model.addService')"
        :size="isMobile ? 'full' : '50%'"
    >
        <DockerStatus
            v-if="visible && selected !== 'account'"
            v-model:is-active="dockerActive"
            v-model:is-exist="dockerExists"
        />
        <div v-loading="initializing" class="service-form">
            <el-alert v-if="loadFailed" :title="$t('aiTools.model.loadFailed')" type="error" :closable="false" />
            <AccountForm
                v-if="activeForm === 'account'"
                ref="accountRef"
                inline
                @search="emit('search')"
                @close="visible = false"
            />
            <component
                :is="vllmForm"
                v-if="activeForm === 'vllm' && vllmForm"
                ref="vllmRef"
                inline
                @search="emit('search')"
                @close="visible = false"
                @task="openTask"
            />
            <template v-if="selected === 'vllm' && (!isProductPro || !vllmAvailable)">
                <el-alert :title="$t('license.vllm')" type="info" :closable="false" />
                <el-button class="mt-4" type="primary" @click="openVllmLicense">
                    {{ $t('license.levelUpPro') }}
                </el-button>
            </template>
        </div>
        <template v-if="activeForm" #footer>
            <el-button :disabled="currentForm?.loading" @click="visible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button
                v-permission
                type="primary"
                :loading="initializing || currentForm?.loading"
                :disabled="initializing || loadFailed || !currentForm || currentForm.disabled"
                @click="currentForm.submit()"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <AccountForm ref="editAccountRef" @search="emit('search')" />
    <TaskLog ref="taskRef" @close="emit('search')" />
</template>

<script setup lang="ts">
import { computed, nextTick, ref, shallowRef, watch } from 'vue';
import type { AI } from '@/api/interface/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { loadOptionalComponent } from '@/extensions/optional';
import AccountForm from '../account-form/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import { vllmAvailable } from '../../deployment';

const props = defineProps({ enginesOnly: Boolean });
const emit = defineEmits(['search', 'manage']);
const { isProductPro, isMobile } = useGlobalStore();
const visible = ref(false);
const selected = ref<'account' | 'vllm'>('account');
const modelDir = ref('');
const recipe = ref<{ id?: number; imageType: string }>();
const dockerActive = ref(false);
const dockerExists = ref(true);
const accountRef = ref();
const editAccountRef = ref();
const vllmRef = ref();
const taskRef = ref();
const vllmForm = shallowRef();
const initializing = ref(false);
const loadFailed = ref(false);
const activeForm = computed(() => {
    if (!visible.value) return '';
    if (selected.value === 'account') return 'account';
    if (!dockerActive.value || !dockerExists.value) return '';
    if (selected.value === 'vllm' && isProductPro.value && vllmAvailable) return 'vllm';
    return '';
});

const currentForm = computed(() => {
    if (activeForm.value === 'account') return accountRef.value;
    if (activeForm.value === 'vllm') return vllmRef.value;
    return undefined;
});

watch(
    activeForm,
    async (form, _, onCleanup) => {
        let cancelled = false;
        onCleanup(() => (cancelled = true));
        loadFailed.value = false;
        initializing.value = Boolean(form);
        if (!form) return;
        try {
            if (form === 'vllm' && !vllmForm.value) {
                const component = await loadOptionalComponent('/src/xpack/views/vllm/operate/index.vue');
                if (cancelled) return;
                vllmForm.value = component;
            }
            await nextTick();
            if (cancelled) return;
            if (form === 'account') await accountRef.value.open();
            if (form === 'vllm') await vllmRef.value.openCreate(modelDir.value, recipe.value);
        } catch {
            if (!cancelled) loadFailed.value = true;
        } finally {
            if (!cancelled) initializing.value = false;
        }
    },
    { flush: 'post' },
);

const open = (engine?: 'vllm', defaultModelDir = '', defaultRecipe?: { id?: number; imageType: string }) => {
    loadFailed.value = false;
    modelDir.value = defaultModelDir;
    recipe.value = defaultRecipe;
    selected.value = engine || (props.enginesOnly ? 'vllm' : 'account');
    dockerActive.value = false;
    dockerExists.value = true;
    visible.value = true;
};

const openAccount = (account: AI.AgentAccountItem) => {
    editAccountRef.value.open({ ...account, baseURL: account.baseUrl });
};

const openVllmLicense = () => {
    visible.value = false;
    emit('manage', { id: 0, name: '' });
};
const openTask = (taskID: string) => taskRef.value.openWithTaskID(taskID);

defineExpose({ open, openAccount });
</script>

<style scoped>
.service-form {
    min-height: 100px;
}
</style>
