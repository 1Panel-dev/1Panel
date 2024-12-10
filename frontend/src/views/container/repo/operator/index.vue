<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="30%"
    >
        <template #header>
            <DrawerHeader
                :header="title + $t('container.repo').toLowerCase()"
                :resource="dialogData.rowData?.name"
                :back="handleClose"
            />
        </template>
        <el-form
            ref="formRef"
            label-position="top"
            v-loading="loading"
            :model="dialogData.rowData"
            :rules="rules"
            label-width="120px"
        >
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input
                            clearable
                            :disabled="dialogData.title === 'edit'"
                            v-model.trim="dialogData.rowData!.name"
                        ></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('container.auth')" prop="auth">
                        <el-radio-group v-model="dialogData.rowData!.auth">
                            <el-radio :value="true">{{ $t('commons.true') }}</el-radio>
                            <el-radio :value="false">{{ $t('commons.false') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item v-if="dialogData.rowData!.auth" :label="$t('commons.login.username')" prop="username">
                        <el-input clearable v-model.trim="dialogData.rowData!.username"></el-input>
                    </el-form-item>
                    <el-form-item v-if="dialogData.rowData!.auth" :label="$t('commons.login.password')" prop="password">
                        <el-input
                            clearable
                            type="password"
                            show-password
                            v-model.trim="dialogData.rowData!.password"
                        ></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('container.downloadUrl')" prop="downloadUrl">
                        <el-input
                            clearable
                            v-model.trim="dialogData.rowData!.downloadUrl"
                            :placeholder="'172.16.10.10:8081'"
                        ></el-input>
                        <span v-if="dialogData.rowData!.downloadUrl" class="input-help">
                            Pull example: docker pull {{ dialogData.rowData!.downloadUrl }}/nginx
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                        <el-radio-group v-model="dialogData.rowData!.protocol">
                            <el-radio label="http">http</el-radio>
                            <el-radio label="https">https</el-radio>
                        </el-radio-group>
                        <span v-if="dialogData.rowData!.protocol === 'http'" class="input-help">
                            {{ $t('container.httpRepo') }}
                        </span>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
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
import { Container } from '@/api/interface/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { createImageRepo, updateImageRepo } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

interface DialogProps {
    title: string;
    rowData?: Container.RepoInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};
const rules = reactive({
    name: [Rules.requiredInput, Rules.name],
    downloadUrl: [{ validator: validateDownloadUrl, trigger: 'blur' }, Rules.illegal],
    protocol: [Rules.requiredSelect],
    username: [Rules.illegal],
    password: [Rules.illegal],
    auth: [Rules.requiredSelect],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

function validateDownloadUrl(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    const pattern = /^(http:\/\/|https:\/\/)/i;
    if (pattern.test(value)) {
        return callback(new Error(i18n.global.t('container.urlWarning')));
    }
    callback();
}

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        if (dialogData.value.title === 'add') {
            await createImageRepo(dialogData.value.rowData!)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await updateImageRepo(dialogData.value.rowData!)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
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
