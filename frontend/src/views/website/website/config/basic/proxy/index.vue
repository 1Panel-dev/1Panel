<template>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain @click="openCreate">{{ $t('website.addProxy') }}</el-button>
        </template>
        <el-table-column :label="$t('commons.table.name')" prop="name"></el-table-column>
        <el-table-column :label="$t('website.proxyPath')" prop="match"></el-table-column>
        <el-table-column :label="$t('website.proxyPass')" prop="proxyPass"></el-table-column>
        <el-table-column :label="$t('website.cache')" prop="cache">
            <template #default="{ row }">
                <el-switch v-model="row.cache"></el-switch>
            </template>
        </el-table-column>
        <el-table-column :label="$t('commons.table.status')" prop="enable">
            <template #default="{ row }">
                <el-button v-if="row.enable" link type="success" :icon="VideoPlay">
                    {{ $t('commons.status.running') }}
                </el-button>
                <el-button v-else link type="danger" :icon="VideoPause">
                    {{ $t('commons.status.stopped') }}
                </el-button>
            </template>
        </el-table-column>
    </ComplexTable>
    <Create ref="createRef" @close="search()" />
</template>

<script lang="ts" setup name="proxy">
import { Website } from '@/api/interface/website';
import { GetProxyConfig } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import Create from './create/index.vue';
import { VideoPlay, VideoPause } from '@element-plus/icons-vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const loading = ref(false);
const data = ref();
const createRef = ref();

const initData = (id: number): Website.ProxyConfig => ({
    id: id,
    operate: 'create',
    enable: true,
    cache: false,
    cacheTime: 1,
    cacheUnit: 'm',
    name: '',
    modifier: '^~',
    match: '/',
    proxyPass: 'http://',
    proxyHost: '$host',
});

const openCreate = () => {
    createRef.value.acceptParams(initData(id.value));
};
const search = async () => {
    try {
        loading.value = true;
        const res = await GetProxyConfig({ id: id.value });
        data.value = res.data || [];
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    search();
});
</script>
