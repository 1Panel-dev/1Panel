<template>
    <el-drawer v-model="createVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('database.create')" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.name')" prop="name">
                            <el-input clearable v-model.trim="form.name">
                                <template #append>
                                    <el-select v-model="form.format" style="width: 120px">
                                        <el-option label="utf8mb4" value="utf8mb4" />
                                        <el-option label="utf-8" value="utf8" />
                                        <el-option label="gbk" value="gbk" />
                                        <el-option label="big5" value="big5" />
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.username')" prop="username">
                            <el-input clearable v-model.trim="form.username" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.password')" prop="password">
                            <el-input type="password" clearable show-password v-model.trim="form.password">
                                <template #append>
                                    <el-button @click="random">{{ $t('commons.button.random') }}</el-button>
                                </template>
                            </el-input>
                        </el-form-item>

                        <el-form-item :label="$t('database.permission')" prop="permission">
                            <el-select v-model="form.permission">
                                <el-option value="%" :label="$t('database.permissionAll')" />
                                <el-option value="ip" :label="$t('database.permissionForIP')" />
                            </el-select>
                        </el-form-item>
                        <el-form-item v-if="form.permission === 'ip'" prop="permissionIPs">
                            <el-input
                                clearable
                                :autosize="{ minRows: 2, maxRows: 5 }"
                                type="textarea"
                                v-model="form.permissionIPs"
                            />
                            <span class="input-help">{{ $t('database.remoteHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input type="textarea" clearable v-model="form.description" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
        </div>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="createVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { addMysqlDB } from '@/api/modules/database';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { getRandomStr } from '@/utils/util';

const loading = ref();
const createVisiable = ref(false);
const form = reactive({
    name: '',
    mysqlName: '',
    format: '',
    username: '',
    password: '',
    permission: '',
    permissionIPs: '',
    description: '',
});
const rules = reactive({
    name: [Rules.requiredInput, Rules.dbName],
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput],
    permission: [Rules.requiredSelect],
    permissionIPs: [Rules.requiredInput],
});
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    mysqlName: string;
}
const acceptParams = (params: DialogProps): void => {
    form.name = '';
    form.mysqlName = params.mysqlName;
    form.format = 'utf8mb4';
    form.username = '';
    form.permission = '%';
    form.permissionIPs = '';
    form.description = '';
    random();
    createVisiable.value = true;
};
const handleClose = () => {
    createVisiable.value = false;
};

const random = async () => {
    form.password = getRandomStr(16);
};

const emit = defineEmits<{ (e: 'search'): void }>();
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.permission === 'ip') {
            form.permission = form.permissionIPs;
        }
        loading.value = true;
        await addMysqlDB(form)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                createVisiable.value = false;
            })
            .catch(() => {
                if (form.permission != '%') {
                    form.permissionIPs = form.permission;
                    form.permission = 'ip';
                }
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
