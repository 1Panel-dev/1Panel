<template>
    <div>
        <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader :header="$t('menu.home')" :back="handleClose" />
            </template>
            <div v-loading="loading">
                <el-form label-position="top" class="ml-5">
                    <el-row type="flex" justify="center" :gutter="20">
                        <el-col :xs="12" :sm="12" :md="12" :lg="12" :xl="12">
                            <el-form-item>
                                <template #label>
                                    <span class="status-label">{{ $t('ssh.successful') }}</span>
                                </template>
                                <span class="status-count">{{ successfulTotalCount }}</span>
                            </el-form-item>
                        </el-col>
                        <el-col :xs="12" :sm="12" :md="12" :lg="12" :xl="12">
                            <el-form-item>
                                <template #label>
                                    <span class="status-label">{{ $t('ssh.failed') }}</span>
                                </template>
                                <span class="status-count">{{ failedTotalCount }}</span>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>

                <el-button type="primary" @click="onChangeStatus('accept', null)" :disabled="selects.length === 0">
                    {{ $t('firewall.allow') }}
                </el-button>
                <el-button type="primary" @click="onChangeStatus('drop', null)" :disabled="selects.length === 0">
                    {{ $t('firewall.deny') }}
                </el-button>
                <ComplexTable v-model:selects="selects" class="mt-5" :data="data" @header-click="changeSort">
                    <el-table-column type="selection" fix :selectable="selectable" />
                    <el-table-column :label="$t('logs.loginIP')" prop="address" min-width="40" />
                    <el-table-column :label="$t('ssh.belong')" prop="area" min-width="40" />
                    <el-table-column prop="successfulCount" min-width="20">
                        <template #header>
                            {{ $t('ssh.successful') }}
                            <el-icon style="cursor: pointer" @click="search('Success')"><CaretBottom /></el-icon>
                        </template>
                        <template #default="{ row }">
                            <el-button type="primary" link>{{ row.successfulCount }}</el-button>
                        </template>
                    </el-table-column>
                    <el-table-column prop="failedCount" min-width="20">
                        <template #header>
                            {{ $t('ssh.failed') }}
                            <el-icon style="cursor: pointer" @click="search('Failed')"><CaretBottom /></el-icon>
                        </template>
                        <template #default="{ row }">
                            <el-button type="danger" link>{{ row.failedCount }}</el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :min-width="30" :label="$t('commons.table.status')" prop="strategy">
                        <template #default="{ row }">
                            <el-button
                                v-if="row.status === 'accept'"
                                :disabled="!selectable(row)"
                                @click="onChangeStatus('drop', row)"
                                link
                                type="success"
                            >
                                {{ $t('commons.status.accept') }}
                            </el-button>
                            <el-button
                                v-else
                                link
                                :disabled="!selectable(row)"
                                type="danger"
                                @click="onChangeStatus('accept', row)"
                            >
                                {{ $t('commons.status.deny') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-drawer>

        <el-dialog
            v-model="dialogVisible"
            :title="$t('firewall.' + (operation === 'drop' ? 'deny' : 'allow'))"
            width="30%"
            :close-on-click-modal="false"
        >
            <el-row>
                <el-col :span="20" :offset="2">
                    <el-alert :title="msg" show-icon type="error" :closable="false"></el-alert>
                    <div class="resource">
                        <table>
                            <tr v-for="(row, index) in operationList" :key="index">
                                <td>
                                    <span>{{ row }}</span>
                                </td>
                            </tr>
                        </table>
                    </div>
                </el-col>
            </el-row>

            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dialogVisible = false" :disabled="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="submitOperation" v-loading="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { loadAnalysis, operateIPRule } from '@/api/modules/host';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { Host } from '@/api/interface/host';

const drawerVisible = ref();
const loading = ref();
const data = ref();
const successfulTotalCount = ref();
const failedTotalCount = ref();
const selects = ref<any>([]);

const dialogVisible = ref();
const msg = ref();
const operation = ref();
const operationList = ref();

const acceptParams = (): void => {
    search('Failed');
    drawerVisible.value = true;
};

const search = async (status: string) => {
    loading.value = true;
    loadAnalysis(status)
        .then((res) => {
            loading.value = false;
            data.value = res.data || [];
            successfulTotalCount.value = 0;
            failedTotalCount.value = 0;
            for (const item of data.value) {
                successfulTotalCount.value += item.successfulCount;
                failedTotalCount.value += item.failedCount;
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

function selectable(row: any): boolean {
    return row.address !== '127.0.0.1' && row.address !== '::1';
}

const changeSort = (column: any) => {
    switch (column.property) {
        case 'successfulCount':
            search('Success');
            return;
        case 'failedCount':
            search('Failed');
            return;
    }
};

const onChangeStatus = async (status: string, row: Host.logAnalysis | null) => {
    operationList.value = [];
    if (row) {
        if (row.status !== status) {
            operationList.value.push(row.address);
        }
    } else {
        for (const item of selects.value) {
            if (item.status !== status) {
                operationList.value.push(item.address);
            }
        }
    }
    if (operationList.value.length === 0) {
        MsgError(
            i18n.global.t('ssh.noAddrWarning', [i18n.global.t('firewall.' + (status === 'drop' ? 'deny' : 'allow'))]),
        );
        return;
    }
    operation.value = status;
    msg.value = status === 'drop' ? i18n.global.t('ssh.denyHelper') : i18n.global.t('ssh.acceptHelper');
    dialogVisible.value = true;
};

const submitOperation = async () => {
    const pros = [];
    for (const item of operationList.value) {
        pros.push(
            operateIPRule({
                operation: operation.value === 'drop' ? 'add' : 'remove',
                address: item,
                strategy: 'drop',
                description: '',
            }),
        );
    }
    loading.value = true;
    Promise.all(pros)
        .then(() => {
            loading.value = false;
            dialogVisible.value = false;
            search('Failed');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
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

<style scoped lang="scss">
.resource {
    margin-top: 10px;
    max-height: 400px;
    overflow: auto;
}
</style>
