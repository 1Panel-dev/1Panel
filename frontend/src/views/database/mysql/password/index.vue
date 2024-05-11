<template>
    <div>
        <el-drawer
            v-model="changeVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            width="30%"
        >
            <template #header>
                <DrawerHeader :header="title" :resource="changeForm.mysqlName" :back="handleClose" />
            </template>
            <el-form v-loading="loading" ref="changeFormRef" :model="changeForm" :rules="rules" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <div v-if="changeForm.operation === 'password'">
                            <el-form-item :label="$t('commons.login.username')" prop="userName">
                                <el-input disabled v-model="changeForm.userName"></el-input>
                            </el-form-item>
                            <el-form-item :label="$t('commons.login.password')" prop="password">
                                <el-input
                                    type="password"
                                    clearable
                                    show-password
                                    v-model="changeForm.password"
                                ></el-input>
                            </el-form-item>
                        </div>
                        <div v-if="changeForm.operation === 'privilege'">
                            <el-form-item :label="$t('database.permission')" prop="privilege">
                                <el-select style="width: 100%" v-model="changeForm.privilege">
                                    <el-option value="%" :label="$t('database.permissionAll')" />
                                    <el-option
                                        v-if="changeForm.from !== 'local'"
                                        value="localhost"
                                        :label="$t('terminal.localhost')"
                                    />
                                    <el-option value="ip" :label="$t('database.permissionForIP')" />
                                </el-select>
                            </el-form-item>
                            <el-form-item v-if="changeForm.privilege === 'ip'" prop="privilegeIPs">
                                <el-input clearable :rows="3" type="textarea" v-model="changeForm.privilegeIPs" />
                                <span class="input-help">{{ $t('database.remoteHelper') }}</span>
                            </el-form-item>
                        </div>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="changeVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitChangeInfo(changeFormRef)">
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
import { deleteCheckMysqlDB, updateMysqlAccess, updateMysqlPassword } from '@/api/modules/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import { checkIp } from '@/utils/util';

const loading = ref();
const changeVisible = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const changeFormRef = ref<FormInstance>();
const title = ref();
const changeForm = reactive({
    id: 0,
    from: '',
    type: '',
    database: '',
    mysqlName: '',
    userName: '',
    password: '',
    operation: '',
    privilege: '',
    privilegeIPs: '',
    value: '',
});
const confirmDialogRef = ref();

const rules = reactive({
    password: [Rules.paramComplexity],
    privilegeIPs: [{ validator: checkIPs, trigger: 'blur', required: true }],
});

function checkIPs(rule: any, value: any, callback: any) {
    let ips = changeForm.privilegeIPs.split(',');
    for (const item of ips) {
        if (checkIp(item)) {
            return callback(new Error(i18n.global.t('commons.rule.ip')));
        }
    }
    callback();
}

interface DialogProps {
    id: number;
    from: string;
    type: string;
    database: string;
    mysqlName: string;
    username: string;
    password: string;
    operation: string;
    privilege: string;
    privilegeIPs: string;
    value: string;
}
const acceptParams = (params: DialogProps): void => {
    title.value =
        params.operation === 'password'
            ? i18n.global.t('database.changePassword')
            : i18n.global.t('database.permission');
    changeForm.id = params.id;
    changeForm.from = params.from;
    changeForm.type = params.type;
    changeForm.database = params.database;
    changeForm.mysqlName = params.mysqlName;
    changeForm.userName = params.username;
    changeForm.password = params.password;
    changeForm.operation = params.operation;
    changeForm.privilege = params.privilege;
    changeForm.privilegeIPs = params.privilegeIPs;
    changeForm.value = params.value;
    changeVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    changeVisible.value = false;
};

const submitChangeInfo = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            id: changeForm.id,
            from: changeForm.from,
            type: changeForm.type,
            database: changeForm.database,
            value: '',
        };
        if (changeForm.operation === 'password') {
            const res = await deleteCheckMysqlDB(param);
            if (res.data && res.data.length > 0) {
                let params = {
                    header: i18n.global.t('database.changePassword'),
                    operationInfo: i18n.global.t('database.changePasswordHelper'),
                    submitInputInfo: i18n.global.t('database.restartNow'),
                };
                confirmDialogRef.value!.acceptParams(params);
            } else {
                param.value = changeForm.password;
                loading.value = true;
                await updateMysqlPassword(param)
                    .then(() => {
                        loading.value = false;
                        emit('search');
                        changeVisible.value = false;
                        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    })
                    .catch(() => {
                        loading.value = false;
                    });
            }
            return;
        }
        if (changeForm.privilege !== 'ip') {
            param.value = changeForm.privilege;
        } else {
            param.value = changeForm.privilegeIPs;
        }
        loading.value = true;
        await updateMysqlAccess(param)
            .then(() => {
                loading.value = false;
                emit('search');
                changeVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onSubmit = async () => {
    let param = {
        id: changeForm.id,
        from: changeForm.from,
        type: changeForm.type,
        database: changeForm.database,
        value: changeForm.password,
    };
    loading.value = true;
    await updateMysqlPassword(param)
        .then(() => {
            loading.value = false;
            emit('search');
            changeVisible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
