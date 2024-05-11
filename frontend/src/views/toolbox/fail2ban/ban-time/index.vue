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
                <DrawerHeader :header="$t('toolbox.fail2ban.banTime')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                :model="form"
                :rules="rules"
                @submit.prevent
                v-loading="loading"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('toolbox.fail2ban.banTime')" prop="banTime">
                            <el-input type="number" v-model.number="form.banTime">
                                <template #append>
                                    <el-select v-model.number="form.banTimeUnit" style="width: 100px">
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
import { FormInstance } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateFail2ban } from '@/api/modules/toolbox';
import { splitTime, transTimeUnit } from '@/utils/util';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    banTime: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    banTime: 300,
    banTimeUnit: 's',
});

const rules = reactive({
    banTime: [{ validator: checkBanTime, trigger: 'blur' }],
});
function checkBanTime(rule: any, value: any, callback: any) {
    if (value === -1) {
        callback();
    }
    const reg = /^[1-9]\d*$/;
    if (!reg.test(value)) {
        return callback(new Error(i18n.global.t('toolbox.fail2ban.banTimeRule')));
    }
    callback();
}

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    let item = splitTime(params.banTime);
    form.banTime = item.time;
    form.banTimeUnit = item.unit;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let itemMsg =
            form.banTime === -1
                ? i18n.global.t('toolbox.fail2ban.banAllTime')
                : form.banTime + transTimeUnit(form.banTimeUnit);
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('toolbox.fail2ban.banTime'), itemMsg]),
            i18n.global.t('toolbox.fail2ban.fail2banChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            let itemValue = form.banTime === -1 ? '-1' : form.banTime + form.banTimeUnit;
            await updateFail2ban({ key: 'bantime', value: itemValue })
                .then(async () => {
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    loading.value = false;
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
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
