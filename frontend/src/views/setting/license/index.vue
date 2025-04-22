<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.license')">
            <template #leftToolBar>
                <el-button type="primary" @click="toUpload()">
                    {{ $t('commons.button.add') }}
                </el-button>
            </template>
            <template #prompt>
                <el-alert type="info" :title="$t('license.licenseAlert')" :closable="false" />
            </template>
            <template #rightToolBar>
                <TableRefresh @search="search()" />
                <TableSetting title="backup-account-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" @sort-change="search" @search="search" :data="data">
                    <el-table-column
                        :label="$t('license.authorizationId')"
                        :min-width="80"
                        prop="licenseName"
                        show-overflow-tooltip
                    />
                    <el-table-column :label="$t('license.authorizedUser')" prop="assigneeName" show-overflow-tooltip>
                        <template #default="{ row }">
                            {{ row.assigneeName || '-' }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('license.trialInfo')" show-overflow-tooltip :min-width="120">
                        <template #default="{ row }">
                            {{ loadVersion(row) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" show-overflow-tooltip>
                        <template #default="{ row }">
                            <div v-if="row.status">
                                <Status :status="row.status" :msg="loadMsg(row)"></Status>
                            </div>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('setting.bindNode')" :min-width="120">
                        <template #default="{ row }">
                            <div v-if="row.status !== 'Free'">
                                <div>
                                    {{ $t('license.pro') }}:
                                    {{ row.bindNode === '127.0.0.1' ? $t('xpack.node.master') : row.bindNode }}
                                </div>
                                <div>
                                    {{ $t('license.oss') }}:
                                    <el-button class="bind-button" @click="onBindFree(row)" link type="primary">
                                        {{ row.bindCount }} / {{ row.freeCount }}
                                    </el-button>
                                </div>
                            </div>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="300px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <LicenseImport ref="licenseRef" />
        <BindFree ref="bindFreeRef" />
        <BindXpack ref="bindXpackRef" />
        <OpDialog ref="opRef" @search="search" />
        <OpDialog ref="opRef2" @search="search" @submit="submitUnbind">
            <template #content>
                <el-form class="mt-4 mb-1" ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="forceUnbind" :label="$t('license.forceUnbind')" />
                        <span class="input-help">
                            {{ $t('license.forceUnbindHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { deleteLicense, searchLicense, syncLicense, unbindLicense } from '@/api/modules/setting';
import LicenseImport from '@/components/license-import/index.vue';
import BindFree from '@/views/setting/license/bind/free.vue';
import BindXpack from '@/views/setting/license/bind/xpack.vue';
import { dateFormat } from '@/utils/util';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import { initFavicon } from '@/utils/xpack';

const globalStore = GlobalStore();
const loading = ref();
const licenseRef = ref();
const bindFreeRef = ref();
const bindXpackRef = ref();
const opRef = ref();
const opRef2 = ref();
const forceUnbind = ref();
const unbindRow = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'license-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    type: '',
    name: '',
});

const loadMsg = (row: any) => {
    if (row.status === 'Exceptional') {
        return i18n.global.t('license.exceptionalHelper') + (row.message || '-');
    }
    if (row.status === 'Lost') {
        return i18n.global.t('license.lostHelper') + (row.message || '-');
    }
    return '';
};

const onSync = async (row: any) => {
    loading.value = true;
    await syncLicense(row.id)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            window.location.reload();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onBindFree = async (row: any) => {
    bindFreeRef.value.acceptParams({
        licenseID: row.id,
        licenseName: row.licenseName,
        freeNodes: row.freeNodes || [],
    });
};

const onUnbind = async (row: any) => {
    if (row.freeNodes && (row.freeNodes.length > 1 || !row.freeNodes[0].isXpack)) {
        MsgError(i18n.global.t('license.licenseUnbindHelper'));
        return;
    }
    unbindRow.value = row;
    opRef2.value.acceptParams({
        title: i18n.global.t('commons.button.unbind'),
        names: [row.licenseName],
        msg: i18n.global.t('license.unbindHelper'),
        api: null,
        params: null,
    });
};
const submitUnbind = async () => {
    loading.value = true;
    await unbindLicense(unbindRow.value.id, forceUnbind.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            if (unbindRow.value.freeCount !== 0) {
                globalStore.isMasterProductPro = false;
                initFavicon();
                window.location.reload();
                return;
            }
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const search = async () => {
    loading.value = true;
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await searchLicense(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            for (const item of data.value) {
                item.productName = 'product-1panel-pro';
                item.expiresAt = item.productPro === '0' ? '' : timestampToDate(Number(item.productPro));
            }
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const timestampToDate = (timestamp: number) => {
    const date = new Date(timestamp * 1000);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${y}-${m}-${d} ${h}:${minute}:${second}`;
};

const toUpload = () => {
    licenseRef.value.acceptParams();
};

const loadVersion = (row: any) => {
    if (row.trial === 'yes') {
        return i18n.global.t('license.trial');
    }
    if (row.productPro && row.productPro !== '0') {
        return i18n.global.t('license.subscription') + ' [ ' + row.expiresAt + ' ] ';
    }
    if (row.versionConstraint && row.versionConstraint !== 'all') {
        return i18n.global.t('license.versionConstraint', ['v' + row.versionConstraint.replace('.x', '')]);
    }
    return i18n.global.t('license.perpetual');
};

const buttons = [
    {
        label: i18n.global.t('commons.button.bind'),
        disabled: (row: any) => {
            return row.status !== 'Free';
        },
        click: (row: any) => {
            bindXpackRef.value.acceptParams({ licenseID: row.id, licenseName: row.licenseName });
        },
    },
    {
        label: i18n.global.t('commons.button.unbind'),
        disabled: (row: any) => {
            return row.status === 'Free';
        },
        click: (row: any) => {
            onUnbind(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: any) => {
            licenseRef.value.acceptParams({ oldLicense: row.licenseName });
        },
    },
    {
        label: i18n.global.t('commons.button.sync'),
        disabled: (row: any) => {
            return row.status === 'Free';
        },
        click: (row: any) => {
            onSync(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: any) => {
            opRef.value.acceptParams({
                title: i18n.global.t('commons.button.delete'),
                names: [row.licenseName],
                msg: i18n.global.t('commons.msg.operatorHelper', [
                    i18n.global.t('setting.license'),
                    i18n.global.t('commons.button.delete'),
                ]),
                api: deleteLicense,
                params: { id: row.id },
            });
        },
    },
];

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
.bind-button {
    margin-top: -2px;
}
</style>
