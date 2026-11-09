<template>
    <div>
        <el-form v-loading="loading" :model="form" :rules="variablesRules" ref="nginxFormRef" label-position="top">
            <el-row :gutter="20">
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="server_names_hash_bucket_size" prop="server_names_hash_bucket_size">
                        <el-input
                            clearable
                            type="number"
                            v-model.number="form.server_names_hash_bucket_size"
                            :disabled="!!sources.server_names_hash_bucket_size"
                        ></el-input>
                        <span class="input-help">{{ $t('nginx.serverNamesHashBucketSizeHelper') }}</span>
                        <span v-if="sources.server_names_hash_bucket_size" class="input-help">
                            {{ sourceHelp('server_names_hash_bucket_size') }}
                        </span>
                    </el-form-item>
                    <el-form-item label="client_header_buffer_size" prop="client_header_buffer_size">
                        <el-input
                            clearable
                            type="number"
                            v-model.number="form.client_header_buffer_size"
                            :disabled="!!sources.client_header_buffer_size"
                        >
                            <template #append>{{ unitLabel('client_header_buffer_size', 'k') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.clientHeaderBufferSizeHelper') }}</span>
                        <span v-if="sources.client_header_buffer_size" class="input-help">
                            {{ sourceHelp('client_header_buffer_size') }}
                        </span>
                    </el-form-item>
                    <el-form-item label="client_max_body_size" prop="client_max_body_size">
                        <el-input
                            clearable
                            type="number"
                            v-model.number="form.client_max_body_size"
                            :disabled="!!sources.client_max_body_size"
                        >
                            <template #append>{{ unitLabel('client_max_body_size', 'm') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.clientMaxBodySizeHelper') }}</span>
                        <span v-if="sources.client_max_body_size" class="input-help">
                            {{ sourceHelp('client_max_body_size') }}
                        </span>
                    </el-form-item>
                    <el-form-item label="keepalive_timeout" prop="keepalive_timeout">
                        <el-input
                            clearable
                            v-model.trim="form.keepalive_timeout"
                            :disabled="!!sources.keepalive_timeout"
                        ></el-input>
                        <span class="input-help">{{ $t('nginx.keepaliveTimeoutHelper') }}</span>
                        <span v-if="sources.keepalive_timeout" class="input-help">
                            {{ sourceHelp('keepalive_timeout') }}
                        </span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="gzip" prop="gzip">
                        <el-select v-model="form.gzip" :disabled="!!sources.gzip">
                            <el-option :label="'on'" :value="'on'"></el-option>
                            <el-option :label="'off'" :value="'off'"></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('nginx.gzipHelper') }}</span>
                        <span v-if="sources.gzip" class="input-help">{{ sourceHelp('gzip') }}</span>
                    </el-form-item>
                    <el-form-item label="gzip_min_length" prop="gzip_min_length">
                        <el-input clearable v-model.number="form.gzip_min_length" :disabled="!!sources.gzip_min_length">
                            <template #append>{{ unitLabel('gzip_min_length', 'k') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.gzipMinLengthHelper') }}</span>
                        <span v-if="sources.gzip_min_length" class="input-help">
                            {{ sourceHelp('gzip_min_length') }}
                        </span>
                    </el-form-item>
                    <el-form-item label="gzip_comp_level" prop="gzip_comp_level">
                        <el-input
                            clearable
                            v-model.number="form.gzip_comp_level"
                            :disabled="!!sources.gzip_comp_level"
                        ></el-input>
                        <span class="input-help">{{ $t('nginx.gzipCompLevelHelper') }}</span>
                        <span v-if="sources.gzip_comp_level" class="input-help">
                            {{ sourceHelp('gzip_comp_level') }}
                        </span>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row :gutter="20">
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="proxy_connect_timeout" prop="proxy_connect_timeout">
                        <el-input
                            clearable
                            v-model.trim="form.proxy_connect_timeout"
                            :disabled="!!sources.proxy_connect_timeout"
                        ></el-input>
                        <span class="input-help">{{ $t('nginx.proxyConnectTimeoutHelper') }}</span>
                        <span v-if="sources.proxy_connect_timeout" class="input-help">
                            {{ sourceHelp('proxy_connect_timeout') }}
                        </span>
                    </el-form-item>
                    <el-form-item label="proxy_send_timeout" prop="proxy_send_timeout">
                        <el-input
                            clearable
                            v-model.trim="form.proxy_send_timeout"
                            :disabled="!!sources.proxy_send_timeout"
                        ></el-input>
                        <span class="input-help">{{ $t('nginx.proxySendTimeoutHelper') }}</span>
                        <span v-if="sources.proxy_send_timeout" class="input-help">
                            {{ sourceHelp('proxy_send_timeout') }}
                        </span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="proxy_read_timeout" prop="proxy_read_timeout">
                        <el-input
                            clearable
                            v-model.trim="form.proxy_read_timeout"
                            :disabled="!!sources.proxy_read_timeout"
                        ></el-input>
                        <span class="input-help">{{ $t('nginx.proxyReadTimeoutHelper') }}</span>
                        <span v-if="sources.proxy_read_timeout" class="input-help">
                            {{ sourceHelp('proxy_read_timeout') }}
                        </span>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row v-if="brotliAvailable" :gutter="20">
                <el-col :span="24" v-if="brotliManagedExternally">
                    <el-alert type="info" :closable="false" class="mb-2">
                        {{ $t('nginx.brotliManagedExternallyHelper') }}
                    </el-alert>
                </el-col>
                <el-col :span="24" v-if="brotliManagedPending">
                    <el-alert type="info" :closable="false" class="mb-2">
                        {{ $t('nginx.brotliManagedPendingHelper') }}
                    </el-alert>
                </el-col>
                <el-col :span="24" v-if="brotliManagedUnavailable">
                    <el-alert type="warning" :closable="false" class="mb-2">
                        {{ $t('nginx.brotliManagedUnavailableHelper') }}
                    </el-alert>
                </el-col>
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="brotli" prop="brotli">
                        <el-select v-model="form.brotli">
                            <el-option :label="'on'" :value="'on'"></el-option>
                            <el-option :label="'off'" :value="'off'"></el-option>
                        </el-select>
                        <span class="input-help">{{ $t('nginx.brotliHelper') }}</span>
                    </el-form-item>
                    <el-form-item label="brotli_min_length" prop="brotli_min_length">
                        <el-input clearable v-model.number="form.brotli_min_length">
                            <template #append>{{ unitLabel('brotli_min_length', 'k') }}</template>
                        </el-input>
                        <span class="input-help">{{ $t('nginx.brotliMinLengthHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                    <el-form-item label="brotli_comp_level" prop="brotli_comp_level">
                        <el-input clearable v-model.number="form.brotli_comp_level"></el-input>
                        <span class="input-help">{{ $t('nginx.brotliCompLevelHelper') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-form-item>
                <el-button v-permission type="primary" :disabled="!loaded" @click="submit(nginxFormRef)">
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
import { MsgError, MsgSuccess } from '@/utils/message';
import { FormInstance, FormItemRule } from 'element-plus';
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
    keepalive_timeout: '75s',
    gzip_min_length: 1,
    gzip_comp_level: 2,
    gzip: 'on',
    brotli: 'on',
    brotli_comp_level: 5,
    brotli_min_length: 1,
    proxy_connect_timeout: '60s',
    proxy_send_timeout: '60s',
    proxy_read_timeout: '60s',
});
let nginxFormRef = ref();
let loading = ref(false);
// Saving waits for the current values, so the form defaults are never
// written over the real configuration.
const loaded = ref(false);

// Files that set a directive nginx.conf leaves unset, by directive name. Such
// a field is read-only and never saved: writing it into nginx.conf would
// duplicate the included value and fail the config test.
const sources = ref<Record<string, string>>({});
const sourceHelp = (name: string) => i18n.global.t('nginx.directiveSourceHelper', [sources.value[name]]);

// Brotli is only offered once the module is enabled and built. Its directives
// live in a panel-managed file rather than nginx.conf, because they have to
// disappear together with the module.
const brotliAvailable = ref(false);
const brotliManagedExternally = ref(false);
// Pending: nginx does not load the managed file yet, and saving wires it in.
const brotliManagedPending = ref(false);
const brotliManagedUnavailable = ref(false);

// Durations are edited as written, so a value such as 1m30s round-trips
// intact. An unchanged one is not sent at all: it may be a built-in default
// that nginx.conf does not set.
const timeKeys = ['keepalive_timeout', 'proxy_connect_timeout', 'proxy_send_timeout', 'proxy_read_timeout'];
// nginx's built-in defaults, shown when nginx.conf does not set the directive.
const timeDefaults: Record<string, string> = {
    keepalive_timeout: '75s',
    proxy_connect_timeout: '60s',
    proxy_send_timeout: '60s',
    proxy_read_timeout: '60s',
};
const baseline: Record<string, string> = { ...timeDefaults };

const normalizeTime = (value: unknown) =>
    String(value ?? '')
        .trim()
        .split(/\s+/)
        .join(' ');

// Mirrors ngx_parse_time: units run from the most to the least significant,
// and a trailing number without a unit means seconds. Months and years are
// rejected for millisecond settings, milliseconds for whole-second ones.
const timeUnits = ['y', 'M', 'w', 'd', 'h', 'm', 's', 'ms'];
const timeUnitMs = [31536000000, 2592000000, 604800000, 86400000, 3600000, 60000, 1000, 1];

const parseNginxTime = (value: string, wholeSeconds: boolean): number | null => {
    const tokens = value.match(/\d+(ms|[yMwdhms])?/g);
    if (!tokens || tokens.join('') !== value) {
        return null;
    }
    let total = 0;
    let previous = -1;
    for (const token of tokens) {
        const unit = token.replace(/^\d+/, '');
        const digits = token.slice(0, token.length - unit.length);
        if (unit === '') {
            total += Number(digits) * 1000;
            continue;
        }
        const index = timeUnits.indexOf(unit);
        if (index <= previous || (wholeSeconds ? unit === 'ms' : index < 2)) {
            return null;
        }
        previous = index;
        total += Number(digits) * timeUnitMs[index];
    }
    return total;
};

const isValidTime = (name: string, value: string) => {
    const parts = value.split(' ');
    if (name === 'keepalive_timeout') {
        // The optional second value is the Keep-Alive response header, in seconds.
        return (
            parts.length <= 2 &&
            parseNginxTime(parts[0], false) !== null &&
            (parts.length === 1 || parseNginxTime(parts[1], true) !== null)
        );
    }
    // Zero would time every request out at once rather than lift the limit.
    return parts.length === 1 && (parseNginxTime(parts[0], false) ?? 0) > 0;
};

const checkTime = (name: string): FormItemRule => ({
    trigger: 'blur',
    validator: (_rule, value, callback) => {
        const text = normalizeTime(value);
        if (text === baseline[name] || isValidTime(name, text)) {
            callback();
        } else {
            callback(new Error(i18n.global.t('nginx.timeValueError')));
        }
    },
});

const variablesRules = reactive({
    server_names_hash_bucket_size: [checkNumberRange(1, 9999)],
    client_header_buffer_size: [checkNumberRange(0, 999999999)],
    client_max_body_size: [checkNumberRange(0, 999999999)],
    keepalive_timeout: [Rules.requiredInput, checkTime('keepalive_timeout')],
    gzip: [Rules.requiredSelect],
    gzip_min_length: [Rules.requiredSelect],
    gzip_comp_level: [checkNumberRange(1, 9)],
    brotli: [Rules.requiredSelect],
    brotli_min_length: [Rules.requiredSelect],
    brotli_comp_level: [checkNumberRange(0, 11)],
    proxy_connect_timeout: [Rules.requiredInput, checkTime('proxy_connect_timeout')],
    proxy_send_timeout: [Rules.requiredInput, checkTime('proxy_send_timeout')],
    proxy_read_timeout: [Rules.requiredInput, checkTime('proxy_read_timeout')],
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
    loaded.value = false;
    // Reset the unit cache: a directive removed since the last load must not
    // hand its remembered unit to a value that no longer carries one.
    units.value = {};
    const res = await getNginxConfigByScope(req.value);
    data.value = res.data;
    const nextSources: Record<string, string> = {};
    for (const param of res.data) {
        if (param.source) {
            nextSources[param.name] = param.source;
        }
        if (timeKeys.includes(param.name)) {
            const value = param.params.length > 0 ? param.params.join(' ') : timeDefaults[param.name];
            form.value[param.name] = value;
            baseline[param.name] = value;
            continue;
        }
        if (param.params.length === 0) {
            continue;
        }
        if (param.name == 'gzip') {
            form.value.gzip = param.params[0];
        } else if (sizeKeys.includes(param.name)) {
            form.value[param.name] = parseSizeParam(param.name, param.params[0]);
        } else if (param.name in form.value) {
            form.value[param.name] = Number(param.params[0].match(/\d+/g)?.[0] ?? 0);
        }
        // Keys the page does not render (gzip_types, gzip_vary, gzip_proxied)
        // are deliberately not stuffed into the form.
    }
    sources.value = nextSources;
    await getBrotliParams();
    loaded.value = true;
};

const getBrotliParams = async () => {
    const modules = await getNginxModules();
    const brotli = modules.data.modules?.find((item) => item.name === 'ngx_brotli');
    brotliAvailable.value = !!brotli && brotli.enable && brotli.buildStatus === 'ready';
    if (!brotliAvailable.value) {
        return;
    }
    const res = await getNginxConfigByScope<Nginx.NginxBrotliRes>({ scope: 'brotli' });
    brotliManagedExternally.value = !!res.data?.managedExternally;
    brotliManagedPending.value = !!res.data?.managedPending;
    brotliManagedUnavailable.value = !!res.data?.managedUnavailable;
    for (const param of res.data?.params ?? []) {
        if (param.params.length === 0) {
            continue;
        }
        if (param.name === 'brotli') {
            form.value.brotli = param.params[0];
        } else if (param.name === 'brotli_min_length') {
            form.value.brotli_min_length = parseSizeParam(param.name, param.params[0]);
        } else if (param.name === 'brotli_comp_level') {
            form.value.brotli_comp_level = Number(param.params[0].match(/\d+/g)?.[0] ?? 0);
        }
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl || !loaded.value) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        let params: Record<string, string> = {
            gzip: form.value.gzip,
            server_names_hash_bucket_size: String(form.value.server_names_hash_bucket_size),
            client_header_buffer_size: withUnit('client_header_buffer_size', form.value.client_header_buffer_size, 'k'),
            client_max_body_size: withUnit('client_max_body_size', form.value.client_max_body_size, 'm'),
            gzip_min_length: withUnit('gzip_min_length', form.value.gzip_min_length, 'k'),
            gzip_comp_level: String(form.value.gzip_comp_level),
        };
        for (const name of timeKeys) {
            const value = normalizeTime(form.value[name]);
            if (value !== baseline[name]) {
                params[name] = value;
            }
        }
        for (const name of Object.keys(sources.value)) {
            delete params[name];
        }
        updateReq.value.params = params;
        updateNginxConfigByScope(updateReq.value)
            .then(() => {
                if (!brotliAvailable.value || brotliManagedUnavailable.value) {
                    return;
                }
                // brotli_types is managed by the panel and deliberately not
                // exposed, so it stays aligned with gzip_types.
                return updateNginxConfigByScope({
                    scope: 'brotli',
                    operate: 'update',
                    params: {
                        brotli: form.value.brotli,
                        brotli_comp_level: String(form.value.brotli_comp_level),
                        brotli_min_length: withUnit('brotli_min_length', form.value.brotli_min_length, 'k'),
                    },
                }).catch(() => {
                    // gzip was already saved by the time brotli failed; say so
                    // instead of leaving the user guessing which half applied.
                    MsgError(i18n.global.t('nginx.brotliSaveFailed'));
                    getParams();
                    throw new Error('brotli save failed');
                });
            })
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                getParams();
            })
            .catch(() => {})
            .finally(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    getParams();
});
</script>
