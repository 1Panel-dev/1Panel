<template>
    <div v-loading="loading">
        <el-form :model="form" :rules="variablesRules" ref="phpFormRef" label-position="top">
            <el-row v-loading="loading">
                <el-col :span="9">
                    <el-form-item label="short_open_tag" prop="short_open_tag">
                        <el-select v-model="form.short_open_tag">
                            <el-option :label="$t('website.isOff')" :value="'Off'"></el-option>
                            <el-option :label="$t('website.isOn')" :value="'On'"></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('php.short_open_tag') }}</span>
                    </el-form-item>
                    <el-form-item label="max_execution_time" prop="max_execution_time">
                        <el-input clearable v-model.number="form.max_execution_time" maxlength="15">
                            <template #append>{{ $t('commons.units.second') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('php.max_execution_time') }}</span>
                    </el-form-item>

                    <el-form-item label="post_max_size" prop="post_max_size">
                        <el-input clearable v-model.number="form.post_max_size" maxlength="15">
                            <template #append>M</template>
                        </el-input>
                        <span class="input-help">{{ $t('php.post_max_size') }}</span>
                    </el-form-item>
                    <el-form-item label="file_uploads" prop="file_uploads">
                        <el-select v-model="form.file_uploads">
                            <el-option :label="$t('website.isOff')" :value="'Off'"></el-option>
                            <el-option :label="$t('website.isOn')" :value="'On'"></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('php.file_uploads') }}</span>
                    </el-form-item>
                    <el-form-item label="upload_max_filesize" prop="upload_max_filesize">
                        <el-input clearable v-model.number="form.upload_max_filesize" maxlength="15">
                            <template #append>M</template>
                        </el-input>
                        <span class="input-help">{{ $t('php.upload_max_filesize') }}</span>
                    </el-form-item>
                    <el-form-item label="max_file_uploads" prop="max_file_uploads">
                        <el-input clearable v-model.number="form.max_file_uploads" maxlength="15"></el-input>
                        <span class="input-help">{{ $t('php.max_file_uploads') }}</span>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="onSaveStart(phpFormRef)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </el-form-item>
                </el-col>
                <el-col :span="1"><br /></el-col>
                <el-col :span="9">
                    <el-form-item label="default_socket_timeout" prop="default_socket_timeout">
                        <el-input clearable v-model.number="form.default_socket_timeout" maxlength="15">
                            <template #append>{{ $t('commons.units.second') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('php.default_socket_timeout') }}</span>
                    </el-form-item>
                    <el-form-item label="error_reporting" prop="error_reporting">
                        <el-input clearable v-model.trim="form.error_reporting"></el-input>
                        <span class="input-help">{{ $t('php.error_reporting') }}</span>
                    </el-form-item>
                    <el-form-item label="display_errors" prop="display_errors">
                        <el-select v-model="form.display_errors">
                            <el-option :label="$t('website.isOff')" :value="'Off'"></el-option>
                            <el-option :label="$t('website.isOn')" :value="'On'"></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('php.display_errors') }}</span>
                    </el-form-item>
                    <el-form-item label="max_input_time" prop="max_input_time">
                        <el-input clearable v-model.number="form.max_input_time" maxlength="15">
                            <template #append>{{ $t('commons.units.second') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('php.max_input_time') }}</span>
                    </el-form-item>
                    <el-form-item label="memory_limit" prop="memory_limit">
                        <el-input clearable v-model.number="form.memory_limit" maxlength="15">
                            <template #append>M</template>
                        </el-input>
                        <span class="input-help">{{ $t('php.memory_limit') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <ConfirmDialog ref="confirmDialogRef" @confirm="submit"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { GetPHPConfig, UpdatePHPConfig } from '@/api/modules/website';
import { checkNumberRange, Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});
const loading = ref(false);
const phpFormRef = ref();
const confirmDialogRef = ref();
let form = reactive({
    short_open_tag: 'Off',
    max_execution_time: 50,
    max_input_time: 50,
    memory_limit: 50,
    post_max_size: 50,
    file_uploads: 'On',
    upload_max_filesize: 50,
    max_file_uploads: 20,
    default_socket_timeout: 50,
    error_reporting: '',
    display_errors: 'On',
});
const variablesRules = reactive({
    max_execution_time: [checkNumberRange(0, 999999999)],
    max_input_time: [checkNumberRange(0, 999999999)],
    memory_limit: [checkNumberRange(0, 999999999)],
    post_max_size: [checkNumberRange(0, 999999999)],
    upload_max_filesize: [checkNumberRange(0, 999999999)],
    max_file_uploads: [checkNumberRange(0, 999999999)],
    default_socket_timeout: [checkNumberRange(0, 999999999)],
    error_reporting: [Rules.requiredInput],
    short_open_tag: [Rules.requiredSelect],
    file_uploads: [Rules.requiredSelect],
    display_errors: [Rules.requiredSelect],
});

const get = () => {
    loading.value = true;
    GetPHPConfig(id.value)
        .then((res) => {
            const param = res.data.params;
            form.short_open_tag = param.short_open_tag;
            form.max_execution_time = Number(param.max_execution_time);
            form.max_input_time = Number(param.max_input_time);
            form.memory_limit = parseFloat(param.memory_limit.replace(/[^\d.]/g, ''));
            form.post_max_size = parseFloat(param.post_max_size.replace(/[^\d.]/g, ''));
            form.file_uploads = param.file_uploads;
            form.upload_max_filesize = parseFloat(param.upload_max_filesize.replace(/[^\d.]/g, ''));
            form.max_file_uploads = Number(param.max_file_uploads);
            form.default_socket_timeout = Number(param.default_socket_timeout);
            form.error_reporting = param.error_reporting;
            form.display_errors = param.display_errors;
        })
        .finally(() => {
            loading.value = false;
        });
};

const onSaveStart = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

const submit = async () => {
    const params = {
        short_open_tag: form.short_open_tag,
        max_execution_time: String(form.max_execution_time),
        max_input_time: String(form.max_input_time),
        memory_limit: form.memory_limit + 'M',
        post_max_size: form.post_max_size + 'M',
        file_uploads: form.file_uploads,
        upload_max_filesize: form.upload_max_filesize + 'M',
        max_file_uploads: String(form.max_file_uploads),
        default_socket_timeout: String(form.default_socket_timeout),
        error_reporting: form.error_reporting,
        display_errors: form.display_errors,
    };
    loading.value = true;
    UpdatePHPConfig({ id: id.value, params: params, scope: 'params' })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    get();
});
</script>
