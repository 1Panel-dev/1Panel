<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.runtimeEnv')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('setting.region')" prop="edition">
                <el-radio-group v-model="form.edition">
                    <el-radio value="cn">
                        <span>{{ $t('setting.cn') }}</span>
                    </el-radio>
                    <el-radio value="intl">
                        <span>{{ $t('setting.intl') }}</span>
                    </el-radio>
                </el-radio-group>
                <span class="input-help">{{ $t('setting.regionTip') }}</span>
            </el-form-item>
            <el-form-item :label="$t('setting.docSource')" prop="docSource">
                <el-radio-group v-model="form.docSource">
                    <el-radio value="withByRegion">
                        <span>{{ $t('setting.withByRegion') }}</span>
                    </el-radio>
                    <el-radio value="withByLang">
                        <span>{{ $t('setting.withByLang') }}</span>
                    </el-radio>
                </el-radio-group>
                <span class="input-help">{{ $t('setting.docSourceTip') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onSaveEdition()">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updateSetting } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { docWithRegion, isIntl } = useGlobalStore();

interface DialogProps {
    edition: string;
    docSource: string;
}
const drawerVisible = ref();
const loading = ref(false);

const form = reactive({
    edition: '',
    docSource: 'withByRegion',
});
const initialForm = reactive({
    edition: '',
    docSource: 'withByRegion',
});

const formRef = ref<FormInstance>();
const emit = defineEmits(['search']);

const acceptParams = (params: DialogProps): void => {
    form.edition = params.edition;
    form.docSource = params.docSource || 'withByRegion';
    initialForm.edition = params.edition;
    initialForm.docSource = params.docSource || 'withByRegion';
    drawerVisible.value = true;
};

const onSaveEdition = async () => {
    const editionChanged = form.edition !== initialForm.edition;
    const docSourceChanged = form.docSource !== initialForm.docSource;
    if (!editionChanged && !docSourceChanged) {
        drawerVisible.value = false;
        return;
    }

    if (editionChanged) {
        try {
            await ElMessageBox.confirm(i18n.global.t('setting.regionHelper'), i18n.global.t('setting.region'), {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                dangerouslyUseHTMLString: true,
            });
        } catch {
            return;
        }
    }

    loading.value = true;
    try {
        if (editionChanged) {
            await updateSetting({ key: 'Edition', value: form.edition });
            isIntl.value = form.edition === 'intl';
        }
        if (docSourceChanged) {
            await updateSetting({ key: 'DocSource', value: form.docSource });
            docWithRegion.value = form.docSource === 'withByRegion';
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        if (editionChanged) {
            location.reload();
            return;
        }
        emit('search');
        drawerVisible.value = false;
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
