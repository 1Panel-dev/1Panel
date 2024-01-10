<template>
    <el-drawer v-model="dialogVisible" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('database.databaseConnInfo')" :back="handleClose" />
        </template>
        <el-form @submit.prevent v-loading="loading" ref="formRef" :model="form" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('database.containerConn')">
                        <el-tag>
                            {{ form.serviceName + ':6379' }}
                        </el-tag>
                        <CopyButton :content="form.serviceName + ':6379'" type="icon" />
                        <span class="input-help">
                            {{ $t('database.containerConnHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('database.remoteConn')">
                        <el-tooltip v-if="loadConnInfo().length > 48" :content="loadConnInfo()" placement="top">
                            <el-tag>{{ loadConnInfo().substring(0, 48) }}...</el-tag>
                        </el-tooltip>
                        <el-tag v-else>{{ loadConnInfo() }}</el-tag>
                        <CopyButton :content="form.systemIP + ':6379'" type="icon" />
                        <span class="input-help">{{ $t('database.remoteConnHelper2') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('commons.login.password')" :rules="Rules.paramComplexity" prop="password">
                        <el-input type="password" show-password clearable v-model="form.password">
                            <template #append>
                                <CopyButton :content="form.password" />
                                <el-button @click="random" class="p-ml-5">
                                    {{ $t('commons.button.random') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="dialogVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { changeRedisPassword } from '@/api/modules/database';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { GetAppConnInfo } from '@/api/modules/app';
import { MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { App } from '@/api/interface/app';
import { getRandomStr } from '@/utils/util';
import { getSettingInfo } from '@/api/modules/setting';

const loading = ref(false);

const dialogVisible = ref(false);
const form = ref<App.DatabaseConnInfo>({
    privilege: false,
    password: '',
    serviceName: '',
    systemIP: '',
    port: 0,
});

const confirmDialogRef = ref();

const emit = defineEmits(['checkExist', 'closeTerminal']);

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const acceptParams = (): void => {
    form.value.password = '';
    loadPassword();
    dialogVisible.value = true;
};
const handleClose = () => {
    dialogVisible.value = false;
};

const random = async () => {
    form.value.password = getRandomStr(16);
};

const loadPassword = async () => {
    const res = await GetAppConnInfo('redis', '');
    const settingInfoRes = await getSettingInfo();
    form.value = res.data;
    form.value.systemIP = settingInfoRes.data.systemIP || i18n.global.t('database.localIP');
};

const onSubmit = async () => {
    loading.value = true;
    emit('closeTerminal');
    await changeRedisPassword(form.value.password)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            dialogVisible.value = false;
            emit('checkExist');
        })
        .catch(() => {
            loading.value = false;
        });
};

function loadConnInfo() {
    return form.value.systemIP + ':' + form.value.port;
}

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

defineExpose({
    acceptParams,
});
</script>
