<template>
    <el-dialog v-model="submitVisible" :destroy-on-close="true" :close-on-click-modal="false" width="20%">
        <template #header>
            <div class="card-header">
                <span>{{ header }}</span>
            </div>
        </template>
        <div>
            <span v-if="operationInfo" style="font-size: 12px">{{ operationInfo }}</span>
            <div :style="{ 'margin-top': operationInfo ? '10px' : '0px' }">
                <span style="font-size: 12px">{{ $t('commons.msg.operateConfirm') }}</span>
                <span style="font-size: 12px; color: red; font-weight: 500">'{{ submitInputInfo }}'</span>
            </div>
            <el-input style="margin-top: 10px" v-model="submitInput"></el-input>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="onCancel">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :disabled="submitInput !== submitInputInfo" @click="onConfirm">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const header = ref();
const operationInfo = ref();
const submitInputInfo = ref();
const submitVisible = ref(false);

const submitInput = ref();

interface DialogProps {
    header: string;
    operationInfo: string;
    submitInputInfo: string;
}

const acceptParams = (props: DialogProps): void => {
    submitVisible.value = true;
    header.value = props.header;
    operationInfo.value = props.operationInfo;
    submitInputInfo.value = props.submitInputInfo;
    submitInput.value = '';
};
const emit = defineEmits(['confirm', 'cancel']);

const onConfirm = async () => {
    emit('confirm');
    submitVisible.value = false;
};

const onCancel = async () => {
    emit('cancel');
    submitVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
