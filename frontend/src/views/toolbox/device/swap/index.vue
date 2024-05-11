<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
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
                        <el-table-column :label="$t('file.path')" min-width="120" prop="path">
                            <template #default="{ row }">
                                <span>{{ row.path }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('file.size')" min-width="150">
                            <template #default="{ row }">
                                <el-input placeholder="1024" type="number" v-model.number="row.size">
                                    <template #append>
                                        <el-select v-model="row.sizeUnit" style="width: 85px">
                                            <el-option label="KB" value="KB" />
                                            <el-option label="MB" value="MB" />
                                            <el-option label="GB" value="GB" />
                                        </el-select>
                                    </template>
                                </el-input>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('home.used')" min-width="70" prop="used">
                            <template #default="{ row }">
                                <span v-if="row.used !== '-'">{{ computeSize(row.used * 1024) }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column min-width="70">
                            <template #default="scope">
                                <el-button link type="primary" @click="onSave(scope.row)">
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                    <span class="input-help">{{ $t('toolbox.swap.swapOff') }}</span>
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
import { MsgError, MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateDeviceSwap, getDeviceBase } from '@/api/modules/toolbox';
import { computeSize, splitSize } from '@/utils/util';
import { loadBaseDir } from '@/api/modules/setting';

const form = reactive({
    swapMemoryTotal: '',
    swapMemoryAvailable: '',
    swapMemoryUsed: '',
    maxSize: 0,

    swapDetails: [],
});

const drawerVisible = ref();
const loading = ref();

const acceptParams = (): void => {
    search();
    drawerVisible.value = true;
};

const search = async () => {
    loading.value = true;
    const res = await getDeviceBase();
    form.swapMemoryTotal = computeSize(res.data.swapMemoryTotal);
    form.swapMemoryUsed = computeSize(res.data.swapMemoryUsed);
    form.swapMemoryAvailable = computeSize(res.data.swapMemoryAvailable);
    form.swapDetails = res.data.swapDetails || [];
    form.maxSize = res.data.maxSize;

    await loadBaseDir()
        .then((res) => {
            loading.value = false;
            loadData(res.data.substring(0, res.data.lastIndexOf('/1panel')));
        })
        .catch(() => {
            loading.value = false;
            loadData('');
        });
};

const loadData = (path: string) => {
    let isExist = false;
    for (const item of form.swapDetails) {
        if (item.path === path + '/.1panel_swap') {
            isExist = true;
        }
        let itemSize = splitSize(item.size * 1024);
        item.size = itemSize.size;
        item.sizeUnit = itemSize.unit;
    }
    if (!isExist) {
        form.swapDetails.push({
            path: path + '/.1panel_swap',
            size: 0,
            used: 0,
            isNew: true,
            sizeUnit: 'MB',
        });
    }
};

const onSave = async (row) => {
    if (row.size === '') {
        MsgError(i18n.global.t('commons.msg.confirmNoNull', ['Swap ' + i18n.global.t('file.size')]));
        return;
    }
    const itemSize = loadItemSize(row);
    if (itemSize < 40 && row.size !== 0) {
        MsgError(i18n.global.t('toolbox.swap.swapMin'));
        return;
    }
    if (itemSize * 1024 > form.maxSize) {
        MsgError(i18n.global.t('toolbox.swap.swapMax', [computeSize(form.maxSize)]));
        return;
    }
    ElMessageBox.confirm(
        i18n.global.t('toolbox.swap.saveSwap', [row.path, row.size + ' ' + row.sizeUnit]),
        i18n.global.t('commons.button.save'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        let params = {
            path: row.path,
            size: itemSize,
            used: '0',

            isNew: row.isNew,
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

const loadItemSize = (row: any) => {
    switch (row.sizeUnit) {
        case 'KB':
            return row.size;
        case 'MB':
            return row.size * 1024;
        case 'GB':
            return row.size * 1024 * 1024;
    }
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
