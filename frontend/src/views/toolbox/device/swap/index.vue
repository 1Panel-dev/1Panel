<template>
    <div v-loading="loading">
        <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader header="Swap" :back="handleClose" />
            </template>

            <el-row type="flex" justify="center" v-loading="loading">
                <el-col :span="22">
                    <el-card>
                        <el-form label-position="top" class="ml-3">
                            <el-row type="flex" justify="center" :gutter="20">
                                <el-col :xs="8" :sm="8" :md="8" :lg="8" :xl="8">
                                    <el-form-item>
                                        <template #label>
                                            <span class="status-label">
                                                {{ $t('home.disk') }} {{ $t('home.total') }}
                                            </span>
                                        </template>
                                        <span class="status-count">{{ form.memoryTotal }}</span>
                                    </el-form-item>
                                </el-col>
                                <el-col :xs="8" :sm="8" :md="8" :lg="8" :xl="8">
                                    <el-form-item>
                                        <template #label>
                                            <span class="status-label">
                                                {{ $t('home.disk') }} {{ $t('home.used') }}
                                            </span>
                                        </template>
                                        <span class="status-count">{{ form.memoryUsed }}</span>
                                    </el-form-item>
                                </el-col>
                                <el-col :xs="8" :sm="8" :md="8" :lg="8" :xl="8">
                                    <el-form-item>
                                        <template #label>
                                            <span class="status-label">
                                                {{ $t('home.disk') }} {{ $t('home.free') }}
                                            </span>
                                        </template>
                                        <span class="status-count">{{ form.memoryAvailable }}</span>
                                    </el-form-item>
                                </el-col>
                            </el-row>
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
                                <el-input v-if="row.isCreate" v-model.number="row.path"></el-input>
                                <span v-else>{{ row.path }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('file.size') + ' (MB)'" min-width="120">
                            <template #default="{ row }">
                                <el-input v-if="row.isCreate || row.isEdit" v-model.number="row.size"></el-input>
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
    memoryTotal: 0,
    memoryAvailable: 0,
    memoryUsed: 0,
    swapMemoryTotal: 0,
    swapMemoryAvailable: 0,
    swapMemoryUsed: 0,

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
const handleDelete = async (row: any) => {
    let params = {
        operate: 'delete',
        path: row.path,
        size: 0,
        used: '0',
        withRemove: true,
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
};

const search = async () => {
    isCreate.value = false;
    const res = await getDeviceBase();
    form.memoryTotal = computeSize(res.data.memoryTotal);
    form.memoryUsed = computeSize(res.data.memoryUsed);
    form.memoryAvailable = computeSize(res.data.memoryAvailable);
    form.swapMemoryTotal = computeSize(res.data.swapMemoryTotal);
    form.swapMemoryUsed = computeSize(res.data.swapMemoryUsed);
    form.swapMemoryAvailable = computeSize(res.data.swapMemoryAvailable);
    form.swapDetails = res.data.swapDetails;
    for (const item of form.swapDetails) {
        item.isCreate = false;
        item.isEdit = false;
        item.size = Number((item.size / 1024).toFixed(2));
    }
};

const onSave = async (row) => {
    let params = {
        operate: row.isCreate ? 'create' : 'update',
        path: row.path,
        size: row.size * 1024,
        used: '0',
        withRemove: false,
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
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
