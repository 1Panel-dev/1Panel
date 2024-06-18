<template>
    <div v-if="persistenceShow">
        <el-row :gutter="20" style="margin-top: 5px" class="row-box">
            <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span>AOF {{ $t('database.persistence') }}</span>
                        </div>
                    </template>
                    <el-form :model="form" ref="formRef" :rules="rules" label-width="120px">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-form>
                                <el-form-item label="appendonly" prop="appendonly">
                                    <el-switch
                                        active-value="yes"
                                        inactive-value="no"
                                        v-model="form.appendonly"
                                    ></el-switch>
                                </el-form-item>
                                <el-form-item label="appendfsync" prop="appendfsync">
                                    <el-radio-group style="width: 100%" v-model="form.appendfsync">
                                        <el-radio value="always">always</el-radio>
                                        <el-radio value="everysec">everysec</el-radio>
                                        <el-radio value="no">no</el-radio>
                                    </el-radio-group>
                                </el-form-item>
                                <el-form-item>
                                    <el-button type="primary" @click="onSave(formRef, 'aof')">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </el-form-item>
                            </el-form>
                        </el-row>
                    </el-form>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span>RDB {{ $t('database.persistence') }}</span>
                        </div>
                    </template>
                    <table style="width: 100%" class="tab-table">
                        <tr v-for="(row, index) in form.saves" :key="index">
                            <td width="32%">
                                <el-input type="number" v-model="row.second"></el-input>
                            </td>
                            <td width="80px">
                                {{ $t('database.rdbHelper1') }}
                            </td>
                            <td width="32%">
                                <el-input type="number" v-model="row.count"></el-input>
                            </td>
                            <td width="10%">
                                {{ $t('database.rdbHelper2') }}
                            </td>
                            <td>
                                <el-button link type="primary" style="font-size: 10px" @click="handleDelete(index)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </td>
                        </tr>
                        <tr>
                            <td align="left">
                                <el-button @click="handleAdd()">{{ $t('commons.button.add') }}</el-button>
                            </td>
                        </tr>
                    </table>
                    <div>
                        <span style="margin-left: 2px; margin-top: 5px">{{ $t('database.rdbHelper3') }}</span>
                    </div>
                    <el-button type="primary" @click="onSave(undefined, 'rbd')" style="margin-top: 10px">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-card>
            </el-col>
        </el-row>
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" @search="search" :data="data">
                <template #toolbar>
                    <el-button type="primary" @click="onBackup">{{ $t('commons.button.backup') }}</el-button>
                    <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="fileName" />
                <el-table-column :label="$t('database.source')" prop="backupType">
                    <template #default="{ row }">
                        <span v-if="row.source">
                            {{ $t('setting.' + row.source) }}
                        </span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('file.dir')" show-overflow-tooltip prop="fileDir" />
                <el-table-column :label="$t('commons.table.createdAt')" :formatter="dateFormat" prop="createdAt" />
                <fu-table-operations
                    width="300px"
                    :buttons="buttons"
                    :ellipsis="10"
                    :label="$t('commons.table.operate')"
                    fix
                />
            </ComplexTable>
        </el-card>

        <OpDialog ref="opRef" @search="search" />
        <ConfirmDialog ref="confirmDialogRef" @confirm="onRecover"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { Database } from '@/api/interface/database';
import { redisPersistenceConf, updateRedisPersistenceConf } from '@/api/modules/database';
import { deleteBackupRecord, handleBackup, handleRecover, searchBackupRecords } from '@/api/modules/setting';
import { Rules } from '@/global/form-rules';
import { dateFormat } from '@/utils/util';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { MsgInfo, MsgSuccess } from '@/utils/message';
import { Backup } from '@/api/interface/backup';

interface saveStruct {
    second: number;
    count: number;
}
const form = reactive({
    appendonly: '',
    appendfsync: 'no',
    saves: [] as Array<saveStruct>,
});
const rules = reactive({
    appendonly: [Rules.requiredSelect],
    appendfsync: [Rules.requiredSelect],
});
const formRef = ref<FormInstance>();
const database = ref();
const opRef = ref();

