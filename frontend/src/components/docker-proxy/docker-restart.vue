<template>
    <DialogPro v-model="open" size="small">
        <el-form ref="formRef" label-position="top" @submit.prevent>
            <el-form-item :label="title">
                <el-radio-group v-model="restart">
                    <el-radio :value="true">{{ $t('setting.restartNow') }}</el-radio>
                    <el-radio :value="false">{{ $t('setting.restartLater') }}</el-radio>
                </el-radio-group>
                <span class="input-help" v-if="restart">{{ $t('xpack.node.syncProxyHelper1') }}</span>
                <span class="input-help" v-else>{{ $t('xpack.node.syncProxyHelper2') }}</span>
            </el-form-item>
        </el-form>
        <slot name="helper" />
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
const open = ref(false);
const title = ref();
const restart = ref(true);

const em = defineEmits(['update:withDockerRestart', 'submit']);
interface DialogProps {
    title: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    title.value = params.title;
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
