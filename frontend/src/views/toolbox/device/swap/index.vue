<template>
    <div v-loading="loading">
        <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader header="Swap" :back="handleClose" />
            </template>

            <el-row type="flex" justify="center" v-loading="loading">
                <el-col :span="22">
                    <el-alert class="common-prompt" :closable="false" type="warning">
                        <template #default>
                            <ul style="margin-left: -20px">
                                <li>{{ $t('toolbox.swap.swapHelper1') }}</li>
                                <li>{{ $t('toolbox.swap.swapHelper2') }}</li>
                                <li>{{ $t('toolbox.swap.swapHelper3') }}</li>
                                <li>{{ $t('toolbox.swap.swapHelper4') }}</li>
                            </ul>
                        </template>
                    </el-alert>
                    <el-card>
                        <el-form label-position="top" class="ml-3">
                            <el-row type="flex" justify="center" :gutter="20">
                                <el-col :xs="8" :sm="8" :md="8" :lg="8" :xl="8">
                                    <el-form-item>
                                        <template #label>
                                            <span class="status-label">Swap {{ $t('home.total') }}</span>
                                        </template>
                                        <span class="status-count">{{ form.swapMemoryTotal }}</span>
                                    </el-form-item>
                                </el-col>
                                <el-col :xs="8" :sm="8" :md="8" :lg="8" :xl="8">
                                    <el-form-item>
                                        <template #label>
                                            <span class="status-label">Swap {{ $t('home.used') }}</span>
                                        </template>
                                        <span class="status-count">{{ form.swapMemoryUsed }}</span>
                                    </el-form-item>
                                </el-col>
                                <el-col :xs="8" :sm="8" :md="8" :lg="8" :xl="8">
                                    <el-form-item>
                                        <template #label>
                                            <span class="status-label">Swap {{ $t('home.free') }}</span>
                                        </template>
                                        <span class="status-count">{{ form.swapMemoryAvailable }}</span>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                        </el-form>
                    </el-card>

                    <el-table :data="form.swapDetails" class="mt-5">
                        <el-table-column :label="$t('file.path')" min-width="150" prop="path">
                            <template #default="{ row }">
                                <el-input
                                    placeholder="/var/swap"
                                    v-if="row.isCreate"
                                    v-model.number="row.path"
                                ></el-input>
                                <span v-else>{{ row.path }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('file.size') + ' (MB)'" min-width="120">
                            <template #default="{ row }">
                                <el-input
                                    placeholder="1024"
                                    v-if="row.isCreate || row.isEdit"
                                    v-model.number="row.size"
                                ></el-input>
                                <span v-else>{{ row.size }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('home.used')" min-width="70" prop="used">
                            <template #default="{ row }">
                                <span v-if="row.used !== '-'">{{ computeSize(row.used * 1024) }}</span>
                                <span v-else>-</span>
                            </template>
                        </el-table-column>
                        <el-table-column min-width="70">
                            <template #default="scope">
                                <div v-if="!scope.row.isCreate && !scope.row.isEdit">
                                    <el-button link type="primary" @click="scope.row.isEdit = true">
                                        {{ $t('commons.button.edit') }}
                                    </el-button>
                                    <el-button link type="primary" @click="handleDelete(scope.row)">
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </div>
                                <div v-if="scope.row.isCreate || scope.row.isEdit">
                                    <el-button link type="primary" @click="onSave(scope.row)">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                    <el-button link type="primary" @click="search()">
                                        {{ $t('commons.button.cancel') }}
                                    </el-button>
                                </div>
                            </template>
                        </el-table-column>
                    </el-table>
                    <el-button class="ml-3 mt-2" @click="handleAdd()">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess, MsgInfo } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateDeviceSwap, getDeviceBase } from '@/api/modules/toolbox';
import { computeSize } from '@/utils/util';

const isCreate = ref();

const form = reactive({
    swapItem: '',
    swapMemoryTotal: '',
    swapMemoryAvailable: '',
    swapMemoryUsed: '',

    swapDetails: [],
});

const drawerVisible = ref();
const loading = ref();

const acceptParams = (): void => {
    search();
    drawerVisible.value = true;
};
const handleAdd = () => {
    if (isCreate.value) {
        MsgInfo(i18n.global.t('toolbox.swap.saveHelper'));
        return;
    }
    let item = {
        path: '',
        size: '',
        used: '-',
        isCreate: true,
        isEdit: false,
    };
    isCreate.value = true;
    form.swapDetails.push(item);
};

const search = async () => {
    isCreate.value = false;
    const res = await getDeviceBase();
    form.swapMemoryTotal = computeSize(res.data.swapMemoryTotal);
    form.swapMemoryUsed = computeSize(res.data.swapMemoryUsed);
    form.swapMemoryAvailable = computeSize(res.data.swapMemoryAvailable);
    form.swapDetails = res.data.swapDetails || [];
    for (const item of form.swapDetails) {
        item.isCreate = false;
        item.isEdit = false;
        item.size = Number((item.size / 1024).toFixed(2));
    }
};

const handleDelete = async (row: any) => {
    ElMessageBox.confirm(
        i18n.global.t('toolbox.swap.swapDeleteHelper', [row.path]),
        i18n.global.t('commons.button.delete'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        let params = {
            operate: 'delete',
            path: row.path,
            size: 0,
            used: '0',
        };
        loading.value = true;
        await updateDeviceSwap(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onSave = async (row) => {
    let itemSize = 0;
    for (const item of form.swapDetails) {
        itemSize += item.size;
    }
    ElMessageBox.confirm(
        i18n.global.t('toolbox.swap.saveSwap', [itemSize + ' M']),
        i18n.global.t('commons.button.save'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        let params = {
            operate: row.isCreate ? 'create' : 'update',
            path: row.path,
            size: row.size * 1024,
            used: '0',
        };
        loading.value = true;
        await updateDeviceSwap(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
