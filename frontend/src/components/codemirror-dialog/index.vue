<template>
    <DrawerPro v-model="codeVisible" :header="header" size="large" @close="handleClose">
        <CodemirrorPro v-model="detailInfo" :height="800" :disabled="true" mode="json"></CodemirrorPro>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="codeVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
const header = ref();
const detailInfo = ref();
const codeVisible = ref(false);

interface DialogProps {
    header: string;
    detailInfo: string;
}

const acceptParams = (props: DialogProps): void => {
    header.value = props.header;
    detailInfo.value = props.detailInfo;
    codeVisible.value = true;
};

const handleClose = () => {
    codeVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
