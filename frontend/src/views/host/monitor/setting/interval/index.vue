<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('monitor.interval')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('monitor.interval')"
                            :rules="[Rules.integerNumber, checkNumberRange(1, 60)]"
                            prop="monitorInterval"
                        >
                            <el-input clearable v-model.number="form.monitorInterval" />
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
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { updateSetting } from '@/api/modules/setting';
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    monitorInterval: number;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    monitorInterval: 1,
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.monitorInterval = params.monitorInterval;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateSetting({ key: 'MonitorInterval', value: form.monitorInterval + '' })
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
