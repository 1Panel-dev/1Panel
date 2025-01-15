<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.license')">
            <template #leftToolBar>
                <el-button type="primary" @click="toUpload()">
                    {{ $t('commons.button.add') }}
                </el-button>
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
                    <el-table-column :label="$t('license.expiresAt')" prop="expiresAt" show-overflow-tooltip>
                        <template #default="{ row }">
                            {{ row.expiresAt || '-' }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" show-overflow-tooltip>
                        <template #default="{ row }">
                            <div v-if="row.status">
                                <Status :status="row.status" :msg="loadMsg(row.status)"></Status>
                            </div>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('setting.bindNode')">
                        <template #default="{ row }">
                            <span v-if="row.freeCount !== 0 && (row.status === 'Free' || row.status === 'Exceptional')">
                                -
                            </span>
                            <div v-else>
                                <span v-if="row.freeCount === 0">{{ row.bindNode || '-' }}</span>
                                <div v-else>
                                    <el-popover
                                        placement="bottom"
                                        :width="120"
                                        trigger="hover"
                                        v-if="row.freeNodes && row.freeNodes.length != 0"
                                    >
                                        <div v-for="(item, index) of row.freeNodes" :key="index">
                                            <el-tag>{{ item.name }}</el-tag>
                                        </div>
                                        <template #reference>
                                            <el-button link type="primary">
                                                ({{ row.bindCount }} / {{ row.freeCount }})
                                            </el-button>
                                        </template>
                                    </el-popover>
                                    <span v-else link type="primary">({{ row.bindCount }} / {{ row.freeCount }})</span>
                                </div>
                            </div>
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
        <LicenseBind ref="bindRef" />
        <LicenseDelete ref="delRef" @search="search" />
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { SearchLicense, syncLicense, unbindLicense } from '@/api/modules/setting';
import LicenseImport from '@/components/license-import/index.vue';
import LicenseDelete from '@/views/setting/license/delete/index.vue';
import LicenseBind from '@/views/setting/license/bind/index.vue';
import { dateFormat } from '@/utils/util';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import { initFavicon } from '@/utils/xpack';

const globalStore = GlobalStore();
const loading = ref();
const licenseRef = ref();
const delRef = ref();
const bindRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'license-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    type: '',
    name: '',
});

const loadMsg = (status: string) => {
    if (status === 'Exceptional') {
        return i18n.global.t('license.exceptionalHelper');
    }
    if (status === 'Lost') {
        return i18n.global.t('license.lostHelper');
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

const onUnbind = async (row: any) => {
    ElMessageBox.confirm(i18n.global.t('license.unbindHelper'), i18n.global.t('license.unbind'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await unbindLicense(row.id)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                if (row.freeCount !== 0) {
                    globalStore.isProductPro = false;
                    initFavicon();
                    window.location.reload();
                    return;
                }
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const search = async () => {
    loading.value = true;
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await SearchLicense(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            for (const item of data.value) {
                item.productName = 'product-1panel-pro';
                item.expiresAt =
                    item.productPro === '0'
                        ? i18n.global.t('license.indefinitePeriod')
                        : timestampToDate(Number(item.productPro));
            }
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

const buttons = [
    {
        label: i18n.global.t('commons.button.bind'),
        disabled: (row: any) => {
            return row.status !== 'Free';
        },
        click: (row: any) => {
            bindRef.value.acceptParams({ licenseID: row.id, licenseName: row.licenseName });
        },
    },
    {
        label: i18n.global.t('commons.button.unbind'),
        disabled: (row: any) => {
            return row.status === 'Free';
        },
        click: (row: any) => {
            if (row.freeCount != 0) {
                if (row.freeNodes) {
                    MsgError(i18n.global.t('license.unbindMasterHelper', [i18n.global.t('commons.button.unbind')]));
                    return;
                }
                for (const item of data.value) {
                    if (item.bindNode && item.freeCount == 0) {
                        MsgError(i18n.global.t('license.unbindMasterHelper', [i18n.global.t('commons.button.unbind')]));
                        return;
                    }
                }
            }
            onUnbind(row);
        },
    },
    {
        label: i18n.global.t('commons.button.sync'),
        disabled: (row: any) => {
            return row.status.indexOf('Lost') !== -1 || row.status === 'Disable';
        },
        click: (row: any) => {
            onSync(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: any) => {
            for (const item of data.value) {
                if (item.bindNode && row.freeCount != 0) {
                    MsgError(i18n.global.t('license.unbindMasterHelper', [i18n.global.t('commons.button.delete')]));
                    return;
                }
            }
            delRef.value.acceptParams({ id: row.id, name: row.licenseName });
        },
    },
];

onMounted(() => {
    search();
});
</script>
