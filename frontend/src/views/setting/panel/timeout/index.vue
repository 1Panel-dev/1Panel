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
                <DrawerHeader :header="$t('setting.sessionTimeout')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                :rules="rules"
                :model="form"
                @submit.prevent
                v-loading="loading"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.sessionTimeout')" prop="sessionTimeout">
                            <el-input clearable v-model.number="form.sessionTimeout" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSaveTimeout(formRef)">
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
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateSetting } from '@/api/modules/setting';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    sessionTimeout: number;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    sessionTimeout: 86400,
});

const rules = reactive({
    sessionTimeout: [Rules.integerNumber, checkNumberRange(300, 864000)],
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.sessionTimeout = params.sessionTimeout;
    drawerVisible.value = true;
};

const onSaveTimeout = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateSetting({ key: 'SessionTimeout', value: form.sessionTimeout + '' })
            .then(async () => {
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
