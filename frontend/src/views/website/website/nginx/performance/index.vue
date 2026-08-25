<template>
    <div>
        <el-form :model="form" :rules="variablesRules" ref="nginxFormRef" label-position="top">
            <el-row v-loading="loading" :gutter="20">
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="server_names_hash_bucket_size" prop="server_names_hash_bucket_size">
                        <el-input
                            clearable
                            type="number"
                            v-model.number="form.server_names_hash_bucket_size"
                        ></el-input>
                        <span class="input-help">{{ $t('nginx.serverNamesHashBucketSizeHelper') }}</span>
                    </el-form-item>
                    <el-form-item label="client_header_buffer_size" prop="client_header_buffer_size">
                        <el-input clearable type="number" v-model.number="form.client_header_buffer_size">
                            <template #append>{{ unitLabel('client_header_buffer_size', 'k') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.clientHeaderBufferSizeHelper') }}</span>
                    </el-form-item>
                    <el-form-item label="client_max_body_size" prop="client_max_body_size">
                        <el-input clearable type="number" v-model.number="form.client_max_body_size">
                            <template #append>{{ unitLabel('client_max_body_size', 'm') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.clientMaxBodySizeHelper') }}</span>
                    </el-form-item>
                    <el-form-item label="keepalive_timeout" prop="keepalive_timeout">
                        <el-input clearable type="number" v-model.number="form.keepalive_timeout"></el-input>
                        <span class="input-help">{{ $t('nginx.keepaliveTimeoutHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="gzip" prop="gzip">
                        <el-select v-model="form.gzip">
                            <el-option :label="'on'" :value="'on'"></el-option>
                            <el-option :label="'off'" :value="'off'"></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('nginx.gzipHelper') }}</span>
                    </el-form-item>
                    <el-form-item label="gzip_min_length" prop="gzip_min_length">
                        <el-input clearable v-model.number="form.gzip_min_length">
                            <template #append>{{ unitLabel('gzip_min_length', 'k') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.gzipMinLengthHelper') }}</span>
                    </el-form-item>
                    <el-form-item label="gzip_comp_level" prop="gzip_comp_level">
                        <el-input clearable v-model.number="form.gzip_comp_level"></el-input>
                        <span class="input-help">{{ $t('nginx.gzipCompLevelHelper') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-form-item>
                <el-button v-permission type="primary" @click="submit(nginxFormRef)">
                    {{ $t('commons.button.save') }}
                </el-button>
            </el-form-item>
        </el-form>
    </div>
</template>
<script lang="ts" setup>
import { Nginx } from '@/api/interface/nginx';
import { getNginxConfigByScope, updateNginxConfigByScope } from '@/api/modules/nginx';
import { checkNumberRange, Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';

let req = ref<Nginx.NginxScopeReq>({
    scope: 'http-per',
});
let updateReq = ref<Nginx.NginxConfigReq>({
    scope: 'http-per',
    operate: 'update',
    params: {},
});
let data = ref();
let form = ref({
    server_names_hash_bucket_size: 512,
    client_header_buffer_size: 32,
    client_max_body_size: 50,
    keepalive_timeout: 60,
    gzip_min_length: 1,
    gzip_comp_level: 2,
    gzip: 'on',
});
let nginxFormRef = ref();
let loading = ref(false);

const variablesRules = reactive({
    server_names_hash_bucket_size: [checkNumberRange(1, 9999)],
    client_header_buffer_size: [checkNumberRange(0, 999999999)],
    client_max_body_size: [checkNumberRange(0, 999999999)],
    keepalive_timeout: [checkNumberRange(0, 999999999)],
    gzip: [Rules.requiredSelect],
    gzip_min_length: [Rules.requiredSelect],
    gzip_comp_level: [checkNumberRange(1, 9)],
});

// nginx size directives may be written with or without a unit suffix, and
// the suffix carries a factor of 1024. Remember the unit that was read so it
// can be written back unchanged, instead of assuming a fixed one.
const sizeKeys = ['client_header_buffer_size', 'client_max_body_size', 'gzip_min_length'];
const units = ref<Record<string, string>>({});

const parseSizeParam = (name: string, value: string) => {
    const matched = /^(\d+)\s*([kKmMgG]?)$/.exec(value.trim());
    if (!matched) {
        return Number(value.match(/\d+/g)?.[0] ?? 0);
    }
    units.value[name] = matched[2] === '' ? '' : matched[2].toLowerCase();
    return Number(matched[1]);
};

const withUnit = (name: string, value: number, defaultUnit: string) => {
    const unit = units.value[name] ?? defaultUnit;
    return String(value) + unit;
};

// Label the input with the unit actually in use, so a value stored in bytes
// is not presented as if it were kilobytes.
const unitLabel = (name: string, defaultUnit: string) => {
    const unit = units.value[name] ?? defaultUnit;
    return unit === '' ? 'B' : unit.toUpperCase();
};

const getParams = async () => {
    const res = await getNginxConfigByScope(req.value);
    data.value = res.data;
    for (const param of res.data) {
        if (param.params.length === 0) {
            continue;
        }
        if (param.name == 'gzip') {
            form.value.gzip = param.params[0];
        } else if (sizeKeys.includes(param.name)) {
            form.value[param.name] = parseSizeParam(param.name, param.params[0]);
        } else {
            form.value[param.name] = Number(param.params[0].match(/\d+/g)?.[0] ?? 0);
        }
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        let params = {
            gzip: form.value.gzip,
            server_names_hash_bucket_size: String(form.value.server_names_hash_bucket_size),
            client_header_buffer_size: withUnit('client_header_buffer_size', form.value.client_header_buffer_size, 'k'),
            client_max_body_size: withUnit('client_max_body_size', form.value.client_max_body_size, 'm'),
            keepalive_timeout: String(form.value.keepalive_timeout),
            gzip_min_length: withUnit('gzip_min_length', form.value.gzip_min_length, 'k'),
            gzip_comp_level: String(form.value.gzip_comp_level),
        };
        updateReq.value.params = params;
        updateNginxConfigByScope(updateReq.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                getParams();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    getParams();
});
</script>
