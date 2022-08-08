<template>
    <div class="upload-box">
        <el-upload
            action="#"
            :class="['upload', disabled ? 'is-view' : '']"
            :multiple="false"
            :disabled="disabled"
            :show-file-list="false"
            :http-request="handleHttpUpload"
            :before-upload="beforeUpload"
            :on-success="uploadSuccess"
            :on-error="uploadError"
            :style="uploadStyle"
            :drag="drag"
            accept="image/jpeg,image/jpeg,image/png"
        >
            <img v-if="imageUrl" :src="imageUrl" class="image" />
            <el-icon v-else class="upload-icon"><Plus /></el-icon>
            <div v-if="imageUrl" class="view-icon" @click.stop="imageView">
                <img src="@/assets/images/enlarge_black.svg" alt="enlarge" />
            </div>
        </el-upload>
        <div class="el-upload__tip">
            <slot name="tip"></slot>
        </div>
        <el-image-viewer
            v-if="dialogVisible"
            @close="imageView"
            :url-list="[imageUrl]"
        />
    </div>
</template>

<script setup lang="ts" name="uploadImg">
import { ref } from 'vue';
import { ElNotification } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import { uploadImg } from '@/api/modules/upload';
import type { UploadProps, UploadRequestOptions } from 'element-plus';

interface UploadFileProps {
    imageUrl: string; // 图片地址 ==> 必传
    drag?: boolean; // 是否支持拖拽上传 ==> 非必传（默认为true）
    disabled?: boolean; // 是否禁用上传组件 ==> 非必传（默认为false）
    fileSize?: number; // 单个文件大小限制 ==> 非必传（默认为 5）
    uploadStyle?: { [key: string]: any }; // 上传组件样式 ==> 非必传
}
// 接受父组件参数
const props = withDefaults(defineProps<UploadFileProps>(), {
    drag: true,
    disabled: false,
    fileSize: 5,
    uploadStyle: () => ({ width: '175px', height: '175px' }),
});

/**
 * @description 图片上传
 * @param options 上传的文件
 * */
interface UploadEmits {
    (e: 'update:imageUrl', value: string): void;
    (e: 'check-validate'): void;
}
const emit = defineEmits<UploadEmits>();
const handleHttpUpload = async (options: UploadRequestOptions) => {
    let formData = new FormData();
    formData.append('file', options.file);
    try {
        const { data } = await uploadImg(formData);
        emit('update:imageUrl', data!.fileUrl);
        emit('check-validate');
    } catch (error) {
        options.onError(error as any);
    }
};

/**
 * @description 文件上传之前判断
 * @param rawFile 上传的文件
 * */
const beforeUpload: UploadProps['beforeUpload'] = (rawFile) => {
    const isLt3M = rawFile.size / 1024 / 1024 < props.fileSize;
    const imgType = ['image/jpg', 'image/jpeg', 'image/png'];
    if (!imgType.includes(rawFile.type))
        ElNotification({
            title: '温馨提示',
            message: '上传图片必须是 JPG/JPEG/PNG 格式！',
            type: 'warning',
        });
    if (!isLt3M)
        ElNotification({
            title: '温馨提示',
            message: `上传图片大小不能超过 ${props.fileSize}MB！`,
            type: 'warning',
        });
    return imgType.includes(rawFile.type) && isLt3M;
};

// 图片上传成功提示
const uploadSuccess = () => {
    ElNotification({
        title: '温馨提示',
        message: '图片上传成功！',
        type: 'success',
    });
};

// 图片上传错误提示
const uploadError = () => {
    ElNotification({
        title: '温馨提示',
        message: '图片上传失败，请您重新上传！',
        type: 'error',
    });
};

// 查看图片
const dialogVisible = ref(false);
const imageView = () => {
    dialogVisible.value = !dialogVisible.value;
};
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
