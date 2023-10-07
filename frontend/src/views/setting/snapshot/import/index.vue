<template>
    <div>
        <el-drawer v-model="drawerVisible" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.importSnapshot')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules" v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.backupAccount')" prop="from">
                            <el-select style="width: 100%" v-model="form.from" @change="loadFiles" clearable>
                                <el-option
                                    v-for="item in backupOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.name')" prop="names">
                            <el-select style="width: 100%" v-model="form.names" multiple clearable>
                                <el-option v-for="item in fileNames" :key="item" :value="item" :label="item" />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input type="textarea" clearable v-model="form.description" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="drawerVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitImport(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { snapshotImport } from '@/api/modules/setting';
import { getBackupList, getFilesFromBackup } from '@/api/modules/setting';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref(false);
const loading = ref();

const formRef = ref();
const backupOptions = ref();
const fileNames = ref();

const form = reactive({
    from: '',
    names: [],
    description: '',
});

const rules = reactive({
    from: [Rules.requiredSelect],
    name: [Rules.requiredSelect],
});

const acceptParams = (): void => {
    form.from = '';
    form.names = [] as Array<string>;
    loadBackups();
    drawerVisible.value = true;
};
const emit = defineEmits(['search']);

const handleClose = () => {
    drawerVisible.value = false;
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
    await getBackupList()
        .then((res) => {
            loading.value = false;
            backupOptions.value = [];
            for (const item of res.data) {
                if (item.type !== 'LOCAL' && item.id !== 0) {
                    backupOptions.value.push({ label: i18n.global.t('setting.' + item.type), value: item.type });
                }
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadFiles = async () => {
    const res = await getFilesFromBackup(form.from);
    fileNames.value = res.data || [];
};

defineExpose({
    acceptParams,
});
</script>
