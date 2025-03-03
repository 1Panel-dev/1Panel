<template>
    <DrawerPro v-model="detailVisible" :header="$t('commons.button.view')" @close="handleClose" size="large">
        <CodemirrorPro
            :placeholder="$t('commons.msg.noneData')"
            v-model="detailInfo"
            mode="yaml"
            :heightDiff="160"
            :disabled="true"
        ></CodemirrorPro>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="detailVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const detailVisible = ref(false);
const detailInfo = ref();

interface DialogProps {
    content: string;
}
const acceptParams = (params: DialogProps): void => {
    detailInfo.value = params.content;
    detailVisible.value = true;
};

const handleClose = () => {
    detailVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
