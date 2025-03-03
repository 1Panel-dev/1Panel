<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.importSnapshot')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('setting.backupAccount')" prop="from">
                <el-select v-model="form.backupAccountID" @change="loadFiles" clearable>
                    <el-option v-for="item in backupOptions" :key="item.label" :value="item.id" :label="item.label" />
                </el-select>
                <div v-if="form.backupAccountID === 0">
                    <span class="import-help">{{ $t('setting.importHelper') }}</span>
                    <span @click="toFolder()" class="import-link-help">{{ backupPath }}</span>
                </div>
            </el-form-item>
            <el-form-item :label="$t('commons.table.name')" prop="names">
                <el-select v-model="form.names" multiple clearable>
                    <el-option
                        :disabled="checkDisable(item)"
                        v-for="item in fileNames"
                        :key="item"
                        :value="item"
                        :label="item"
                    />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input type="textarea" clearable v-model="form.description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="drawerVisible = false">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" type="primary" @click="submitImport(formRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import i18n from '@/lang';
import { snapshotImport } from '@/api/modules/setting';
import { listBackupOptions, getFilesFromBackup } from '@/api/modules/backup';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import router from '@/routers';

const drawerVisible = ref(false);
const loading = ref();

const formRef = ref();
const backupOptions = ref();
const fileNames = ref();
const existNames = ref();
const backupPath = ref('');

const form = reactive({
    backupAccountID: 0,
    names: [],
    description: '',
});

const rules = reactive({
    backupAccountID: [Rules.requiredSelect],
    names: [Rules.requiredSelect],
});

interface DialogProps {
    names: Array<string>;
}

const acceptParams = (params: DialogProps): void => {
    form.backupAccountID = undefined;
    existNames.value = params.names;
    form.names = [] as Array<string>;
    loadBackups();
    drawerVisible.value = true;
};
const emit = defineEmits(['search']);

const handleClose = () => {
    drawerVisible.value = false;
};

const checkDisable = (val: string) => {
    for (const item of existNames.value) {
        if (val === item + '.tar.gz') {
            return true;
        }
    }
    return false;
};
const toFolder = async () => {
    router.push({ path: '/hosts/files', query: { path: backupPath.value } });
};

const submitImport = async (formEl: FormInstance | undefined) => {
    loading.value = true;
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await snapshotImport(form)
            .then(() => {
                emit('search');
                loading.value = false;
                drawerVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const loadBackups = async () => {
    loading.value = true;
    await listBackupOptions()
        .then((res) => {
            loading.value = false;
            backupOptions.value = [];
            for (const item of res.data) {
                backupOptions.value.push({
                    id: item.id,
                    label: i18n.global.t('setting.' + item.type),
                    value: item.type,
                });
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadFiles = async () => {
    form.names = [];
    const res = await getFilesFromBackup(form.backupAccountID);
    fileNames.value = res.data || [];
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.import-help {
    font-size: 12px;
    color: #8f959e;
}
.import-link-help {
    color: $primary-color;
    cursor: pointer;
}

.import-link-help:hover {
    opacity: 0.6;
}
</style>
