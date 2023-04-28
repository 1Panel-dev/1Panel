<template>
    <div>
        <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.entrance')" :back="handleClose" />
            </template>
            <el-form label-position="top" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.entrance')" prop="days">
                            <el-input clearable v-model="securityEntrance">
                                <template #append>
                                    <el-button @click="random" icon="RefreshRight"></el-button>
                                </template>
                            </el-input>
                            <span class="input-help">
                                {{ $t('setting.entranceInputHelper') }}
                            </span>
                            <span class="input-error" v-if="codeError">
                                {{ $t('setting.entranceError') }}
                            </span>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="submitEntrance">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { updateSetting } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { getRandomStr } from '@/utils/util';
const globalStore = GlobalStore();

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    securityEntrance: string;
}
const securityEntrance = ref();
const drawerVisiable = ref();
const loading = ref();
const codeError = ref();

const acceptParams = (params: DialogProps): void => {
    securityEntrance.value = params.securityEntrance;
    drawerVisiable.value = true;
};

const random = async () => {
    securityEntrance.value = getRandomStr(8);
};

const submitEntrance = async () => {
    if (securityEntrance.value !== '') {
        const reg = /^[A-Za-z0-9]{6,10}$/;
        if (!reg.test(securityEntrance.value)) {
            codeError.value = true;
            return;
        }
    }
    let param = {
        key: 'SecurityEntrance',
        value: securityEntrance.value,
    };
    loading.value = true;
    await updateSetting(param)
        .then(() => {
            globalStore.entrance = securityEntrance.value;
            loading.value = false;
            drawerVisiable.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
        })
        .catch(() => {
            loading.value = false;
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
.margintop {
    margin-top: 10px;
}
</style>
