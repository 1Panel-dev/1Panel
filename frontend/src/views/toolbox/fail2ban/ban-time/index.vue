<template>
    <div>
        <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('toolbox.fail2ban.banTime')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('toolbox.fail2ban.banTime')" prop="banTime" :rules="Rules.number">
                            <el-input type="number" v-model.number="form.banTime">
                                <template #append>
                                    <el-select v-model.number="form.banTimeUnit" style="width: 115px">
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
    banTime: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    banTime: -1,
    banTimeUnit: 's',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.banTime = Number(params.banTime);
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateSetting({ key: 'banTime', value: form.banTime })
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
