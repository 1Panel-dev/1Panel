<template>
    <div>
        <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('toolbox.fail2ban.findTime')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('toolbox.fail2ban.findTime')"
                            prop="findTime"
                            :rules="Rules.requiredInput"
                        >
                            <el-input clearable v-model="form.findTime">
                                <template #append>
                                    <el-select v-model="form.findTimeUnit" style="width: 115px">
                                        <el-option :label="$t('commons.units.second')" value="s" />
                                        <el-option :label="$t('commons.units.minute')" value="m" />
                                        <el-option :label="$t('commons.units.hour')" value="h" />
                                        <el-option :label="$t('commons.units.day')" value="d" />
                                        <el-option :label="$t('commons.units.year')" value="y" />
                                    </el-select>
                                </template>
                            </el-input>
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
import { updateSetting } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    findTime: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    findTime: '',
    findTimeUnit: 's',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.findTime = Number(params.findTime);
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateSetting({ key: 'findTime', value: form.findTime })
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
