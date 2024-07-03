<template>
    <el-form-item prop="enable" :label="$t('website.enableOrNot')">
        <el-switch v-model="enable" @change="changeEnable" :disabled="data.length === 0"></el-switch>
    </el-form-item>
    <ComplexTable :data="data" @search="search" v-loading="loading">
        <template #toolbar>
            <el-button type="primary" plain @click="openCreate">
                {{ $t('commons.button.create') }}
            </el-button>
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
    <Create ref="createRef" @close="search()" />

    <OpDialog ref="opRef" @search="search" />
</template>

<script lang="ts" setup name="proxy">
import { Website } from '@/api/interface/website';
import { OperateAuthConfig, GetAuthConfig } from '@/api/modules/website';
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

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Website.NginxAuthConfig) {
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

const initData = (id: number): Website.NginxAuthConfig => ({
    websiteID: id,
    operate: 'create',
    username: '',
    password: '',
    remark: '',
});

const openCreate = () => {
    createRef.value.acceptParams(initData(id.value));
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
        api: OperateAuthConfig,
        params: authConfig,
    });
};

const changeEnable = () => {
    const req = initData(id.value);
    req.operate = enable.value ? 'enable' : 'disable';
    loading.value = true;
    OperateAuthConfig(req)
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
        const res = await GetAuthConfig({ websiteID: id.value });
        data.value = res.data.items || [];
        enable.value = res.data.enable;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    search();
});
</script>
