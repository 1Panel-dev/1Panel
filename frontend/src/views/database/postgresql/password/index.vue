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
                <DrawerHeader :header="title" :resource="changeForm.postgresqlName" :back="handleClose" />
            </template>
            <el-form v-loading="loading" ref="changeFormRef" :rules="rules" :model="changeForm" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <div v-if="changeForm.operation === 'password'">
                            <el-form-item :label="$t('commons.login.username')" prop="username">
                                <el-input disabled v-model="changeForm.username"></el-input>
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
import { deleteCheckPostgresqlDB, updatePostgresqlPassword } from '@/api/modules/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';

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
    postgresqlName: '',
    username: '',
    password: '',
    operation: '',
    value: '',
});
const confirmDialogRef = ref();
const rules = reactive({
    password: [Rules.paramComplexity],
});

interface DialogProps {
    id: number;
    from: string;
    type: string;
    database: string;
    postgresqlName: string;
    username: string;
    password: string;
    operation: string;
    privilege: string;
    privilegeIPs: string;
    value: string;
}
const acceptParams = (params: DialogProps): void => {
    title.value = i18n.global.t('database.changePassword');
    changeForm.id = params.id;
    changeForm.from = params.from;
    changeForm.type = params.type;
    changeForm.database = params.database;
    changeForm.postgresqlName = params.postgresqlName;
    changeForm.username = params.username;
    changeForm.password = params.password;
    changeForm.operation = params.operation;
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
            const res = await deleteCheckPostgresqlDB(param);
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
                await updatePostgresqlPassword(param)
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

        loading.value = true;
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
    await updatePostgresqlPassword(param)
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
