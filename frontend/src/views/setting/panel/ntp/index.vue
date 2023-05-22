<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.syncTime')" :back="handleClose" />
            </template>
            <el-alert v-if="canChangeZone()" style="margin-bottom: 20px" :closable="false" type="warning">
                <template #default>
                    <span>
                        <span>{{ $t('setting.timeZoneHelper') }}</span>
                    </span>
                </template>
            </el-alert>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
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

                        <el-form-item :label="$t('setting.syncTime')" prop="localTime">
                            <el-input v-model="form.localTime" disabled />
                        </el-form-item>

                        <el-form-item :label="$t('setting.syncSite')" prop="ntpSite" :rules="Rules.requiredInput">
                            <el-input v-model="form.ntpSite" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSyncTime(formRef)">
                        {{ $t('commons.button.sync') }}
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
import { loadTimeZone, syncTime } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    timeZone: string;
    localTime: string;
    ntpSite: string;
}
const drawerVisiable = ref();
const loading = ref();
const zones = ref<Array<string>>([]);
const oldTimeZone = ref();

const form = reactive({
    timeZone: '',
    localTime: '',
    ntpSite: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    loadTimeZones();
    oldTimeZone.value = params.timeZone;
    form.timeZone = params.timeZone;
    form.localTime = params.localTime;
    form.ntpSite = params.ntpSite;
    drawerVisiable.value = true;
};

const canChangeZone = () => {
    return zones.value.length === 0;
};

const loadTimeZones = async () => {
    const res = await loadTimeZone();
    zones.value = res.data;
};
const onSyncTime = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await syncTime(form.timeZone, form.ntpSite)
            .then((res) => {
                loading.value = false;
                form.localTime = res.data;
                emit('search');
                handleClose();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
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
