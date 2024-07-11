<template>
    <el-dialog
        v-model="open"
        :title="$t('ssl.selfSigned')"
        :close-on-click-modal="false"
        width="40%"
        :before-close="handleClose"
        :destroy-on-close="true"
    >
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form @submit.prevent ref="obtainForm" label-position="top" :model="obtain" :rules="rules">
                    <el-form-item :label="$t('website.domain')" prop="domains">
                        <el-input
                            type="textarea"
                            :rows="4"
                            v-model="obtain.domains"
                            :placeholder="$t('ssl.domainHelper')"
                        ></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.remark')" prop="description">
                        <el-input v-model="obtain.description"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.keyType')" prop="keyType">
                        <el-select v-model="obtain.keyType">
                            <el-option
                                v-for="(keyType, index) in KeyTypes"
                                :key="index"
                                :label="keyType.label"
                                :value="keyType.value"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('ssl.days')" prop="time">
                        <el-input type="number" v-model.number="obtain.time">
                            <template #append>
                                <el-select v-model="obtain.unit" style="width: 100px">
                                    <el-option :label="$t('commons.units.day')" value="day"></el-option>
                                    <el-option :label="$t('commons.units.year')" value="year"></el-option>
                                </el-select>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="''" prop="autoRenew">
                        <el-checkbox v-model="obtain.autoRenew" :label="$t('ssl.autoRenew')" />
                    </el-form-item>
                    <el-form-item :label="''" prop="pushDir">
                        <el-checkbox v-model="obtain.pushDir" :label="$t('ssl.pushDir')" />
                    </el-form-item>
                    <el-form-item :label="$t('ssl.dir')" prop="dir" v-if="obtain.pushDir">
                        <el-input v-model.trim="obtain.dir">
                            <template #prepend>
                                <FileList :path="obtain.dir" @choose="getPath" :dir="true"></FileList>
                            </template>
                        </el-input>
                        <span class="input-help">
                            {{ $t('ssl.pushDirHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="''" prop="execShell">
                        <el-checkbox v-model="obtain.execShell" :label="$t('ssl.execShell')" />
                    </el-form-item>
                    <el-form-item :label="$t('ssl.shell')" prop="shell" v-if="obtain.execShell">
                        <el-input type="textarea" :rows="4" v-model="obtain.shell" />
                        <span class="input-help">
                            {{ $t('ssl.shellHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(obtainForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ObtainSSLByCA } from '@/api/modules/website';
import { Rules, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { KeyTypes } from '@/global/mimetype';

const open = ref(false);
const loading = ref(false);
const obtainForm = ref<FormInstance>();
const em = defineEmits(['close']);

const rules = ref({
    keyType: [Rules.requiredSelect],
    domains: [Rules.requiredInput],
    dir: [Rules.requiredInput],
    time: [Rules.integerNumber, checkNumberRange(1, 10000)],
    shell: [Rules.requiredInput],
});

const initData = () => ({
    keyType: 'P256',
    domains: '',
    id: 0,
    time: 10,
    unit: 'year',
    pushDir: false,
    dir: '',
    autoRenew: true,
    description: '',
    execShell: false,
    shell: '',
});
const obtain = ref(initData());

const acceptParams = (id: number) => {
    open.value = true;
    obtain.value.id = id;
};

const handleClose = () => {
    open.value = false;
    em('close', false);
    resetForm();
};

const resetForm = () => {
    obtainForm.value?.resetFields();
    obtain.value = initData();
};

const getPath = (dir: string) => {
    obtain.value.dir = dir;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;

        ObtainSSLByCA(obtain.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
