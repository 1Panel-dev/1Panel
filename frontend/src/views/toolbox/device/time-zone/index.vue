<template>
    <DrawerPro v-model="drawerVisible" :header="$t('toolbox.device.timeZone')" @close="handleClose" size="small">
        <el-alert :title="$t('toolbox.device.timeZoneHelper')" class="common-prompt" :closable="false" type="warning" />
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('toolbox.device.timeZone')" prop="timeZone" :rules="Rules.requiredInput">
                <el-select filterable :disabled="canChangeZone()" v-model="form.timeZone">
                    <el-option v-for="item in zones" :key="item" :label="item" :value="item" />
                </el-select>
                <el-button
                    :disabled="canChangeZone()"
                    type="primary"
                    link
                    class="tagClass"
                    @click="form.timeZone = 'Asia/Shanghai'"
                >
                    {{ $t('toolbox.device.timeZoneCN') }}
                </el-button>
                <el-button
                    :disabled="canChangeZone()"
                    type="primary"
                    link
                    class="tagClass"
                    @click="form.timeZone = 'America/Los_Angeles'"
                >
                    {{ $t('toolbox.device.timeZoneAM') }}
                </el-button>
                <el-button
                    :disabled="canChangeZone()"
                    type="primary"
                    link
                    class="tagClass"
                    @click="form.timeZone = 'America/New_York'"
                >
                    {{ $t('toolbox.device.timeZoneNY') }}
                </el-button>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { GlobalStore } from '@/store';
import { loadTimeZoneOptions, updateDevice } from '@/api/modules/toolbox';
const globalStore = GlobalStore();

interface DialogProps {
    timeZone: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    timeZone: '',
});

const formRef = ref<FormInstance>();
const zones = ref<Array<string>>([]);

const acceptParams = (params: DialogProps): void => {
    loadTimeZones();
    form.timeZone = params.timeZone;
    drawerVisible.value = true;
};

const loadTimeZones = async () => {
    const res = await loadTimeZoneOptions();
    zones.value = res.data;
};

const canChangeZone = () => {
    return zones.value.length === 0;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('toolbox.device.timeZoneChangeHelper'),
            i18n.global.t('toolbox.device.timeZone'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            await updateDevice('TimeZone', form.timeZone)
                .then(async () => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    if (globalStore.currentNode === 'local') {
                        globalStore.isLogin = false;
                        let href = window.location.href;
                        window.open(href, '_self');
                    } else {
                        globalStore.currentNode = 'local';
                        location.reload();
                    }
                })
                .catch(() => {
                    loading.value = false;
                    let href = window.location.href;
                    window.open(href, '_self');
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
<style scoped lang="scss">
.tagClass {
    margin-top: 5px;
}
</style>
