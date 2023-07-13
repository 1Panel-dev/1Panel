<template>
    <el-drawer v-model="dialogVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('database.databaseConnInfo')" :back="handleClose" />
        </template>
        <el-form @submit.prevent v-loading="loading" ref="formRef" :model="form" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('commons.login.password')" :rules="Rules.requiredInput" prop="password">
                        <el-input type="password" show-password clearable v-model="form.password">
                            <template #append>
                                <el-button @click="onCopy(form.password)">{{ $t('commons.button.copy') }}</el-button>
                                <el-divider direction="vertical" />
                                <el-button style="margin-left: 1px" @click="random">
                                    {{ $t('commons.button.random') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('database.serviceName')" prop="serviceName">
                        <el-tag>{{ form.serviceName }}</el-tag>
                        <el-button @click="onCopy(form.serviceName)" icon="DocumentCopy" link></el-button>
                        <span class="input-help">{{ $t('database.serviceNameHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('database.containerConn')">
                        <el-tag>
                            {{ form.serviceName + ':6379' }}
                        </el-tag>
                        <el-button @click="onCopy(form.serviceName + ':6379')" icon="DocumentCopy" link></el-button>
                        <span class="input-help">
                            {{ $t('database.containerConnHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('database.remoteConn')">
                        <el-tag>{{ $t('database.localIP') + ':' + form.port }}</el-tag>
                        <span class="input-help">{{ $t('database.remoteConnHelper2') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="dialogVisiable = false">
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
import { MsgError, MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { App } from '@/api/interface/app';
import { getRandomStr } from '@/utils/util';
import useClipboard from 'vue-clipboard3';
const { toClipboard } = useClipboard();

const loading = ref(false);

const dialogVisiable = ref(false);
const form = ref<App.DatabaseConnInfo>({
    password: '',
    serviceName: '',
    port: 0,
});

const confirmDialogRef = ref();

const emit = defineEmits(['checkExist', 'closeTerminal']);

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const acceptParams = (): void => {
    form.value.password = '';
    loadPassword();
    dialogVisiable.value = true;
};
const handleClose = () => {
    dialogVisiable.value = false;
};

const random = async () => {
    form.value.password = getRandomStr(16);
};

const onCopy = async (value: string) => {
    try {
        await toClipboard(value);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyfailed'));
    }
};

const loadPassword = async () => {
    const res = await GetAppConnInfo('redis');
    form.value = res.data;
};

const onSubmit = async () => {
    let param = {
        id: 0,
        value: form.value.password,
    };
    loading.value = true;
    emit('closeTerminal');
    await changeRedisPassword(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            dialogVisiable.value = false;
            emit('checkExist');
        })
        .catch(() => {
            loading.value = false;
        });
};

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
