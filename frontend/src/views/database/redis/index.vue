<template>
    <div>
        <Submenu activeName="redis" />
        <el-dropdown size="default" split-button style="margin-top: 20px; margin-bottom: 5px">
            {{ redisName }}
            <template #dropdown>
                <el-dropdown-menu v-model="redisName">
                    <el-dropdown-item v-for="item in redisNames" :key="item" @click="onChangeName(item)">
                        {{ item }}
                    </el-dropdown-item>
                </el-dropdown-menu>
            </template>
        </el-dropdown>
        <el-button
            v-if="!isOnSetting"
            style="margin-top: 20px; margin-left: 10px"
            size="default"
            icon="Setting"
            @click="onSetting"
        >
            {{ $t('database.setting') }}
        </el-button>
        <el-button
            v-if="isOnSetting"
            style="margin-top: 20px; margin-left: 10px"
            size="default"
            icon="Back"
            @click="onBacklist"
        >
            {{ $t('commons.button.back') }}列表
        </el-button>

        <Setting ref="settingRef"></Setting>
        <el-card v-if="!isOnSetting">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" @search="search" :data="data">
                <template #toolbar>
                    <el-button type="primary" @click="onOperate">{{ $t('commons.button.create') }}</el-button>
                    <el-button type="primary" @click="onCleanAll">{{ $t('database.cleanAll') }}</el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                    <el-select v-model="currentDB" @change="search" style="margin-left: 20px">
                        <el-option
                            v-for="item in dbOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('database.key')" prop="key" />
                <el-table-column :label="$t('database.value')" prop="value" />
                <el-table-column :label="$t('database.type')" prop="type" />
                <el-table-column :label="$t('database.length')" prop="length" />
                <el-table-column :label="$t('database.expiration')" prop="expiration">
                    <template #default="{ row }">
                        <span v-if="row.expiration === -1">{{ $t('database.forever') }}</span>
                        <span v-else>{{ row.expiration }} {{ $t('database.second') }}</span>
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </el-card>

        <el-dialog v-model="redisVisiable" :destroy-on-close="true" width="30%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('database.changePassword') }}</span>
                </div>
            </template>
            <el-form>
                <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
                    <el-form-item prop="db">
                        <el-select v-model="form.db">
                            <el-option
                                v-for="item in dbOptions"
                                :key="item.label"
                                :value="item.value"
                                :label="item.label"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('database.key')" prop="key">
                        <el-input clearable v-model="form.key"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('database.value')" prop="value">
                        <el-input clearable v-model="form.value"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('database.expiration')" prop="expiration">
                        <el-input type="number" clearable v-model.number="form.expiration">
                            <template #append>{{ $t('database.second') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('database.expirationHelper') }}</span>
                    </el-form-item>
                </el-form>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="redisVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button @click="submit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import Submenu from '@/views/database/index.vue';
import Setting from '@/views/database/redis/setting/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { cleanRedisKey, deleteRedisKey, loadRedisVersions, searchRedisDBs, setRedis } from '@/api/modules/database';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';
import { Database } from '@/api/interface/database';
import { ElForm, ElMessage, ElMessageBox } from 'element-plus';
import { Rules } from '@/global/form-rules';

const selects = ref<any>([]);
const currentDB = ref(0);
const dbOptions = ref([
    { label: 'DB0', value: 0 },
    { label: 'DB1', value: 1 },
    { label: 'DB2', value: 2 },
    { label: 'DB3', value: 3 },
    { label: 'DB4', value: 4 },
    { label: 'DB5', value: 5 },
    { label: 'DB6', value: 6 },
    { label: 'DB7', value: 7 },
    { label: 'DB8', value: 8 },
    { label: 'DB9', value: 9 },
]);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const form = reactive({
    redisName: '',
    key: '',
    value: '',
    db: 0,
    expiration: 0,
});
const rules = reactive({
    db: [Rules.requiredSelect],
    key: [Rules.requiredInput],
    value: [Rules.requiredInput],
    expiration: [Rules.requiredInput, Rules.number],
});
const redisVisiable = ref(false);

const redisNames = ref();
const redisName = ref();
const isOnSetting = ref(false);
const settingRef = ref();
const onSetting = async () => {
    isOnSetting.value = true;
    let params = {
        redisName: redisName.value,
        db: currentDB.value,
    };
    settingRef.value!.acceptParams(params);
};
const onBacklist = async () => {
    isOnSetting.value = false;
    search();
    settingRef.value!.onClose();
};

const loadRunningNames = async () => {
    const res = await loadRedisVersions();
    redisNames.value = res.data;
    if (redisNames.value.length != 0) {
        redisName.value = redisNames.value[0];
        search();
    }
};
const onChangeName = async (val: string) => {
    redisName.value = val;
    search();
    if (isOnSetting.value) {
        let params = {
            redisName: redisName.value,
        };
        settingRef.value!.acceptParams(params);
    }
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        redisName: redisName.value,
        db: currentDB.value,
    };
    const res = await searchRedisDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onBatchDelete = async (row: Database.RedisData | null) => {
    let names: Array<string> = [];
    if (row) {
        names.push(row.key);
    } else {
        selects.value.forEach((item: Database.RedisData) => {
            names.push(item.key);
        });
    }
    let params = {
        redisName: redisName.value,
        db: form.db,
        names: names,
    };
    await useDeleteData(deleteRedisKey, params, 'commons.msg.delete', true);
    search();
};

const onCleanAll = async () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete') + '?', i18n.global.t('database.cleanAll'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
        draggable: true,
    }).then(async () => {
        let params = {
            redisName: redisName.value,
            db: currentDB.value,
        };
        await cleanRedisKey(params);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const onOperate = async (row: Database.RedisData | undefined) => {
    if (row) {
        form.db = currentDB.value;
        form.key = row.key;
        form.value = row.value;
        form.expiration = row.expiration === -1 ? 0 : row.expiration;
    } else {
        form.db = currentDB.value;
        form.key = '';
        form.value = '';
        form.expiration = 0;
    }
    redisVisiable.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        form.redisName = redisName.value;
        await setRedis(form);
        redisVisiable.value = false;
        currentDB.value = form.db;
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Database.RedisData) => {
            onOperate(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Database.RedisData) => {
            onBatchDelete(row);
        },
    },
];
onMounted(() => {
    loadRunningNames();
});
</script>
