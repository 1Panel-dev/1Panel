<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.upgrade')" :resource="form.containerName" :back="handleClose" />
        </template>

        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-alert
                    :title="$t('container.appHelper')"
                    v-if="form.fromApp"
                    class="common-prompt"
                    :closable="false"
                    type="warning"
                />
                <el-form @submit.prevent ref="formRef" :model="form" label-position="top">
                    <el-form-item :label="$t('container.oldImage')" prop="oldImage">
                        <el-tooltip
                            placement="top-start"
                            :content="form.oldImageName"
                            v-if="form.oldImageName.length > 50"
                        >
                            <el-tag>{{ form.oldImageName.substring(0, 50) }}...</el-tag>
                        </el-tooltip>
                        <el-tag v-else>{{ form.oldImageName }}</el-tag>
                    </el-form-item>
                    <el-form-item prop="newImageName" :rules="Rules.imageName">
                        <template #label>
                            {{ $t('container.targetImage') }}
                            <span v-if="!form.hasName">
                                {{ ' (' + $t('container.imageLoadErr') + ')' }}
                            </span>
                        </template>
                        <el-input v-model="form.newImageName" />
                        <span class="input-help">{{ $t('container.upgradeHelper') }}</span>
                    </el-form-item>
                    <el-form-item prop="forcePull">
                        <el-checkbox v-model="form.forcePull">
                            {{ $t('container.forcePull') }}
                        </el-checkbox>
                        <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
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

<script lang="ts" setup>
import { upgradeContainer } from '@/api/modules/container';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElForm } from 'element-plus';
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';

const loading = ref(false);

const form = reactive({
    containerName: '',
    oldImageName: '',
    newImageName: '',
    hasName: true,

    fromApp: false,
    forcePull: false,
});

const formRef = ref<FormInstance>();

const drawerVisible = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;

interface DialogProps {
    container: string;
    image: string;
    fromApp: boolean;
}
const acceptParams = (props: DialogProps): void => {
    form.containerName = props.container;
    form.oldImageName = props.image;
    form.fromApp = props.fromApp;
    form.hasName = props.image.indexOf('sha256:') === -1;
    if (form.hasName) {
        form.newImageName = props.image;
    } else {
        form.newImageName = '';
    }
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('container.upgradeWarning2'), i18n.global.t('commons.button.upgrade'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            await upgradeContainer(form.containerName, form.newImageName, form.forcePull)
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
