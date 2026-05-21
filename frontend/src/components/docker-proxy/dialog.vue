<template>
    <DialogPro v-model="open" :title="$t('commons.msg.infoTitle')" size="small">
        <el-form ref="formRef" label-position="top" @submit.prevent>
            <el-form-item :label="$t('xpack.node.syncProxyHelper')">
                <el-radio-group v-model="restart">
                    <el-radio :value="true">{{ $t('setting.restartNow') }}</el-radio>
                    <el-radio :value="false">{{ $t('setting.restartLater') }}</el-radio>
                </el-radio-group>
                <span class="input-help" v-if="restart">{{ $t('xpack.node.syncProxyHelper1') }}</span>
                <span class="input-help" v-else>{{ $t('xpack.node.syncProxyHelper2') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onConfirm">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { getSettingBaseInfo } from '@/api/modules/setting';
import { getXpackProxyDocker } from '@/extensions/xpack';

const open = ref(false);
const restart = ref(true);

const em = defineEmits(['update:withDockerRestart', 'submit']);
interface DialogProps {
    syncList: string;
    open: boolean;
}
const emit = () => {
    em('update:withDockerRestart', false);
    em('submit');
};
const acceptParams = async (props: DialogProps): Promise<void> => {
    if (props.syncList.indexOf('SyncSystemProxy') === -1) {
        emit();
        return;
    }
    if (props.open) {
        open.value = true;
        return;
    }
    try {
        const res = await getSettingBaseInfo();
        if (res.data.proxyType === '' || res.data.proxyType === 'close') {
            emit();
            return;
        }
    } catch (error) {
        emit();
        return;
    }

    const res = await getXpackProxyDocker();
    if (!res) {
        emit();
        return;
    }
    if (res.data.proxyDocker !== 'Enable') {
        emit();
        return;
    }
    open.value = true;
};

const onConfirm = async () => {
    em('update:withDockerRestart', restart.value);
    em('submit');
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>
