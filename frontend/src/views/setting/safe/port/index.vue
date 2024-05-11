<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.panelPort')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.panelPort')" prop="serverPort" :rules="Rules.port">
                            <el-input clearable v-model.number="form.serverPort" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSavePort(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updatePort } from '@/api/modules/setting';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { GlobalStore } from '@/store';
import DrawerHeader from '@/components/drawer-header/index.vue';
const globalStore = GlobalStore();

interface DialogProps {
    serverPort: number;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    serverPort: 9999,
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.serverPort = params.serverPort;
    drawerVisible.value = true;
};

const onSavePort = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('setting.portChangeHelper'), i18n.global.t('setting.portChange'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            loading.value = true;
            let param = {
                serverPort: form.serverPort,
            };
            await updatePort(param)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    globalStore.isLogin = false;
                    let href = window.location.href;
                    let ip = href.split('//')[1].split(':')[0];
                    if (globalStore.entrance) {
                        window.open(
                            `${href.split('//')[0]}//${ip}:${form.serverPort}/${globalStore.entrance}`,
                            '_self',
                        );
                    } else {
                        window.open(`${href.split('//')[0]}//${ip}:${form.serverPort}/login`, '_self');
                    }
                })
                .catch(() => {
                    loading.value = false;
                });
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
