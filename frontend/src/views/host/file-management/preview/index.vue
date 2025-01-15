<template>
    <el-dialog
        v-model="open"
        :show-close="false"
        :before-close="handleClose"
        destroy-on-close
        append-to-body
        @opened="onOpen"
        :class="isFullscreen ? 'w-full' : '!w-3/4'"
        :top="'5vh'"
        :fullscreen="isFullscreen"
    >
        <template #header>
            <div class="flex items-center justify-between">
                <span>{{ $t('commons.button.preview') + ' - ' + filePath }}</span>
                <el-space alignment="center" :size="10" class="dialog-header-icon">
                    <el-tooltip :content="loadTooltip()" placement="top" v-if="fileType !== 'excel'">
                        <el-icon @click="toggleFullscreen"><FullScreen /></el-icon>
                    </el-tooltip>
                    <el-icon @click="handleClose" size="20"><Close /></el-icon>
                </el-space>
            </div>
        </template>
        <div v-loading="loading" :style="isFullscreen ? 'height: 90vh' : 'height: 80vh'">
            <div class="flex items-center justify-center h-full">
                <el-image
                    v-if="fileType === 'image'"
                    :src="fileUrl"
                    :style="isFullscreen ? 'height: 90vh' : 'height: 80vh'"
                    fit="contain"
                    :preview-src-list="[fileUrl]"
                />

                <video v-else-if="fileType === 'video'" :src="fileUrl" controls autoplay class="size-3/4"></video>

                <audio v-else-if="fileType === 'audio'" :src="fileUrl" controls></audio>

                <vue-office-docx
                    v-else-if="fileType === 'word'"
                    :src="fileUrl"
                    :style="isFullscreen ? 'height: 90vh' : 'height: 80vh'"
                    class="w-full"
                    @rendered="renderedHandler"
                    @error="errorHandler"
                />

                <vue-office-excel
                    v-else-if="fileType === 'excel'"
                    :src="fileUrl"
                    :style="isFullscreen ? 'height: 90vh;' : 'height: 80vh'"
                    class="w-full"
                    @rendered="renderedHandler"
                    @error="errorHandler"
                />
            </div>
        </div>
    </el-dialog>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { ref } from 'vue';

import { Close, FullScreen } from '@element-plus/icons-vue';
import VueOfficeDocx from '@vue-office/docx';
import VueOfficeExcel from '@vue-office/excel';
import '@vue-office/docx/lib/index.css';
import '@vue-office/excel/lib/index.css';
import { MsgError } from '@/utils/message';

interface EditProps {
    fileType: string;
    path: string;
    name: string;
    extension: string;
}

const open = ref(false);
const loading = ref(false);
const filePath = ref('');
const fileName = ref('');
const fileType = ref('');
const fileUrl = ref('');

const fileExtension = ref('');
const isFullscreen = ref(false);
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;

    em('close', open.value);
};

const renderedHandler = () => {
    loading.value = false;
};
const errorHandler = () => {
    open.value = false;
    MsgError(i18n.global.t('commons.msg.unSupportType'));
};
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullscreen.value ? 'quitFullscreen' : 'fullscreen'));
};

const toggleFullscreen = () => {
    isFullscreen.value = !isFullscreen.value;
};

const acceptParams = (props: EditProps) => {
    fileExtension.value = props.extension;
    fileName.value = props.name;
    filePath.value = props.path;
    fileType.value = props.fileType;
    isFullscreen.value = fileType.value === 'excel';

    loading.value = true;
    fileUrl.value = `${import.meta.env.VITE_API_URL as string}/files/download?path=${encodeURIComponent(
        props.path,
    )}&timestamp=${new Date().getTime()}`;
    open.value = true;
    loading.value = false;
};

const onOpen = () => {};

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.dialog-top {
    top: 0;
}

.dialog-header-icon {
    color: var(--el-color-info);
}
</style>
