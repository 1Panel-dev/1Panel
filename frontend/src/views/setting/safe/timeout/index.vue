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
                <DrawerHeader :header="$t('setting.expirationTime')" :back="handleClose" />
            </template>
            <el-form ref="timeoutFormRef" @submit.prevent label-position="top" :model="form">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('setting.days')"
                            prop="days"
                            :rules="[Rules.integerNumberWith0, checkNumberRange(0, 60)]"
                        >
                            <el-input clearable v-model.number="form.days" />
                            <span class="input-help">{{ $t('setting.expirationHelper') }}</span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submitTimeout(timeoutFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { updateSetting } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import DrawerHeader from '@/components/drawer-header/index.vue';
import i18n from '@/lang';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    expirationDays: number;
}
const drawerVisible = ref();
const loading = ref();

const timeoutFormRef = ref();
const form = reactive({
    days: 0,
});

const acceptParams = (params: DialogProps): void => {
    form.days = params.expirationDays;
    drawerVisible.value = true;
};

const submitTimeout = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateSetting({ key: 'ExpirationDays', value: form.days + '' })
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
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
