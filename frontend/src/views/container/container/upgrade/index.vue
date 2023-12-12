<template>
    <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('commons.button.upgrade')" :resource="form.name" :back="handleClose" />
        </template>

        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-alert
                    v-if="form.fromApp"
                    :title="$t('container.appHelper')"
                    class="common-prompt"
                    :closable="false"
                    type="error"
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
                            <el-tooltip
                                placement="top-start"
                                :content="form.imageHelper"
                                v-if="form.imageHelper.length > 40"
                            >
                                <span>
                                    {{
                                        $t('container.targetImage') + ' (' + form.imageHelper.substring(0, 40) + '...)'
                                    }}
                                </span>
                            </el-tooltip>
                            <span v-else>
                                {{ $t('container.targetImage') + ' (' + form.imageHelper + ')' }}
                            </span>
                        </template>
                        <el-input v-model="form.newImageName" :placeholder="$t('container.imageNameHelper')" />
                        <span class="input-help">{{ $t('container.upgradeHelper') }}</span>
                    </el-form-item>
                    <el-form-item prop="ignoreCompare">
                        <el-checkbox v-model="form.ignoreCompare">
                            {{ $t('container.ignoreCompare') }}
                        </el-checkbox>
                        <span class="input-help">{{ $t('container.ignoreCompareHelper') }}</span>
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
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { ElForm } from 'element-plus';
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';

const loading = ref(false);

const form = reactive({
    name: '',
    oldImageName: '',
    newImageName: '',
    imageHelper: '',
    fromApp: false,
    forcePull: false,

    ignoreCompare: false,
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
    form.name = props.container;
    form.oldImageName = props.image;
    form.fromApp = props.fromApp;
    form.ignoreCompare = false;

    if (props.image.indexOf('sha256:') !== -1) {
        form.imageHelper = i18n.global.t('container.imageLoadErr');
        drawerVisible.value = true;
        return;
    }
    form.imageHelper = props.image.indexOf(':') !== -1 ? props.image.split(':')[0] : props.image;
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!form.ignoreCompare && !compareVersion(form.newImageName, form.oldImageName)) {
            MsgWarning(i18n.global.t('container.upgradeWarning'));
            return;
        }
        ElMessageBox.confirm(i18n.global.t('container.upgradeWarning2'), i18n.global.t('commons.button.upgrade'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            loading.value = true;
            await upgradeContainer(form.name, form.newImageName, form.forcePull)
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

function compareVersion(vNew, vOld) {
    let newImageName = vNew.indexOf(':') !== -1 ? vNew.split(':')[0] : vNew;
    let oldImageName = vOld.indexOf(':') !== -1 ? vOld.split(':')[0] : vOld;
    if (newImageName !== oldImageName) {
        return true;
    }
    let newTag = vNew.indexOf(':') !== -1 ? vNew.split(':')[1] : 'latest';

    if (newTag === 'latest') {
        return true;
    }
    let oldTag = vOld.indexOf(':') !== -1 ? vOld.split(':')[1] : 'latest';
    let v1 = newTag
        .replace('-', '.')
        .replace(/[^\d.]/g, '')
        .split('.');
    let v2 = oldTag
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
