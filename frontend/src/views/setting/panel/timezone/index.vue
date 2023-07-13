<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.timeZone')" :back="handleClose" />
            </template>
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-alert
                        v-if="canChangeZone()"
                        :title="$t('setting.timeZoneHelper')"
                        class="common-prompt"
                        :closable="false"
                        type="warning"
                    />
                    <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                        <el-form-item :label="$t('setting.timeZone')" prop="timeZone" :rules="Rules.requiredInput">
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
                                {{ $t('setting.timeZoneCN') }}
                            </el-button>
                            <el-button
                                :disabled="canChangeZone()"
                                type="primary"
                                link
                                class="tagClass"
                                @click="form.timeZone = 'America/Los_Angeles'"
                            >
                                {{ $t('setting.timeZoneAM') }}
                            </el-button>
                            <el-button
                                :disabled="canChangeZone()"
                                type="primary"
                                link
                                class="tagClass"
                                @click="form.timeZone = 'America/New_York'"
                            >
                                {{ $t('setting.timeZoneNY') }}
                            </el-button>
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { loadTimeZone, updateSetting } from '@/api/modules/setting';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { GlobalStore } from '@/store';
import DrawerHeader from '@/components/drawer-header/index.vue';
const globalStore = GlobalStore();

interface DialogProps {
    timeZone: string;
}
const drawerVisiable = ref();
const loading = ref();

const form = reactive({
    timeZone: '',
});

const formRef = ref<FormInstance>();
const zones = ref<Array<string>>([]);

const acceptParams = (params: DialogProps): void => {
    loadTimeZones();
    form.timeZone = params.timeZone;
    drawerVisiable.value = true;
};

const loadTimeZones = async () => {
    const res = await loadTimeZone();
    zones.value = res.data;
};

const canChangeZone = () => {
    return zones.value.length === 0;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('setting.timeZoneChangeHelper'), i18n.global.t('setting.timeZone'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            await updateSetting({ key: 'TimeZone', value: form.timeZone })
                .then(async () => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    globalStore.isLogin = false;
                    let href = window.location.href;
                    window.open(href, '_self');
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const handleClose = () => {
    drawerVisiable.value = false;
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
