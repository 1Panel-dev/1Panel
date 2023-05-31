<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('container.registries')" :back="handleClose" />
            </template>
            <el-form label-position="top" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('container.registries')">
                            <el-input
                                type="textarea"
                                :placeholder="$t('container.registrieHelper')"
                                :autosize="{ minRows: 8, maxRows: 10 }"
                                v-model="registries"
                            />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit" />
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { updateDaemonJson } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

const confirmDialogRef = ref();

const registries = ref();
interface DialogProps {
    registries: string;
}
const drawerVisiable = ref();
const loading = ref();

const acceptParams = (params: DialogProps): void => {
    registries.value = params.registries || params.registries.replaceAll(',', '\n');
    drawerVisiable.value = true;
};

const onSave = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onSubmit = async () => {
    loading.value = true;
    await updateDaemonJson('Registries', registries.value.replaceAll('\n', ','))
        .then(() => {
            loading.value = false;
            handleClose();
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    drawerVisiable.value = false;
};

defineExpose({
    acceptParams,
});
</script>
