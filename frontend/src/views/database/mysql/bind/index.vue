<template>
    <div>
        <DrawerPro
            v-model="bindVisible"
            :header="$t('database.userBind')"
            :resource="form.mysqlName"
            @close="handleClose"
            size="small"
        >
            <el-form v-loading="loading" ref="changeFormRef" :model="form" :rules="rules" label-position="top">
                <el-form-item :label="$t('commons.login.username')" prop="username">
                    <el-input v-model="form.username"></el-input>
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')" prop="password">
                    <el-input type="password" clearable show-password v-model="form.password"></el-input>
                    <span class="input-help">{{ $t('commons.rule.illegalChar') }}</span>
                </el-form-item>
                <el-form-item :label="$t('database.permission')" prop="permission">
                    <el-select v-model="form.permission">
                        <el-option value="%" :label="$t('database.permissionAll')" />
                        <el-option
                            v-if="form.from !== 'local'"
                            value="localhost"
                            :label="$t('terminal.localhost') + '(localhost)'"
                        />
                        <el-option value="ip" :label="$t('database.permissionForIP')" />
                    </el-select>
                    <span v-if="form.from !== 'local'" class="input-help">
                        {{ $t('database.localhostHelper') }}
                    </span>
                </el-form-item>
                <el-form-item v-if="form.permission === 'ip'" prop="permissionIPs">
                    <el-input clearable :rows="3" type="textarea" v-model="form.permissionIPs" />
                    <span class="input-help">{{ $t('database.remoteHelper') }}</span>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="bindVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit(changeFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DrawerPro>
        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { bindUser } from '@/api/modules/database';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const bindVisible = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const changeFormRef = ref<FormInstance>();
const form = reactive({
    from: '',
    database: '',
    mysqlName: '',
    username: '',
    password: '',
    permission: '',
    permissionIPs: '',
});
const confirmDialogRef = ref();

const rules = reactive({
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
    permission: [Rules.requiredSelect],
    permissionIPs: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
});

interface DialogProps {
    from: string;
    database: string;
    mysqlName: string;
}
const acceptParams = (params: DialogProps): void => {
    form.database = params.database;
    form.mysqlName = params.mysqlName;
    form.username = '';
    form.password = '';
    form.permission = '%';
    form.from = params.from;
    bindVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    bindVisible.value = false;
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            database: form.database,
            db: form.mysqlName,
            username: form.username,
            password: form.password,
            permission: form.permission === 'ip' ? form.permissionIPs : form.permission,
        };
        loading.value = true;
        await bindUser(param)
            .then(() => {
                loading.value = false;
                emit('search');
                bindVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
