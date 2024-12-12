<template>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain @click="openCreate">
                {{ $t('commons.button.create') }}
            </el-button>
        </template>
        <el-table-column :label="$t('commons.table.name')" prop="name" min-width="60px" show-overflow-tooltip />
        <el-table-column :label="$t('website.sourceDomain')" prop="domain" min-width="80px" show-overflow-tooltip>
            <template #default="{ row }">
                <span v-if="row.type === 'domain'">{{ row.domains.join(',') }}</span>
                <span v-else>{{ row.path }}</span>
            </template>
        </el-table-column>
        <el-table-column :label="$t('commons.table.type')" prop="type" min-width="60px">
            <template #default="{ row }">
                <span v-if="row.type != 404">{{ $t('website.' + row.type) }}</span>
                <span v-else>{{ 404 }}</span>
            </template>
        </el-table-column>
        <el-table-column :label="$t('website.redirectWay')" prop="redirect" min-width="50px"></el-table-column>
        <el-table-column :label="$t('website.targetURL')" prop="target" min-width="100px" show-overflow-tooltip />
        <el-table-column :label="$t('website.keepPath')" prop="keepPath" min-width="80px" show-overflow-tooltip>
            <template #default="{ row }">
                <span v-if="row.type != '404'">
                    {{ row.keepPath ? $t('website.keep') : $t('website.notKeep') }}
                </span>
                <span v-else></span>
            </template>
        </el-table-column>
        <el-table-column :label="$t('commons.table.status')" prop="enable" min-width="50px">
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
            width="180px"
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
import { OperateRedirectConfig, GetRedirectConfig } from '@/api/modules/website';
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
        click: function (row: Website.RedirectConfig) {
            openEditFile(row);
        },
        disabled: (row: Website.RedirectConfig) => {
            return !row.enable;
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Website.RedirectConfig) {
            openEdit(row);
        },
        disabled: (row: Website.ProxyConfig) => {
            return !row.enable;
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.RedirectConfig) {
            deleteProxy(row);
        },
    },
];

const initData = (id: number): Website.RedirectConfig => ({
    websiteID: id,
    operate: 'create',
    enable: true,
    name: '',
    domains: [],
    keepPath: true,
    type: '',
    redirect: '',
    target: '',
});

const openCreate = () => {
    createRef.value.acceptParams(initData(id.value));
};

const openEdit = (proxyConfig: Website.RedirectConfig) => {
    let proxy = JSON.parse(JSON.stringify(proxyConfig));
    proxy.operate = 'edit';
    createRef.value.acceptParams(proxy);
};

const openEditFile = (proxyConfig: Website.RedirectConfig) => {
    fileRef.value.acceptParams({
        name: proxyConfig.name,
        content: proxyConfig.content,
        websiteID: proxyConfig.websiteID,
    });
};

const deleteProxy = async (redirectConfig: Website.RedirectConfig) => {
    redirectConfig.operate = 'delete';
    opRef.value.acceptParams({
        title: i18n.global.t('commons.msg.deleteTitle'),
        names: [redirectConfig.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.redirect'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: OperateRedirectConfig,
        params: redirectConfig,
    });
};

const submit = async (redirectConfig: Website.RedirectConfig) => {
    loading.value = true;
    await OperateRedirectConfig(redirectConfig)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const opProxy = (redirectConfig: Website.RedirectConfig) => {
    let proxy = JSON.parse(JSON.stringify(redirectConfig));
    proxy.enable = !redirectConfig.enable;
    let message = '';
    if (proxy.enable) {
        proxy.operate = 'enable';
        message = i18n.global.t('commons.button.start') + i18n.global.t('website.redirect');
    } else {
        proxy.operate = 'disable';
        message = i18n.global.t('commons.button.stop') + i18n.global.t('website.redirect');
    }
    ElMessageBox.confirm(message, i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    })
        .then(async () => {
            await submit(proxy);
        })
        .catch(() => {});
};

const search = async () => {
    try {
        loading.value = true;
        const res = await GetRedirectConfig({ websiteID: id.value });
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
