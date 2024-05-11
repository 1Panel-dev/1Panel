<template>
    <div>
        <el-drawer
            v-model="bindVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            width="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('database.userBind')" :resource="form.mysqlName" :back="handleClose" />
            </template>
            <el-form v-loading="loading" ref="changeFormRef" :model="form" :rules="rules" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.login.username')" prop="username">
                            <el-input v-model="form.username"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.password')" prop="password">
                            <el-input type="password" clearable show-password v-model="form.password"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('database.permission')" prop="permission">
                            <el-select v-model="form.permission">
                                <el-option value="%" :label="$t('database.permissionAll')" />
                                <el-option
                                    v-if="form.from !== 'local'"
                                    value="localhost"
                                    :label="$t('terminal.localhost')"
                                />
                                <el-option value="ip" :label="$t('database.permissionForIP')" />
                            </el-select>
                        </el-form-item>
                        <el-form-item v-if="form.permission === 'ip'" prop="permissionIPs">
                            <el-input clearable :rows="3" type="textarea" v-model="form.permissionIPs" />
                            <span class="input-help">{{ $t('database.remoteHelper') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
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
        </el-drawer>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { bindUser } from '@/api/modules/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import { checkIp } from '@/utils/util';

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
    password: [Rules.paramComplexity],
    permission: [Rules.requiredSelect],
    permissionIPs: [{ validator: checkIPs, trigger: 'blur', required: true }],
});

function checkIPs(rule: any, value: any, callback: any) {
    let ips = form.permissionIPs.split(',');
    for (const item of ips) {
        if (checkIp(item)) {
            return callback(new Error(i18n.global.t('commons.rule.ip')));
        }
    }
    callback();
}

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
