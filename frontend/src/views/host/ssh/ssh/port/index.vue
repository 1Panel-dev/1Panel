<template>
    <DrawerPro v-model="drawerVisible" :header="$t('ssh.port')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('ssh.port')" prop="port" :rules="rules.port">
                <el-input clearable v-model="form.port" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, FormInstance } from 'element-plus';
import { updateSSH } from '@/api/modules/host';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    port: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    port: '22',
});

const rules = reactive({
    port: [{ validator: checkPorts, trigger: 'blur', required: true }],
});

function checkPorts(rule: any, value: string, callback: any) {
    const ports = value
        .split(',')
        .map((item) => item.trim())
        .filter((item) => item !== '');
    if (ports.length === 0) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    const portSet = new Set<string>();
    for (const port of ports) {
        const portNumber = Number(port);
        if (!/^\d+$/.test(port) || Number.isNaN(portNumber) || portNumber < 1 || portNumber > 65535) {
            callback(new Error(i18n.global.t('commons.rule.port')));
            return;
        }
        if (portSet.has(port)) {
            callback(new Error(i18n.global.t('commons.rule.port')));
            return;
        }
        portSet.add(port);
    }
    callback();
}

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.port = params.port;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('commons.table.port'), form.port]),
            i18n.global.t('ssh.sshChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        )
            .then(async () => {
                let params = {
                    key: 'Port',
                    newValue: form.port,
                };
                loading.value = true;
                await updateSSH(params)
                    .then(() => {
                        loading.value = false;
                        handleClose();
                        emit('search');
                        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    })
                    .catch(() => {
                        loading.value = false;
                    });
            })
            .catch(() => {
                emit('search');
            });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
