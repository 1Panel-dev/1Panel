<template>
    <DrawerPro v-model="visible" :header="$t('container.imageRemoteSearch')" size="large" @close="cancelRequest">
        <el-form @submit.prevent="search(1)">
            <el-form-item>
                <el-input v-model="query" :maxlength="256" clearable :placeholder="$t('commons.button.search')">
                    <template #append>
                        <el-button :loading="loading" :disabled="!query.trim()" @click="search(1)">
                            {{ $t('commons.button.search') }}
                        </el-button>
                    </template>
                </el-input>
                <span class="input-help">{{ $t('container.imageRemoteSearchHelper') }}</span>
            </el-form-item>
        </el-form>
        <el-table :tooltip-options="TABLE_OVERFLOW_TOOLTIP_OPTIONS" v-loading="loading" :data="items">
            <el-table-column prop="name" :label="$t('container.imageName')" min-width="180" />
            <el-table-column
                prop="description"
                :label="$t('commons.table.description')"
                min-width="240"
                show-overflow-tooltip
            />
            <el-table-column :label="$t('commons.table.operate')" width="90">
                <template #default="{ row }">
                    <el-button v-permission link type="primary" @click="emit('pull', row.name)">
                        {{ $t('container.pull') }}
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
        <el-pagination
            v-if="total > 0"
            class="mt-4"
            :current-page="page"
            :page-size="20"
            :total="total"
            :disabled="loading"
            layout="total, prev, pager, next"
            @current-change="search($event, true)"
        />
    </DrawerPro>
</template>

<script setup lang="ts">
import { TABLE_OVERFLOW_TOOLTIP_OPTIONS } from '@/components/table/shared';
import { onBeforeUnmount, ref, watch } from 'vue';
import { searchRemoteImages, type RemoteImage } from '@/api/modules/container';
import { getSettingBaseInfo } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { showImageProxyRequired, isRuntimeProxyRequired } from '@/utils/image-remote';

const emit = defineEmits<{ (e: 'pull', name: string): void }>();
const { currentNode } = useGlobalStore();
const visible = ref(false);
const loading = ref(false);
const opening = ref(false);
const query = ref('');
const submittedQuery = ref('');
const page = ref(1);
const total = ref(0);
const items = ref<RemoteImage[]>([]);
let controller: AbortController | undefined;
let generation = 0;
const cancelRequest = () => {
    generation++;
    controller?.abort();
    controller = undefined;
    loading.value = false;
    opening.value = false;
};
const warnProxy = (node: string) =>
    showImageProxyRequired('container.imageRemoteSearchProxyRequired', () => node === currentNode.value);
const open = async () => {
    if (opening.value) return;
    cancelRequest();
    const version = generation;
    const node = currentNode.value;
    opening.value = true;
    try {
        if (node === 'local') {
            const res = await getSettingBaseInfo();
            if (version !== generation || node !== currentNode.value) return;
            if (isRuntimeProxyRequired(res.data)) {
                await warnProxy(node);
                return;
            }
        }
        visible.value = true;
    } catch {
        // The request interceptor displays errors.
    } finally {
        if (version === generation) opening.value = false;
    }
};
const search = async (nextPage = 1, paginate = false) => {
    const keyword = paginate ? submittedQuery.value : query.value.trim();
    if (!keyword || !visible.value) return;
    cancelRequest();
    const version = generation;
    const node = currentNode.value;
    controller = new AbortController();
    loading.value = true;
    if (keyword !== submittedQuery.value) {
        items.value = [];
        total.value = 0;
        page.value = 1;
    }
    try {
        const res = await searchRemoteImages({ query: keyword, page: nextPage, pageSize: 20 }, node, controller.signal);
        if (version !== generation || node !== currentNode.value || !visible.value) return;
        if (res.data.proxyRequired) {
            await warnProxy(node);
            return;
        }
        items.value = res.data.items;
        total.value = res.data.total;
        page.value = nextPage;
        submittedQuery.value = keyword;
    } catch {
        // The request interceptor displays errors. Keep the last successful page when pagination fails.
    } finally {
        if (version === generation) loading.value = false;
    }
};
watch(currentNode, () => {
    cancelRequest();
    visible.value = false;
    query.value = '';
    submittedQuery.value = '';
    items.value = [];
    total.value = 0;
    page.value = 1;
});
onBeforeUnmount(cancelRequest);
defineExpose({ open });
</script>
