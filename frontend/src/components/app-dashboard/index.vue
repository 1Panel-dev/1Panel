<template>
    <el-popover
        placement="right"
        trigger="hover"
        :disabled="disabled || !installed"
        :popper-style="{ width: 'fit-content', maxWidth: 'min(92vw, 560px)' }"
    >
        <template #reference>
            <slot name="reference" :disabled="disabled || !loaded" :on-click="onClick">
                <el-button type="primary" plain :disabled="disabled || !loaded" @click="onClick">
                    <slot />
                </el-button>
            </slot>
        </template>
        <div class="dashboard-links">
            <el-button v-for="link in links" :key="link" type="primary" link @click="toLink(link)">
                {{ link }}
            </el-button>
            <span v-if="!links.length">
                {{ $t('app.webUIConfig') }}
                <el-link icon="Position" type="primary" @click="routerToName('Panel')">
                    {{ $t('firewall.quickJump') }}
                </el-link>
            </span>
        </div>
    </el-popover>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { App } from '@/api/interface/app';
import { getAppInstalled } from '@/api/modules/app';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { toLink } from '@/utils/browser';
import { routerToName } from '@/utils/router';

const props = defineProps<{ appKey: string; disabled?: boolean }>();
const emit = defineEmits<{ notInstalled: [] }>();
const { currentNodeAddr, isMaster } = useGlobalStore();
const installed = ref<App.AppInstalled>();
const host = ref('');
const loaded = ref(false);
const links = computed(() => {
    const app = installed.value;
    if (!app) return [];
    const urls: string[] = [];
    if (host.value) {
        const address = host.value.includes(':') && !host.value.startsWith('[') ? `[${host.value}]` : host.value;
        if (app.httpPort > 0) urls.push(`http://${address}:${app.httpPort}`);
        if (app.httpsPort > 0) urls.push(`https://${address}:${app.httpsPort}`);
    }
    if (app.webUI) urls.push(app.webUI);
    return [...new Set(urls)];
});
const onClick = () => {
    if (loaded.value && !props.disabled && !installed.value) emit('notInstalled');
};
onMounted(async () => {
    try {
        const settings = await getAgentSettingInfo();
        host.value =
            settings.data.systemIP ||
            (!isMaster.value || currentNodeAddr.value !== '127.0.0.1' ? currentNodeAddr.value : '');
        let page = 1;
        let count = 0;
        let total = 0;
        do {
            const res = await getAppInstalled({ page: page++, pageSize: 100, type: '' });
            const items = res.data.items || [];
            installed.value = items.find((app) => app.appKey === props.appKey);
            count += items.length;
            total = res.data.total;
            if (installed.value || !items.length) break;
        } while (count < total);
        loaded.value = true;
    } catch {
        // API errors are reported by the shared request handler; keep the button disabled.
    }
});
</script>

<style scoped lang="scss">
.dashboard-links {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    max-height: 65vh;
    overflow-y: auto;
    .el-button.is-link {
        margin-left: 0;
        white-space: normal;
        text-align: left;
        height: auto;
        line-height: 1.4;
        word-break: break-all;
    }
}
</style>
