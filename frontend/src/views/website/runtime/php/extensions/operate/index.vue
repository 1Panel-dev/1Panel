<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.' + operate) + $t('php.extensions')"
        :close-on-click-modal="false"
        width="30%"
        :before-close="handleClose"
    >
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form @submit.prevent ref="extensionsForm" label-position="top" :model="extensions" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="extensions.name" :disabled="operate == 'edit'"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('php.extension')" prop="extensions">
                        <el-input
                            type="textarea"
                            :placeholder="$t('php.extensionsHelper')"
                            :rows="3"
                            v-model="extensions.extensions"
                        />
                    </el-form-item>
                    <el-link target="_blank" type="primary" :href="globalStore.docsUrl + phpDocURL">
                        {{ $t('php.toExtensionsList') }}
                    </el-link>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit(extensionsForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import { FormInstance } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import { CreatePHPExtensions, UpdatePHPExtensions } from '@/api/modules/runtime';
import i18n from '@/lang';
import { Runtime } from '@/api/interface/runtime';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const phpDocURL = globalStore.isIntl ? `/user_manual/websites/runtime_php/` : '/user_manual/websites/php/#php_1';

const open = ref(false);
const operate = ref('create');
const loading = ref(false);
const updateID = ref(0);
const extensionsForm = ref<FormInstance>();
const rules = ref({
    name: [Rules.requiredInput, Rules.name],
    extensions: [Rules.requiredInput, Rules.phpExtensions],
});
const em = defineEmits(['close']);

const initData = () => ({
    name: '',
    extensions: '',
});

const extensions = ref(initData());

const acceptParams = (op: string, extend: Runtime.PHPExtensions) => {
    operate.value = op;
    open.value = true;
    extensions.value = initData();
    if (operate.value == 'edit') {
        extensions.value = extend;
        updateID.value = extend.id;
    }
};

const handleClose = () => {
    open.value = false;
    extensionsForm.value?.resetFields();
    em('close', false);
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        if (operate.value == 'create') {
            CreatePHPExtensions(extensions.value)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            UpdatePHPExtensions({
                id: updateID.value,
                name: extensions.value.name,
                extensions: extensions.value.extensions,
            })
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
