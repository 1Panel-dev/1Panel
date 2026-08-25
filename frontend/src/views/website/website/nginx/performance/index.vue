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
            <el-row v-if="brotliAvailable" v-loading="loading" :gutter="20">
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="brotli" prop="brotli">
                        <el-select v-model="brotliForm.brotli">
                            <el-option :label="'on'" :value="'on'"></el-option>
                            <el-option :label="'off'" :value="'off'"></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('nginx.brotliHelper') }}</span>
                    </el-form-item>
                    <el-form-item label="brotli_min_length" prop="brotli_min_length">
                        <el-input clearable v-model.number="brotliForm.brotli_min_length">
                            <template #append>{{ unitLabel('brotli_min_length', 'k') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.gzipMinLengthHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="brotli_comp_level" prop="brotli_comp_level">
                        <el-input clearable v-model.number="brotliForm.brotli_comp_level"></el-input>
                        <span class="input-help">{{ $t('nginx.brotliCompLevelHelper') }}</span>
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
import { getNginxConfigByScope, getNginxModules, updateNginxConfigByScope } from '@/api/modules/nginx';
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

// Brotli is only offered once the module is enabled and built. Its directives
// live in a panel-managed file rather than nginx.conf, because they have to
// disappear together with the module.
const brotliAvailable = ref(false);
const brotliForm = ref({
    brotli: 'on',
    brotli_comp_level: 5,
    brotli_min_length: 1,
});

const variablesRules = reactive({
    server_names_hash_bucket_size: [checkNumberRange(1, 9999)],
    client_header_buffer_size: [checkNumberRange(0, 999999999)],
    client_max_body_size: [checkNumberRange(0, 999999999)],
    keepalive_timeout: [checkNumberRange(0, 999999999)],
    gzip: [Rules.requiredSelect],
    gzip_min_length: [Rules.requiredSelect],
    gzip_comp_level: [checkNumberRange(1, 9)],
    brotli: [Rules.requiredSelect],
    brotli_min_length: [Rules.requiredSelect],
    brotli_comp_level: [checkNumberRange(0, 11)],
});

// nginx size directives may be written with or without a unit suffix, and
// the suffix carries a factor of 1024. Remember the unit that was read so it
// can be written back unchanged, instead of assuming a fixed one.
const sizeKeys = ['client_header_buffer_size', 'client_max_body_size', 'gzip_min_length', 'brotli_min_length'];
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
    await getBrotliParams();
};

const getBrotliParams = async () => {
    const modules = await getNginxModules();
    const brotli = modules.data.modules?.find((item) => item.name === 'ngx_brotli');
    brotliAvailable.value = !!brotli && brotli.enable && brotli.buildStatus === 'ready';
    if (!brotliAvailable.value) {
        return;
    }
    const res = await getNginxConfigByScope({ scope: 'brotli' });
    for (const param of res.data) {
        if (param.params.length === 0) {
            continue;
        }
        if (param.name === 'brotli') {
            brotliForm.value.brotli = param.params[0];
        } else if (param.name === 'brotli_min_length') {
            brotliForm.value.brotli_min_length = parseSizeParam(param.name, param.params[0]);
        } else if (param.name === 'brotli_comp_level') {
            brotliForm.value.brotli_comp_level = Number(param.params[0].match(/\d+/g)?.[0] ?? 0);
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
                if (!brotliAvailable.value) {
                    return;
                }
                // brotli_types is managed by the panel and deliberately not
                // exposed, so it stays aligned with gzip_types.
                return updateNginxConfigByScope({
                    scope: 'brotli',
                    operate: 'update',
                    params: {
                        brotli: brotliForm.value.brotli,
                        brotli_comp_level: String(brotliForm.value.brotli_comp_level),
                        brotli_min_length: withUnit('brotli_min_length', brotliForm.value.brotli_min_length, 'k'),
                    },
                });
            })
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
