<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.makeImage')" :resource="form.containerName" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form @submit.prevent ref="formRef" :model="form" label-position="top">
                    <el-form-item prop="newImageName" :rules="Rules.imageName">
                        <template #label>
                            {{ $t('container.newImageName') }}
                        </template>
                        <el-input v-model="form.newImageName" />
                    </el-form-item>
                    <el-form-item prop="comment">
                        <template #label>
                            {{ $t('container.commitMessage') }}
                        </template>
                        <el-input v-model="form.comment" />
                    </el-form-item>
                    <el-form-item prop="author">
                        <template #label>
                            {{ $t('container.author') }}
                        </template>
                        <el-input v-model="form.author" />
                    </el-form-item>
                    <el-form-item prop="pause">
                        <el-checkbox v-model="form.pause">
                            {{ $t('container.ifPause') }}
                        </el-checkbox>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { commitContainer } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref<boolean>(false);
const emit = defineEmits<{ (e: 'search'): void }>();
const loading = ref(false);
const form = reactive({
    containerID: '',
    containerName: '',
    newImageName: '',
    comment: '',
    author: '',
    pause: false,
});

interface DialogProps {
    containerID: string;
    containerName: string;
}
const acceptParams = (props: DialogProps): void => {
    form.containerID = props.containerID;
    form.containerName = props.containerName;
    drawerVisible.value = true;
};

const formRef = ref<FormInstance>();
type FormInstance = InstanceType<typeof ElForm>;

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('container.ifMakeImageWithContainer'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            await commitContainer(form)
                .then(() => {
                    loading.value = false;
                    emit('search');
                    drawerVisible.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const handleClose = async () => {
    drawerVisible.value = false;
    emit('search');
};

defineExpose({
    acceptParams,
});
</script>
