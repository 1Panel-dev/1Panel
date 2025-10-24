<template>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain @click="openCreate">{{ $t('commons.button.create') }}</el-button>
            <el-button @click="openCache">{{ $t('website.proxyCache') }}</el-button>
            <el-button type="primary" @click="clear" link>
                {{ $t('nginx.clearProxyCache') }}
            </el-button>
        </template>
        <el-table-column :label="$t('commons.table.name')" prop="name"></el-table-column>
        <el-table-column :label="$t('website.proxyPath')" prop="match"></el-table-column>
        <el-table-column :label="$t('website.proxyPass')" prop="proxyPass"></el-table-column>
        <el-table-column :label="$t('website.cache')" prop="cache">
            <template #default="{ row }">
                <el-tag :type="row.cacheTime > 0 ? 'success' : 'info'">
                    {{ $t('website.browserCache') + ':' }}
                    {{ row.cacheTime > 0 ? row.cacheTime + row.cacheUnit : $t('setting.sslDisable') }}
                </el-tag>
                <el-tag class="ml-2" :type="row.serverCacheTime > 0 ? 'success' : 'info'">
                    {{ $t('website.serverCache') + ':' }}
                    {{ row.serverCacheTime > 0 ? row.serverCacheTime + row.serverCacheUnit : $t('setting.sslDisable') }}
                </el-tag>
            </template>
        </el-table-column>
        <el-table-column :label="$t('commons.table.status')" prop="enable" width="100">
            <template #default="{ row }">
                <Status :status="row.enable ? 'enable' : 'disable'" @click="opProxy(row)" :operate="true" />
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
    <Cache ref="cacheRef" @close="search()" />
</template>

<script lang="ts" setup name="proxy">
import { Website } from '@/api/interface/website';
import { operateProxyConfig, getProxyConfig, clearProxyCache } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { GlobalStore } from '@/store';
import Create from './create/index.vue';
import File from './file/index.vue';
import Cache from './cache/index.vue';
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
const cacheRef = ref();
const hasCache = ref(false);

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
    cacheTime: 0,
    cacheUnit: '',
    name: '',
    modifier: '',
    match: '/',
    proxyPass: 'http://',
    proxyHost: '$host',
    replaces: {},
    proxySSLName: '',
    serverCacheTime: 10,
    serverCacheUnit: 'm',
    cors: false,
    allowOrigins: '',
    allowMethods: '',
    allowHeaders: '',
    allowCredentials: false,
    preflight: false,
});

const openCreate = () => {
    createRef.value.acceptParams(initData(id.value));
};

const openCache = () => {
    cacheRef.value.acceptParams(id.value, hasCache.value);
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
    proxyConfig.operate = 'delete';
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [proxyConfig.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.proxy'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: operateProxyConfig,
        params: proxyConfig,
    });
};

const submit = async (proxyConfig: Website.ProxyConfig) => {
    loading.value = true;
    operateProxyConfig(proxyConfig)
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
        const res = await getProxyConfig({ id: id.value });
        data.value = res.data || [];
        hasCache.value = data.value.some((item: Website.ProxyConfig) => item.cache);
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const clear = () => {
    ElMessageBox.confirm(i18n.global.t('nginx.clearProxyCacheWarn'), i18n.global.t('nginx.clearProxyCache'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await clearProxyCache({ websiteID: id.value });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

onMounted(() => {
    search();
});
</script>
