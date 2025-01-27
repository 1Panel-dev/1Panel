<template>
    <DrawerPro v-model="open" :header="$t('commons.button.log')" size="large" :back="handleClose">
        <LogFile :config="config" :height-diff="config.heightDiff"></LogFile>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import LogFile from '@/components/log-file/index.vue';

interface LogProps {
    id: number;
    type: string;
    style: string;
    name: string;
    tail: boolean;
    heightDiff: number;
}

const open = ref(false);
const config = ref();
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = (props: LogProps) => {
    config.value = props;
    open.value = true;
};

defineExpose({ acceptParams });
</script>

<style lang="scss">
.fullScreen {
    border: none;
}
</style>
