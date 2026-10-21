<template>
    <DrawerPro v-model="visible" :header="$t('setting.panelCache.title')" size="normal">
        <el-card shadow="never">
            <div v-for="key in categories" :key="key" class="flex items-center justify-between py-3">
                <span>{{ $t(`setting.panelCache.${key}`) }}</span>
                <el-switch v-model="form[key]" :aria-label="$t(`setting.panelCache.${key}`)" />
            </div>
            <div class="flex items-center justify-between gap-4 mt-4">
                <div class="flex items-center gap-2 shrink-0">
                    <span>{{ $t('setting.panelCache.duration') }}</span>
                    <el-tooltip
                        :content="$t('setting.panelCache.helper')"
                        placement="top"
                        :trigger="isMobile ? 'click' : 'hover'"
                        :popper-style="{ maxWidth: 'min(360px, calc(100vw - 32px))' }"
                    >
                        <el-button
                            link
                            type="info"
                            icon="InfoFilled"
                            class="!ml-0 shrink-0"
                            :aria-label="$t('setting.panelCache.helper')"
                            @click.stop
                        />
                    </el-tooltip>
                </div>
                <el-select
                    v-model="form.duration"
                    class="!w-40 min-w-0"
                    :aria-label="$t('setting.panelCache.duration')"
                >
                    <el-option
                        v-for="value in durations"
                        :key="value"
                        :value="value"
                        :label="$t(`setting.panelCache.${value}`)"
                    />
                </el-select>
            </div>
        </el-card>
        <template #footer>
            <el-button :disabled="loading" @click="Object.assign(form, defaultPanelCache)">
                {{ $t('commons.button.setDefault') }}
            </el-button>
            <el-button :disabled="loading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :loading="loading" type="primary" @click="save">{{ $t('commons.button.confirm') }}</el-button>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { updateSetting } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { applyPanelCacheSettings, defaultPanelCache } from '@/utils/panel-cache';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const { panelCache, isMobile } = useGlobalStore();
const visible = ref(false);
const loading = ref(false);
const form = reactive({ ...defaultPanelCache });
const categories = ['tabs', 'pagination', 'refresh', 'search'] as const;
const durations = ['24h', 'week', 'month', 'forever'] as const;
const acceptParams = () => {
    Object.assign(form, panelCache.value);
    visible.value = true;
};
const save = async () => {
    loading.value = true;
    try {
        const value = JSON.stringify(form);
        await updateSetting({ key: 'PanelCache', value });
        applyPanelCacheSettings(value);
        visible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        loading.value = false;
    }
};
defineExpose({ acceptParams });
</script>
