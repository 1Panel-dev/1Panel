<template>
    <div v-for="(p, index) in paramObjs" :key="index">
        <el-form-item :label="getLabel(p)" :prop="p.prop">
            <el-input
                v-model.trim="form[p.envKey]"
                v-if="p.type == 'text'"
                :type="p.type"
                @change="updateParam"
                :disabled="p.disabled"
            ></el-input>
            <el-input
                v-model.number="form[p.envKey]"
                @blur="form[p.envKey] = Number(form[p.envKey])"
                v-if="p.type == 'number'"
                maxlength="15"
                @change="updateParam"
                :disabled="p.disabled"
            ></el-input>
            <el-input
                v-model.trim="form[p.envKey]"
                v-if="p.type == 'password'"
                :type="p.type"
                show-password
                clearable
                @change="updateParam"
            ></el-input>
            <el-select
                class="p-w-200"
                v-model="form[p.envKey]"
                v-if="p.type == 'service'"
                @change="changeService(form[p.envKey], p.services)"
            >
                <el-option
                    v-for="service in p.services"
                    :key="service.label"
                    :value="service.value"
                    :label="service.label"
                ></el-option>
            </el-select>
            <span v-if="p.type === 'service' && p.services.length === 0" class="ml-1.5">
                <el-link type="primary" :underline="false" @click="toPage(p.key)">
                    {{ $t('app.toInstall') }}
                </el-link>
            </span>
            <el-select v-model="form[p.envKey]" v-if="p.type == 'select'" :multiple="p.multiple" class="p-w-200">
                <el-option
                    v-for="service in p.values"
                    :key="service.label"
                    :value="service.value"
                    :label="service.label"
                ></el-option>
            </el-select>
            <el-row :gutter="10" v-if="p.type == 'apps'">
                <el-col :span="12">
                    <el-form-item :prop="p.prop">
                        <el-select
                            v-model="form[p.envKey]"
                            @change="getServices(p.child.envKey, form[p.envKey], p)"
                            class="p-w-200"
                        >
                            <el-option
                                v-for="service in p.values"
                                :label="service.label"
                                :key="service.value"
                                :value="service.value"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col :span="12">
                    <el-form-item :prop="p.childProp">
                        <el-select
                            v-model="form[p.child.envKey]"
                            v-if="p.child.type == 'service'"
                            @change="changeService(form[p.child.envKey], p.services)"
                            class="p-w-200"
                        >
                            <el-option
                                v-for="service in p.services"
                                :key="service.label"
                                :value="service.value"
                                :label="service.label"
                            >
                                <span>{{ service.label }}</span>
                                <span class="float-right" v-if="service.from != ''">
                                    <el-tag v-if="service.from === 'local'">{{ $t('database.local') }}</el-tag>
                                    <el-tag v-else type="success">{{ $t('database.remote') }}</el-tag>
                                </span>
                            </el-option>
                        </el-select>
                    </el-form-item>
                </el-col>
                <el-col>
                    <span v-if="p.child.type === 'service' && p.services.length === 0">
                        <el-link type="primary" :underline="false" @click="toPage(form[p.envKey])">
                            {{ $t('app.toInstall') }}
                        </el-link>
                    </span>
                </el-col>
            </el-row>
        </el-form-item>
    </div>
</template>
<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { getRandomStr } from '@/utils/util';
import { GetAppService } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import { App } from '@/api/interface/app';
import { getDBName } from '@/utils/util';

interface ParamObj extends App.FromField {
    services: App.AppService[];
    prop: string;
    disabled: false;
    childProp: string;
}

const emit = defineEmits(['update:form', 'update:rules']);

const props = defineProps({
    form: {
        type: Object,
        default: function () {
            return {};
        },
    },
    params: {
        type: Object,
        default: function () {
            return {};
        },
    },
    rules: {
        type: Object,
        default: function () {
            return {};
        },
    },
    propStart: {
        type: String,
        default: '',
    },
});

const form = reactive({});
let rules = reactive({});
const params = computed({
    get() {
        return props.params;
    },
    set() {},
});
const propStart = computed({
    get() {
        return props.propStart;
    },
    set() {},
});
const paramObjs = ref<ParamObj[]>([]);

const updateParam = () => {
    emit('update:form', form);
};

const handleParams = () => {
    rules = props.rules;
    if (params.value != undefined && params.value.formFields != undefined) {
        for (const p of params.value.formFields) {
            const pObj = p;
            pObj.prop = propStart.value + p.envKey;
            pObj.disabled = p.disabled;
            paramObjs.value.push(pObj);
            if (p.random) {
                if (p.envKey === 'PANEL_DB_NAME') {
                    form[p.envKey] = p.default + '_' + getDBName(6);
                } else {
                    form[p.envKey] = p.default + '_' + getRandomStr(6);
                }
            } else {
                form[p.envKey] = p.default;
            }
            if (p.required) {
                if (p.type === 'service' || p.type === 'apps') {
                    rules[p.envKey] = [Rules.requiredSelect];
                    if (p.child) {
                        p.childProp = propStart.value + p.child.envKey;
                        if (p.child.type === 'service') {
                            rules[p.child.envKey] = [Rules.requiredSelect];
                        }
                    }
                } else {
                    rules[p.envKey] = [Rules.requiredInput];
                }
                if (p.rule && p.rule != '') {
                    rules[p.envKey].push(Rules[p.rule]);
                }
            } else {
                delete rules[p.envKey];
            }
            if (p.type === 'apps') {
                getServices(p.child.envKey, p.default, p);
                p.child.services = [];
                form[p.child.envKey] = '';
            }
            if (p.type === 'service') {
                getServices(p.envKey, p.key, p);
                p.services = [];
                form[p.envKey] = '';
            }
            emit('update:rules', rules);
            updateParam();
        }
    }
};

const getServices = async (childKey: string, key: string | undefined, pObj: ParamObj | undefined) => {
    pObj.services = [];
    await GetAppService(key).then((res) => {
        pObj.services = res.data || [];
        form[childKey] = '';
        if (res.data && res.data.length > 0) {
            form[childKey] = res.data[0].value;
            if (pObj.params) {
                pObj.params.forEach((param: App.FromParam) => {
                    if (param.key === key) {
                        form[param.envKey] = param.value;
                    }
                });
            }
            changeService(form[childKey], pObj.services);
        }
    });
};

const changeService = (value: string, services: App.AppService[]) => {
    services.forEach((item) => {
        if (item.value === value && item.config) {
            Object.entries(item.config).forEach(([k, v]) => {
                if (form.hasOwnProperty(k)) {
                    form[k] = v;
                }
            });
        }
    });
    updateParam();
};

const getLabel = (row: ParamObj): string => {
    const language = localStorage.getItem('lang') || 'zh';
    let lang = language == 'tw' ? 'zh-Hant' : language;
    if (row.label && row.label[lang] != '') {
        return row.label[lang];
    }
    if (language == 'zh' || language == 'tw') {
        return row.labelZh;
    } else {
        return row.labelEn;
    }
};

const toPage = (key: string) => {
    window.location.href = '/apps/all?install=' + key;
};

onMounted(() => {
    handleParams();
});
</script>
