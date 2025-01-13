<template>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain @click="openCreate">{{ $t('commons.button.create') }}</el-button>
        </template>
        <el-table-column :label="$t('commons.table.name')" prop="name"></el-table-column>
        <el-table-column :label="$t('website.proxyPath')" prop="match"></el-table-column>
        <el-table-column :label="$t('website.proxyPass')" prop="proxyPass"></el-table-column>
        <el-table-column :label="$t('website.cache')" prop="cache">
            <template #default="{ row }">
                <el-switch v-model="row.cache" @change="changeCache(row)" :disabled="!row.enable"></el-switch>
            </template>
        </el-table-column>
        <el-table-column :label="$t('commons.table.status')" prop="enable">
            <template #default="{ row }">
                <el-button v-if="row.enable" link type="success" :icon="VideoPlay" @click="opProxy(row)">
                    {{ $t('commons.status.running') }}
                </el-button>
                <el-button v-else link type="danger" :icon="VideoPause" @click="opProxy(row)">
                    {{ $t('commons.status.stopped') }}
                </el-button>
            </template>
        </el-table-column>
        <fu-table-operations
            :ellipsis="10"
            width="260px"
            :buttons="buttons"
            :label="$t('commons.table.operate')"
            :fixed="mobile ? false : 'right'"
            fix
        />
    </ComplexTable>

    <Create ref="createRef" @close="search()" />
    <File ref="fileRef" @close="search()" />
    <OpDialog ref="opRef" @search="search()" />
</template>

<script lang="ts" setup name="proxy">
import { Website } from '@/api/interface/website';
import { OperateProxyConfig, GetProxyConfig, DelProxy } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import Create from './create/index.vue';
import File from './file/index.vue';
import { VideoPlay, VideoPause } from '@element-plus/icons-vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const mobile = computed(() => {
    return globalStore.isMobile();
});
const id = computed(() => {
    return props.id;
});
const loading = ref(false);
const data = ref();
const createRef = ref();
const fileRef = ref();
const opRef = ref();

const buttons = [
    {
        label: i18n.global.t('website.sourceFile'),
        click: function (row: Website.ProxyConfig) {
            openEditFile(row);
        },
        disabled: (row: Website.ProxyConfig) => {
            return !row.enable;
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Website.ProxyConfig) {
            openEdit(row);
        },
        disabled: (row: Website.ProxyConfig) => {
            return !row.enable;
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.ProxyConfig) {
            deleteProxy(row);
        },
    },
];

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
    replaces: {},
    sni: false,
    proxySSLName: '',
});

const openCreate = () => {
    createRef.value.acceptParams(initData(id.value));
};

const openEdit = (proxyConfig: Website.ProxyConfig) => {
    let proxy = JSON.parse(JSON.stringify(proxyConfig));
    proxy.operate = 'edit';
    if (proxy.replaces == null) {
        proxy.replaces = {};
    }
    createRef.value.acceptParams(proxy);
};

const openEditFile = (proxyConfig: Website.ProxyConfig) => {
    fileRef.value.acceptParams({ name: proxyConfig.name, content: proxyConfig.content, websiteID: proxyConfig.id });
};

const deleteProxy = async (proxyConfig: Website.ProxyConfig) => {
    const del = {
        id: proxyConfig.id,
        name: proxyConfig.name,
    };
    opRef.value.acceptParams({
        title: i18n.global.t('commons.msg.deleteTitle'),
        names: [proxyConfig.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.proxy'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: DelProxy,
        params: del,
    });
};

const changeCache = (proxyConfig: Website.ProxyConfig) => {
    proxyConfig.operate = 'edit';
    if (proxyConfig.cache) {
        proxyConfig.cacheTime = 1;
        proxyConfig.cacheUnit = 'm';
    }
    submit(proxyConfig);
};

const submit = async (proxyConfig: Website.ProxyConfig) => {
    loading.value = true;
    OperateProxyConfig(proxyConfig)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const opProxy = (proxyConfig: Website.ProxyConfig) => {
    let proxy = JSON.parse(JSON.stringify(proxyConfig));
    proxy.enable = !proxyConfig.enable;
    let message = '';
    if (proxy.enable) {
        proxy.operate = 'enable';
        message = i18n.global.t('website.startProxy');
    } else {
        proxy.operate = 'disable';
        message = i18n.global.t('website.stopProxy');
    }
    ElMessageBox.confirm(message, i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            await submit(proxy);
            search();
        })
        .catch(() => {});
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
