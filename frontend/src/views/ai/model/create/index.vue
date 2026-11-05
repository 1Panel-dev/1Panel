<template>
    <DrawerPro v-model="visible" :header="$t('aiTools.model.addService')" :size="isMobile ? 'full' : '50%'">
        <div v-loading="initializing" class="service-form">
            <el-alert v-if="loadFailed" :title="$t('aiTools.model.loadFailed')" type="error" :closable="false" />
            <AccountForm v-if="visible" ref="accountRef" inline @search="emit('search')" @close="visible = false" />
        </div>
        <template v-if="visible" #footer>
            <el-button :disabled="accountRef?.loading" @click="visible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button
                v-permission
                type="primary"
                :loading="initializing || accountRef?.loading"
                :disabled="initializing || loadFailed || !accountRef || accountRef.disabled"
                @click="accountRef.submit()"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <AccountForm ref="editAccountRef" @search="emit('search')" />
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import type { AI } from '@/api/interface/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';
import AccountForm from '../account-form/index.vue';

const emit = defineEmits(['search']);
const { isMobile } = useGlobalStore();
const visible = ref(false);
const accountRef = ref();
const editAccountRef = ref();
const initializing = ref(false);
const loadFailed = ref(false);
watch(
    visible,
    async (value, _, onCleanup) => {
        let cancelled = false;
        onCleanup(() => (cancelled = true));
        loadFailed.value = false;
        initializing.value = value;
        if (!value) return;
        try {
            await nextTick();
            if (cancelled) return;
            await accountRef.value.open();
        } catch {
            if (!cancelled) loadFailed.value = true;
        } finally {
            if (!cancelled) initializing.value = false;
        }
    },
    { flush: 'post' },
);
const open = () => {
    loadFailed.value = false;
    visible.value = true;
};
const openAccount = (account: AI.AgentAccountItem) => {
    editAccountRef.value.open({ ...account, baseURL: account.baseUrl });
};
defineExpose({ open, openAccount });
</script>

<style scoped>
.service-form {
    min-height: 100px;
}
</style>