interface DialogProps {
    database: string;
    status: string;
}
const persistenceShow = ref(false);
const acceptParams = (prop: DialogProps): void => {
    persistenceShow.value = true;
    database.value = prop.database;
    if (prop.status === 'Running') {
        loadform();
        search();
    }
};
const emit = defineEmits(['loading']);

const data = ref();
const selects = ref<any>([]);
const currentRow = ref();
const confirmDialogRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'redis-backup-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const handleAdd = () => {
    let item = {
        second: 0,
        count: 0,
    };
    form.saves.push(item);
};
const handleDelete = (index: number) => {
    form.saves.splice(index, 1);
};

const search = async () => {
    let params = {
        type: 'redis',
        name: database.value,
        detailName: '',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    const res = await searchBackupRecords(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};
const onBackup = async () => {
    emit('loading', true);
    await handleBackup({ name: database.value, detailName: '', type: 'redis', secret: '' })
        .then(() => {
            emit('loading', false);
            search();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            emit('loading', false);
        });
};
const onRecover = async () => {
    let param = {
        source: currentRow.value.source,
        type: 'redis',
        name: database.value,
        detailName: '',
        file: currentRow.value.fileDir + '/' + currentRow.value.fileName,
        secret: '',
    };
    emit('loading', true);
    await handleRecover(param)
        .then(() => {
            emit('loading', false);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            emit('loading', false);
        });
};

const onBatchDelete = async (row: Backup.RecordInfo | null) => {
    let ids: Array<number> = [];
    let names: Array<string> = [];
    if (row) {
        ids.push(row.id);
        names.push(row.fileName);
    } else {
        selects.value.forEach((item: Backup.RecordInfo) => {
            ids.push(item.id);
            names.push(item.fileName);
        });
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('commons.button.backup'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteBackupRecord,
        params: { ids: ids },
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.recover'),
        click: (row: Backup.RecordInfo) => {
            currentRow.value = row;
            let params = {
                header: i18n.global.t('commons.button.recover'),
                operationInfo: i18n.global.t('database.recoverHelper', [row.fileName]),
                submitInputInfo: i18n.global.t('database.submitIt'),
            };
            confirmDialogRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Backup.RecordInfo) => {
            onBatchDelete(row);
        },
    },
];

const onSave = async (formEl: FormInstance | undefined, type: string) => {
    let param = {} as Database.RedisConfPersistenceUpdate;
    param.database = database.value;
    if (type == 'aof') {
        if (!formEl) return;
        formEl.validate(async (valid) => {
            if (!valid) return;
            param.type = type;
            param.appendfsync = form.appendfsync;
            param.appendonly = form.appendonly;
            emit('loading', true);
            await updateRedisPersistenceConf(param)
                .then(() => {
                    emit('loading', false);
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    emit('loading', false);
                });
        });
        return;
    }
    let itemSaves = [] as Array<string>;
    for (const item of form.saves) {
        if (item.count < 0 || item.count > 100000 || item.second < 0 || item.second > 100000) {
            MsgInfo(i18n.global.t('database.rdbInfo'));
            return;
        }
        itemSaves.push(item.second + ' ' + item.count);
    }
    param.type = type;
    param.save = itemSaves.join(',');
    emit('loading', true);
    await updateRedisPersistenceConf(param)
        .then(() => {
            emit('loading', false);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            emit('loading', false);
        });
};

const loadform = async () => {
    form.saves = [];
    const res = await redisPersistenceConf(database.value);
    form.appendonly = res.data?.appendonly;
    form.appendfsync = res.data?.appendfsync;
    let itemSaves = res.data?.save.split(' ');
    for (let i = 0; i < itemSaves.length; i++) {
        if (i % 2 === 1) {
            form.saves.push({ second: Number(itemSaves[i - 1]), count: Number(itemSaves[i]) });
        }
    }
};

defineExpose({
    acceptParams,
});
</script>
