<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.title')" :back="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('setting.title')" prop="panelName">
                <el-input clearable v-model="form.panelName" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onSavePanelName(formRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updateSetting } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    panelName: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    panelName: '',
});
const rules = reactive({
    panelName: [{ validator: checkPanelName, trigger: 'blur', required: true }],
});

function checkPanelName(rule: any, value: any, callback: any) {
    if (value === '') {
        return callback(new Error(i18n.global.t('setting.titleHelper')));
    }
    const reg = /^[a-zA-Z0-9\u4e00-\u9fa5 .,:!@#%&^*_+[\]{}~\-=?，。！｜？：；「」『』【】（）《》·]{3,30}$/;
    if (!reg.test(value)) {
        return callback(new Error(i18n.global.t('setting.titleHelper')));
    }
    callback();
}

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.panelName = params.panelName;
    drawerVisible.value = true;
};

const onSavePanelName = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateSetting({ key: 'PanelName', value: form.panelName })
            .then(async () => {
                globalStore.setThemeConfig({ ...themeConfig.value, panelName: form.panelName });
                document.title = form.panelName;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                return;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
