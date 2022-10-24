<template>
    <el-dialog v-model="createVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>创建数据库</span>
            </div>
        </template>
        <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input clearable v-model="form.name">
                    <template #append>
                        <el-select v-model="form.format" style="width: 125px">
                            <el-option label="utf8mb4" value="utf8mb4" />
                            <el-option label="utf-8" value="utf-8" />
                            <el-option label="gbk" value="gbk" />
                            <el-option label="big5" value="big5" />
                        </el-select>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('auth.username')" prop="username">
                <el-input clearable v-model="form.username" />
            </el-form-item>
            <el-form-item :label="$t('auth.password')" prop="password">
                <el-input type="password" clearable show-password v-model="form.password" />
            </el-form-item>

            <el-form-item :label="$t('database.permission')" prop="permission">
                <el-select style="width: 100%" v-model="form.permission">
                    <el-option value="localhost" :label="$t('database.permissionLocal')" />
                    <el-option value="%" :label="$t('database.permissionAll')" />
                    <el-option value="ip" :label="$t('database.permissionForIP')" />
                </el-select>
            </el-form-item>
            <el-form-item v-if="form.permission === 'ip'" prop="permissionIPs">
                <el-input clearable v-model="form.permissionIPs" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input type="textarea" clearable v-model="form.description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="createVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { addMysqlDB } from '@/api/modules/database';

const createVisiable = ref(false);
const form = reactive({
    name: '',
    version: '',
    format: '',
    username: '',
    password: '',
    permission: '',
    permissionIPs: '',
    description: '',
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput],
    permission: [Rules.requiredSelect],
    permissionIPs: [Rules.requiredInput],
});
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    version: string;
}
const acceptParams = (params: DialogProps): void => {
    form.name = '';
    form.version = params.version;
    form.format = 'utf8mb4';
    form.username = '';
    form.password = '';
    form.permission = 'localhost';
    form.permissionIPs = '';
    form.description = '';
    createVisiable.value = true;
};

const emit = defineEmits<{ (e: 'search'): void }>();
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.permission === 'ip') {
            form.permission = form.permissionIPs;
        }
        await addMysqlDB(form);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
        createVisiable.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
