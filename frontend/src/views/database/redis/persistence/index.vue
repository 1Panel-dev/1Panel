<template>
    <div v-if="persistenceShow">
        <el-row :gutter="20" style="margin-top: 5px" class="row-box">
            <el-col :span="12">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span>AOF {{ $t('database.persistence') }}</span>
                        </div>
                    </template>
                    <el-form :model="form" ref="formRef" :rules="rules" label-width="120px">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form>
                                    <el-form-item label="appendonly" prop="appendonly">
                                        <el-switch
                                            active-value="yes"
                                            inactive-value="no"
                                            v-model="form.appendonly"
                                        ></el-switch>
                                    </el-form-item>
                                    <el-form-item label="appendfsync" prop="appendfsync">
                                        <el-radio-group v-model="form.appendfsync">
                                            <el-radio label="always">always</el-radio>
                                            <el-radio label="everysec">everysec</el-radio>
                                            <el-radio label="no">no</el-radio>
                                        </el-radio-group>
                                    </el-form-item>
                                </el-form>
                                <el-button
                                    type="primary"
                                    @click="onSave(formRef, 'aof')"
                                    style="bottom: 10px; width: 90px"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </el-col>
                        </el-row>
                    </el-form>
                </el-card>
            </el-col>
            <el-col :span="12">
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
                            <td width="55px">
                                <span>{{ $t('database.rdbHelper1') }}</span>
                            </td>
                            <td width="32%">
                                <el-input type="number" v-model="row.count"></el-input>
                            </td>
                            <td width="12%">
                                <span>{{ $t('database.rdbHelper2') }}</span>
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
                    <el-button type="primary" @click="onSave(undefined, 'rbd')" style="margin-top: 10px; width: 90px">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-card>
            </el-col>
        </el-row>
        <el-card style="margin-top: 20px">
            <ComplexTable
                :pagination-config="paginationConfig"
                v-model:selects="selects"
                @search="loadBackupRecords"
                :data="data"
            >
                <template #toolbar>
                    <el-button type="primary" @click="onBackup">{{ $t('setting.backup') }}</el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="fileName" />
                <el-table-column :label="$t('file.dir')" show-overflow-tooltip prop="fileDir" />
                <el-table-column :label="$t('file.size')" prop="size">
                    <template #default="{ row }">
                        {{ computeSize(row.size) }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" />
                <fu-table-operations
                    width="300px"
                    :buttons="buttons"
                    :ellipsis="10"
                    :label="$t('commons.table.operate')"
                    fix
                />
            </ComplexTable>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { Database } from '@/api/interface/database';
import {
    backupRedis,
    deleteBackupRedis,
    recoverRedis,
    redisBackupRedisRecords,
    RedisPersistenceConf,
    updateRedisPersistenceConf,
} from '@/api/modules/database';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElMessage, FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { useDeleteData } from '@/hooks/use-delete-data';
import { computeSize } from '@/utils/util';

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

const persistenceShow = ref(false);
const acceptParams = (): void => {
    persistenceShow.value = true;
    loadform();
    loadBackupRecords();
};
const onClose = (): void => {
    persistenceShow.value = false;
};

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
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

const loadBackupRecords = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    const res = await redisBackupRedisRecords(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};
const onBackup = async () => {
    await backupRedis();
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    loadBackupRecords();
};
const onRecover = async (row: Database.RedisBackupRecord) => {
    let param = {
        fileName: row.fileName,
        fileDir: row.fileDir,
    };
    await recoverRedis(param);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const onBatchDelete = async (row: Database.RedisBackupRecord | null) => {
    let names: Array<string> = [];
    let fileDir: string = '';
    if (row) {
        fileDir = row.fileDir;
        names.push(row.fileName);
    } else {
        selects.value.forEach((item: Database.RedisBackupRecord) => {
            fileDir = item.fileDir;
            names.push(item.fileName);
        });
    }
    await useDeleteData(deleteBackupRedis, { fileDir: fileDir, names: names }, 'commons.msg.delete', true);
    loadBackupRecords();
};
const buttons = [
    {
        label: i18n.global.t('commons.button.recover'),
        click: (row: Database.RedisBackupRecord) => {
            onRecover(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Database.RedisBackupRecord) => {
            onBatchDelete(row);
        },
    },
];

const onSave = async (formEl: FormInstance | undefined, type: string) => {
    let param = {} as Database.RedisConfPersistenceUpdate;
    if (type == 'aof') {
        if (!formEl) return;
        formEl.validate(async (valid) => {
            if (!valid) return;
            param.type = type;
            param.appendfsync = form.appendfsync;
            param.appendonly = form.appendonly;
            await updateRedisPersistenceConf(param);
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            return;
        });
        return;
    }
    let itemSaves = [] as Array<string>;
    for (const item of form.saves) {
        if (item.count === 0 || item.second === 0) {
            ElMessage.info(i18n.global.t('database.rdbInfo'));
            return;
        }
        itemSaves.push(item.second + '', item.count + '');
    }
    param.type = type;
    param.save = itemSaves.join(' ');
    await updateRedisPersistenceConf(param);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const loadform = async () => {
    form.saves = [];
    const res = await RedisPersistenceConf();
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
    onClose,
});
</script>
