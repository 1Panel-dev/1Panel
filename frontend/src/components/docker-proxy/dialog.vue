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
import { getSettingInfo } from '@/api/modules/setting';

const open = ref(false);
const restart = ref(true);

const em = defineEmits(['update:withDockerRestart', 'submit']);
interface DialogProps {
    syncList: string;
}
const acceptParams = async (props: DialogProps): Promise<void> => {
    if (props.syncList.indexOf('SyncSystemProxy') === -1) {
        em('update:withDockerRestart', false);
        em('submit');
        return;
    }
    await getSettingInfo()
        .then((res) => {
            if (res.data.proxyType === '' || res.data.proxyType === 'close') {
                em('update:withDockerRestart', false);
                em('submit');
                return;
            }
        })
        .catch(() => {
            em('update:withDockerRestart', false);
            em('submit');
            return;
        });
    let searchXSetting;
    const xpackModules = import.meta.glob('../../xpack/api/modules/setting.ts', { eager: true });
    if (xpackModules['../../xpack/api/modules/setting.ts']) {
        searchXSetting = xpackModules['../../xpack/api/modules/setting.ts']['searchXSetting'] || {};
        const res = await searchXSetting();
        if (!res) {
            em('update:withDockerRestart', false);
            em('submit');
            return;
        }
        if (res.data.proxyDocker === '') {
            em('update:withDockerRestart', false);
            em('submit');
            return;
        }
        open.value = true;
    }
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
