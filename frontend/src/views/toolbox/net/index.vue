<template>
    <div>
        <el-card>
            <el-form :inline="true" @submit.prevent>
                <el-form-item :label="$t('toolbox.net.tool')">
                    <el-select v-model="form.tool" style="width: 160px">
                        <el-option label="Ping" value="ping" />
                        <el-option label="Traceroute" value="traceroute" />
                        <el-option :label="$t('toolbox.net.dnsLookup')" value="dns" />
                        <el-option :label="$t('toolbox.net.httpCheck')" value="http" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('toolbox.net.host')">
                    <el-input
                        v-model.trim="form.host"
                        :placeholder="hostPlaceholder"
                        style="width: 300px"
                        @keyup.enter="onRun"
                    />
                </el-form-item>
                <el-form-item v-if="form.tool === 'ping'">
                    <el-input-number v-model="form.count" :min="1" :max="20" :step="1" style="width: 100px" />
                    <span class="ml-2 text-sm text-gray-500">{{ $t('toolbox.net.pingCount') }}</span>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="onRun" :loading="loading">
                        {{ $t('toolbox.net.run') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-card>

        <el-card class="mt-4" v-if="output || loading">
            <template #header>
                <span>{{ $t('toolbox.net.result') }}</span>
            </template>
            <div v-loading="loading">
                <pre class="net-output">{{ output || $t('toolbox.net.waiting') }}</pre>
            </div>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { runNetTool } from '@/api/modules/toolbox-net';
import i18n from '@/lang';

const loading = ref(false);
const output = ref('');

const form = reactive({
    tool: 'ping' as 'ping' | 'traceroute' | 'dns' | 'http',
    host: '',
    count: 4,
});

const hostPlaceholder = computed(() => {
    switch (form.tool) {
        case 'ping':
        case 'traceroute':
        case 'dns':
            return 'example.com';
        case 'http':
            return 'https://example.com';
        default:
            return '';
    }
});

const onRun = async () => {
    if (!form.host) return;
    loading.value = true;
    output.value = '';
    try {
        const res = await runNetTool({
            tool: form.tool,
            host: form.host,
            count: form.tool === 'ping' ? form.count : undefined,
        });
        output.value = res.data.output;
    } catch (e: any) {
        output.value = e.message || i18n.global.t('toolbox.net.failed');
    } finally {
        loading.value = false;
    }
};
</script>

<style scoped>
.net-output {
    font-family: 'Courier New', Courier, monospace;
    font-size: 13px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--el-fill-color-light);
    padding: 16px;
    border-radius: 4px;
    max-height: 500px;
    overflow-y: auto;
}
</style>
