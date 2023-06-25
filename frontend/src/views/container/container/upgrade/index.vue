<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('container.upgrade')" :resource="form.name" :back="handleClose" />
        </template>

        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-alert
                    v-if="form.fromApp"
                    :title="$t('container.appHelper')"
                    style="margin-bottom: 20px"
                    :closable="false"
                    type="error"
                />
                <el-form @submit.prevent ref="formRef" v-loading="loading" :model="form" label-position="top">
                    <el-form-item :label="$t('container.oldImage')" prop="oldImage">
                        <el-tag>{{ form.imageName }}:{{ form.oldTag }}</el-tag>
                    </el-form-item>
                    <el-form-item :label="$t('container.targetImage')" prop="newTag" :rules="Rules.requiredInput">
                        <el-input v-model="form.newTag">
                            <template #prefix>{{ form.imageName }}:</template>
                        </el-input>
                        <span class="input-help">{{ $t('container.upgradeHelper') }}</span>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisiable = false">
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
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { ElForm } from 'element-plus';
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';

const loading = ref(false);

const form = reactive({
    name: '',
    imageName: '',
    oldTag: '',
    newTag: '',
    fromApp: false,
});

const formRef = ref<FormInstance>();

const drawerVisiable = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;

interface DialogProps {
    container: string;
    image: string;
    fromApp: boolean;
}
const acceptParams = (props: DialogProps): void => {
    form.name = props.container;
    form.imageName = props.image.indexOf(':') !== -1 ? props.image.split(':')[0] : props.image;
    form.oldTag = props.image.indexOf(':') !== -1 ? props.image.split(':')[1] : 'latest';
    form.newTag = '';
    form.fromApp = props.fromApp;
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!compareVersion(form.newTag, form.oldTag)) {
            MsgWarning(i18n.global.t('container.upgradeWarning'));
            return;
        }
        ElMessageBox.confirm(i18n.global.t('container.upgradeWarning2'), i18n.global.t('container.upgrade'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            await upgradeContainer(form.name, form.imageName + ':' + form.newTag)
                .then(() => {
                    loading.value = false;
                    emit('search');
                    drawerVisiable.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};
const handleClose = async () => {
    drawerVisiable.value = false;
    emit('search');
};

function compareVersion(vNew, vOld) {
    if (vNew === 'latest') {
        return true;
    }
    let v1 = vNew
        .replace('-', '.')
        .replace(/[^\d.]/g, '')
        .split('.');
    let v2 = vOld
        .replace('-', '.')
        .replace(/[^\d.]/g, '')
        .split('.');

    for (let i = 0; i < Math.max(v1.length, v2.length); i++) {
        let num1 = parseInt(v1[i] || 0);
        let num2 = parseInt(v2[i] || 0);

        if (num1 > num2) {
            return true;
        } else if (num1 < num2) {
            return false;
        }
    }

    return false;
}

defineExpose({
    acceptParams,
});
</script>
