<template>
    <el-dialog v-model="repoVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="50%">
        <template #header>
            <div class="card-header">
                <span>{{ title }}{{ $t('container.repo') }}</span>
            </div>
        </template>
        <el-form ref="formRef" :model="dialogData.rowData" label-position="left" :rules="rules" label-width="120px">
            <el-form-item :label="$t('container.name')" prop="name">
                <el-input :disabled="dialogData.title === 'edit'" v-model="dialogData.rowData!.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.auth')" prop="auth">
                <el-radio-group v-model="dialogData.rowData!.auth">
                    <el-radio :label="true">{{ $t('commons.true') }}</el-radio>
                    <el-radio :label="false">{{ $t('commons.false') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="dialogData.rowData!.auth" :label="$t('auth.username')" prop="username">
                <el-input v-model="dialogData.rowData!.username"></el-input>
            </el-form-item>
            <el-form-item v-if="dialogData.rowData!.auth" :label="$t('auth.password')" prop="password">
                <el-input type="password" v-model="dialogData.rowData!.password"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.downloadUrl')" prop="downloadUrl">
                <el-input v-model="dialogData.rowData!.downloadUrl" :placeholder="'172.16.10.10:8081'"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.imageRepo')" prop="repoName">
                <el-checkbox v-model="hasRepo">{{ $t('container.repoHelper') }}</el-checkbox>
                <el-input v-if="hasRepo" v-model="dialogData.rowData!.repoName"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="repoVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { Container } from '@/api/interface/container';
import { repoCreate, repoUpdate } from '@/api/modules/container';

interface DialogProps {
    title: string;
    rowData?: Container.RepoInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const repoVisiable = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    hasRepo.value = params.rowData?.repoName ? params.rowData?.repoName.length !== 0 : false;
    repoVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const hasRepo = ref(false);
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    downloadUrl: [Rules.requiredInput],
    repoName: [Rules.requiredInput],
    username: [Rules.requiredInput],
    password: [Rules.requiredInput],
    auth: [Rules.requiredSelect],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

function restForm() {
    if (formRef.value) {
        formRef.value.resetFields();
    }
}
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (dialogData.value.title === 'create') {
            await repoCreate(dialogData.value.rowData!);
        }
        if (dialogData.value.title === 'edit') {
            await repoUpdate(dialogData.value.rowData!);
        }
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        restForm();
        emit('search');
        repoVisiable.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
