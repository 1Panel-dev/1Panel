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
                <DrawerHeader :header="$t('setting.entrance')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                :model="form"
                @submit.prevent
                v-loading="loading"
                :rules="rules"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.entrance')" prop="securityEntrance">
                            <el-input clearable v-model="form.securityEntrance">
                                <template #append>
                                    <el-button @click="random">{{ $t('setting.randomGenerate') }}</el-button>
                                </template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('setting.entranceInputHelper') }}
                            </span>
                        </el-form-item>
                        <el-form-item>
                            <el-checkbox v-model="show" :label="$t('setting.showEntrance')" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="submitEntrance(formRef)">
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
import { GlobalStore } from '@/store';
import { getRandomStr } from '@/utils/util';
import { FormInstance } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
const globalStore = GlobalStore();

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    securityEntrance: string;
}
const drawerVisible = ref();
const loading = ref();
const show = ref();

const form = reactive({
    securityEntrance: '',
});

const formRef = ref<FormInstance>();
const rules = reactive({
    securityEntrance: [{ validator: checkSecurityEntrance, trigger: 'blur' }],
});

function checkSecurityEntrance(rule: any, value: any, callback: any) {
    if (form.securityEntrance !== '') {
        const reg = /^[A-Za-z0-9]{5,116}$/;
        if (!reg.test(form.securityEntrance)) {
            return callback(new Error(i18n.global.t('setting.entranceError')));
        }
    }
    callback();
}

const acceptParams = (params: DialogProps): void => {
    form.securityEntrance = params.securityEntrance;
    show.value = globalStore.showEntranceWarn;
    drawerVisible.value = true;
};

const random = async () => {
    form.securityEntrance = getRandomStr(10);
};

const submitEntrance = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            key: 'SecurityEntrance',
            value: form.securityEntrance,
        };
        loading.value = true;
        await updateSetting(param)
            .then(() => {
                globalStore.setShowEntranceWarn(show.value);
                globalStore.entrance = form.securityEntrance;
                loading.value = false;
                drawerVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
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
