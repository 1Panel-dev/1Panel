<template>
    <el-dialog
        v-model="open"
        :title="$t('website.batchPHP')"
        width="600px"
        :close-on-click-modal="false"
        :close-on-press-escape="!loading"
        :show-close="!loading"
    >
        <el-alert :title="$t('website.batchPHPHelper')" type="info" :closable="false" />
        <p>{{ $t('website.batchPHPCount', [sites.length, eligible.length, sites.length - eligible.length]) }}</p>
        <el-select v-model="runtimeID" :disabled="loading || !!results" :loading="fetching" class="w-full">
            <el-option
                v-for="item in versions"
                :key="item.id"
                :value="item.id"
                :label="item.name + ' [' + item.version + ']'"
                :disabled="item.resource === 'local'"
            />
        </el-select>
        <el-table v-if="results" :data="results" max-height="350">
            <el-table-column prop="domain" :label="$t('website.primaryDomain')" />
            <el-table-column :label="$t('website.batchPHPResult')">
                <template #default="{ row }">
                    {{ $t('website.batchPHPStatus_' + row.status) }}
                    <span v-if="row.message">
                        :
                        {{
                            ['notPHP', 'localPHP', 'samePHP'].includes(row.message)
                                ? $t('website.' + row.message)
                                : row.message
                        }}
                    </span>
                </template>
            </el-table-column>
        </el-table>
        <template #footer>
            <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.close') }}</el-button>
            <el-button
                v-if="!results"
                v-permission
                type="primary"
                :loading="loading"
                :disabled="fetching || !runtimeID || !eligible.length"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </el-dialog>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';
import { Website } from '@/api/interface/website';
import { SearchRuntimes } from '@/api/modules/runtime';
import { batchChangePHPVersion } from '@/api/modules/website';
const emit = defineEmits(['updated']);
type Site = { id: number; primaryDomain: string; type: string; runtimeType?: string };
const open = ref(false);
const loading = ref(false);
const fetching = ref(false);
const runtimeID = ref<number>();
const sites = ref<Site[]>([]);
const versions = ref<{ id: number; name: string; version: string; resource: string }[]>([]);
const results = ref<Website.BatchPHPVersionResult[]>();
const eligible = computed(() => sites.value.filter((site) => site.type === 'runtime' && site.runtimeType === 'php'));
let runtimeRequestID = 0;
const acceptParams = async (selected: Site[]) => {
    const requestID = ++runtimeRequestID;
    sites.value = [...selected];
    runtimeID.value = undefined;
    results.value = undefined;
    versions.value = [];
    open.value = true;
    fetching.value = true;
    try {
        const options: typeof versions.value = [];
        let page = 1;
        while (true) {
            const { data } = await SearchRuntimes({ page, pageSize: 200, type: 'php' });
            if (requestID !== runtimeRequestID || !open.value) return;
            const items = data.items || [];
            options.push(
                ...items.map((item) => ({
                    id: item.id,
                    name: item.name,
                    resource: item.resource,
                    version: item.params?.['PHP_VERSION'] || item.version,
                })),
            );
            if (!items.length || options.length >= data.total) break;
            page++;
        }
        versions.value = options;
    } catch {
        // The request interceptor reports loading errors. Keep the list empty.
    } finally {
        if (requestID === runtimeRequestID) fetching.value = false;
    }
};
const submit = async () => {
    if (loading.value || fetching.value || !runtimeID.value || !eligible.value.length) return;
    loading.value = true;
    try {
        const { data } = await batchChangePHPVersion({
            ids: sites.value.map((site) => site.id),
            runtimeID: runtimeID.value,
        });
        results.value = data;
    } finally {
        loading.value = false;
        emit('updated');
    }
};
defineExpose({ acceptParams });
</script>
