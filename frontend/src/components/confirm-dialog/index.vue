<template>
    <el-dialog v-model="submitVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
        <template #header>
            <div class="card-header">
                <span>{{ header }}</span>
            </div>
        </template>
        <div>
            <span style="font-size: 12px">{{ operationInfo }}</span>
            <el-input v-model="submitInput"></el-input>
            <span style="font-size: 12px">{{ $t('commons.msg.operateConfirm') }}</span>
            <span style="font-size: 12px; color: red; font-weight: 500">'{{ submitInputInfo }}'</span>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="submitVisiable = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="submitInput !== submitInputInfo" @click="onConfirm">
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
const submitVisiable = ref(false);

const submitInput = ref();

interface DialogProps {
    header: string;
    operationInfo: string;
    submitInputInfo: string;
}

const acceptParams = (props: DialogProps): void => {
    submitVisiable.value = true;
    header.value = props.header;
    operationInfo.value = props.operationInfo;
    submitInputInfo.value = props.submitInputInfo;
    submitInput.value = '';
};
const emit = defineEmits<{ (e: 'confirm'): void }>();

const onConfirm = async () => {
    emit('confirm');
    submitVisiable.value = false;
};

defineExpose({
    acceptParams,
});
</script>
