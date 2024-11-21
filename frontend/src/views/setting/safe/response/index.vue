<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.noAuthSetting')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('setting.responseSetting')"
                            prop="noAuthSetting"
                            :rules="Rules.requiredSelect"
                        >
                            <el-select v-model="form.noAuthSetting">
                                <el-option
                                    v-for="item in options"
                                    :key="item.value"
                                    :label="item.label"
                                    :value="item.value"
                                />
                            </el-select>
                            <span class="input-help">{{ $t('setting.noAuthSettingHelper') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { updateSetting } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const drawerVisible = ref();
const loading = ref();
const formRef = ref<FormInstance>();
const emit = defineEmits<{ (e: 'search'): void }>();

const form = reactive({
    noAuthSetting: '',
});

const options = [
    {
        value: '200',
        label: '200 - ' + i18n.global.t('setting.help200'),
    },
    {
        value: '400',
        label: '400 - ' + i18n.global.t('setting.error400'),
    },
    {
        value: '401',
        label: '401 - ' + i18n.global.t('setting.error401'),
    },
    {
        value: '403',
        label: '403 - ' + i18n.global.t('setting.error403'),
    },
    {
        value: '404',
        label: '404 - ' + i18n.global.t('setting.error404'),
    },
    {
        value: '408',
        label: '408 - ' + i18n.global.t('setting.error408'),
    },
    {
        value: '416',
        label: '416 - ' + i18n.global.t('setting.error416'),
    },
    {
        value: '444',
        label: '444 - ' + i18n.global.t('setting.error444'),
    },
    {
        value: '500',
        label: '500 - ' + i18n.global.t('setting.error500'),
    },
];

interface DialogProps {
    noAuthSetting: string;
    noAuthOptions: [{ value: string; label: string }];
}

const acceptParams = (params: DialogProps): void => {
    form.noAuthSetting = params.noAuthSetting;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateSetting({ key: 'NoAuthSetting', value: form.noAuthSetting })
            .then(() => {
                loading.value = false;
                handleClose();
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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

<style scoped lang="scss"></style>
