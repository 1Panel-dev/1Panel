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
                <DrawerHeader :header="$t('database.userBind')" :resource="form.name" :back="handleClose" />
            </template>
            <el-form v-loading="loading" ref="changeFormRef" :model="form" :rules="rules" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-alert type="warning" :title="$t('database.pgBindHelper')" :closable="false" />
                        <el-form-item class="mt-5" :label="$t('database.userBind')" prop="username">
                            <el-input v-model="form.username" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.password')" prop="password">
                            <el-input type="password" clearable show-password v-model="form.password" />
                        </el-form-item>
                        <el-form-item :label="$t('database.permission')" prop="superUser">
                            <el-checkbox v-model="form.superUser">{{ $t('database.pgSuperUser') }}</el-checkbox>
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
import { bindPostgresqlUser } from '@/api/modules/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const bindVisible = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const changeFormRef = ref<FormInstance>();
const form = reactive({
    database: '',
    name: '',
    username: '',
    password: '',
    superUser: true,
});
const confirmDialogRef = ref();

const rules = reactive({
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.paramComplexity],
});

interface DialogProps {
    database: string;
    name: string;
}
const acceptParams = (params: DialogProps): void => {
    form.database = params.database;
    form.name = params.name;
    form.username = '';
    form.password = '';
    form.superUser = true;
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
            name: form.name,
            database: form.database,
            username: form.username,
            password: form.password,
            superUser: form.superUser,
        };
        loading.value = true;
        await bindPostgresqlUser(param)
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
