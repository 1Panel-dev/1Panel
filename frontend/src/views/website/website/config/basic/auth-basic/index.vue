<template>
    <el-tabs type="border-card" @tab-change="searchAll()">
        <el-tab-pane :label="$t('website.global')">
            <ComplexTable :data="data" @search="search" v-loading="loading" :heightDiff="420">
                <template #toolbar>
                    <el-button type="primary" plain @click="openCreate('root')">
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <el-switch
                        class="ml-5"
                        v-model="enable"
                        @change="changeEnable"
                        :disabled="data.length === 0"
                    ></el-switch>
                </template>
                <el-table-column :label="$t('commons.login.username')" prop="username"></el-table-column>
                <el-table-column :label="$t('website.remark')" prop="remark"></el-table-column>
                <fu-table-operations
                    :ellipsis="10"
                    width="260px"
                    :buttons="buttons"
                    :label="$t('commons.table.operate')"
                    :fixed="mobile ? false : 'right'"
                    fix
                />
            </ComplexTable>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.path')">
            <ComplexTable :data="pathData" @search="searchPath" v-loading="loading" :heightDiff="420">
                <template #toolbar>
                    <el-button type="primary" plain @click="openCreate('path')">
                        {{ $t('commons.button.create') }}
                    </el-button>
                </template>
                <el-table-column :label="$t('commons.table.name')" prop="name"></el-table-column>
                <el-table-column :label="$t('website.path')" prop="path"></el-table-column>
                <el-table-column :label="$t('commons.login.username')" prop="username"></el-table-column>
                <el-table-column :label="$t('website.remark')" prop="remark"></el-table-column>
                <fu-table-operations
                    :ellipsis="10"
                    width="260px"
                    :buttons="pathButtons"
                    :label="$t('commons.table.operate')"
                    :fixed="mobile ? false : 'right'"
                    fix
                />
            </ComplexTable>
        </el-tab-pane>
    </el-tabs>

    <Create ref="createRef" @close="searchAll()" />
    <OpDialog ref="opRef" @search="searchAll" />
</template>

<script lang="ts" setup name="proxy">
import { Website } from '@/api/interface/website';
import { operateAuthConfig, getAuthConfig, getPathAuthConfig, operatePathAuthConfig } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import i18n from '@/lang';
import Create from './create/index.vue';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const mobile = computed(() => {
    return globalStore.isMobile();
});
const loading = ref(false);
const data = ref([]);
const createRef = ref();
const enable = ref(false);
const opRef = ref();
const pathData = ref([]);

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Website.NginxAuthConfig) {
            row.scope = 'root';
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.NginxAuthConfig) {
            deleteAuth(row);
        },
    },
];

const pathButtons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Website.NginxAuthConfig) {
            row.scope = 'path';
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.NginxPathAuthConfig) {
            deletePathAuth(row);
        },
    },
];

const initData = (id: number): Website.NginxAuthConfig => ({
    websiteID: id,
    operate: 'create',
    username: '',
    password: '',
    remark: '',
    scope: 'root',
});

const openCreate = (scope: string) => {
    let req = initData(id.value);
    req.scope = scope;
    createRef.value.acceptParams(req);
};

const openEdit = (authConfig: Website.NginxAuthConfig) => {
    let authParam = JSON.parse(JSON.stringify(authConfig));
    authParam.operate = 'edit';
    authParam.websiteID = id.value;
    createRef.value.acceptParams(authParam);
};

const deleteAuth = async (authConfig: Website.NginxAuthConfig) => {
    authConfig.operate = 'delete';
    authConfig.websiteID = id.value;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [authConfig.username],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.basicAuth'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: operateAuthConfig,
        params: authConfig,
    });
};

const deletePathAuth = async (authConfig: Website.NginxPathAuthConfig) => {
    authConfig.operate = 'delete';
    authConfig.websiteID = id.value;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [authConfig.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.basicAuth'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: operatePathAuthConfig,
        params: authConfig,
    });
};

const changeEnable = () => {
    const req = initData(id.value);
    req.operate = enable.value ? 'enable' : 'disable';
    loading.value = true;
    operateAuthConfig(req)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const search = async () => {
    try {
        loading.value = true;
        const res = await getAuthConfig({ websiteID: id.value });
        data.value = res.data.items || [];
        enable.value = res.data.enable;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const searchPath = async () => {
    try {
        loading.value = true;
        const res = await getPathAuthConfig({ websiteID: id.value });
        pathData.value = res.data || [];
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const searchAll = () => {
    search();
    searchPath();
};

onMounted(() => {
    searchAll();
});
</script>
