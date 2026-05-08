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
            <div v-if="p.type == 'service'" class="service-param-row">
                <el-select
                    class="service-param-select p-w-200"
                    v-model="form[p.envKey]"
                    @change="changeService(form[p.envKey], p.services)"
                >
                    <el-option
                        v-for="service in p.services"
                        :key="service.label"
                        :value="service.value"
                        :label="service.label"
                    ></el-option>
                </el-select>
                <span v-if="p.services.length === 0" class="service-install-link">
                    <el-link class="service-install-link__text" type="primary" underline="never" @click="toPage(p.key)">
                        {{ $t('app.toInstall') }}
                    </el-link>
                </span>
            </div>
            <el-select
                v-model="form[p.envKey]"
                v-if="p.type == 'select'"
                :multiple="p.multiple"
                :allowCreate="p.allowCreate"
                filterable
            >
                <el-option
                    v-for="service in p.values"
                    :key="service.label"
                    :value="service.value"
                    :label="service.label"
                ></el-option>
            </el-select>
            <div v-if="p.type == 'apps'" class="app-service-row">
                <div class="app-service-row__app">
                    <el-form-item :prop="p.prop">
                        <el-select
                            v-model="form[p.envKey]"
                            @change="getServices(p.child.envKey, form[p.envKey], p)"
                            class="app-service-row__select p-w-200"
                        >
                            <el-option
                                v-for="service in p.values"
                                :label="service.label"
                                :key="service.value"
                                :value="service.value"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                </div>
                <div class="app-service-row__service">
                    <el-form-item :prop="p.childProp">
                        <el-select
                            v-model="form[p.child.envKey]"
                            v-if="p.child.type == 'service'"
                            @change="changeService(form[p.child.envKey], p.services)"
                            class="app-service-row__select p-w-300"
                        >
                            <el-option
                                v-for="service in p.services"
                                :key="service.label"
                                :value="service.value"
                                :label="service.label"
                                :disabled="service.status != 'Running'"
                            >
                                <el-row :gutter="5">
                                    <el-col :span="14">
                                        <span>{{ service.label }}</span>
                                    </el-col>
                                    <el-col :span="6">
                                        <span v-if="service.from != ''">
                                            <el-tag v-if="service.from === 'local'">
                                                {{ $t('commons.table.local') }}
                                            </el-tag>
                                            <el-tag v-else type="success">{{ $t('database.remote') }}</el-tag>
                                            <Status
                                                class="ml-2"
                                                :key="service.status"
                                                :status="service.status"
                                            ></Status>
                                        </span>
                                    </el-col>
                                </el-row>
                            </el-option>
                        </el-select>
                    </el-form-item>
                </div>
                <span v-if="p.child.type === 'service' && p.services.length === 0" class="service-install-link">
                    <el-link
                        class="service-install-link__text"
                        type="primary"
                        underline="never"
                        @click="toPage(form[p.envKey])"
                    >
                        {{ $t('app.toInstall') }}
                    </el-link>
                </span>
            </div>
            <span class="input-help" v-if="p.description">{{ getDescription(p) }}</span>
        </el-form-item>
        <el-form-item v-if="isMysql(form, p.envKey)" :label="$t('database.format')" prop="format">
            <el-select filterable v-model="form.format" @change="loadCollations()">
                <el-option v-for="item of formatOptions" :key="item.format" :label="item.format" :value="item.format" />
            </el-select>
        </el-form-item>
        <el-form-item v-if="isMysql(form, p.envKey)" :label="$t('database.collation')" prop="collation">
            <el-select filterable v-model="form.collation">
                <el-option v-for="item of collationOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <span class="input-help">{{ $t('database.collationHelper', [form.format]) }}</span>
        </el-form-item>
    </div>
</template>
<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { getRandomStr } from '@/utils/id';
import { getAppService } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import { App } from '@/api/interface/app';
import { getDBName, getLabel, getDescription } from '@/utils/app-store';
import { loadWebsiteDir } from '@/api/modules/setting';
import { loadFormatCollations } from '@/api/modules/database';

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

const form = reactive({
    format: '',
    collation: '',
});
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

const isMysql = (form: object, envKey: string) => {
    return form['PANEL_DB_HOST'] != undefined && (form[envKey] == 'mysql' || form[envKey] == 'mariadb');
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
            if (p.type == 'text' && p.envKey == 'WEBSITE_DIR') {
                loadWebsiteDir().then((res) => {
                    form[p.envKey] = res.data;
                });
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
    appKey.value = key || '';
    if (appKey.value == 'mysql' || appKey.value == 'mariadb') {
        form.format = 'utf8mb4';
    }
    await getAppService(key).then((res) => {
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
    if (appKey.value == 'mysql' || appKey.value == 'mariadb') {
        loadOptions(value);
    }
    updateParam();
};

const toPage = (key: string) => {
    window.location.href = '/apps/all?install=' + key;
};

const formatOptions = ref();
const collationOptions = ref();
const appKey = ref('');

const loadOptions = async (database: string) => {
    const defaultOptions = [{ format: 'utf8mb4' }, { format: 'utf8mb3' }, { format: 'gbk' }, { format: 'big5' }];
    await loadFormatCollations(database).then((res) => {
        formatOptions.value = res.data || defaultOptions;
        loadCollations();
    });
};

const loadCollations = async () => {
    collationOptions.value = formatOptions.value?.find((item) => item.format === form.format)?.collations || [];
};

onMounted(() => {
    handleParams();
});
</script>

<style lang="scss" scoped>
.service-param-row,
.app-service-row {
    display: inline-flex;
    align-items: center;
    gap: 12px;
    max-width: 100%;
    flex-wrap: nowrap;
}

.service-param-select,
.app-service-row__app,
.app-service-row__service {
    min-width: 0;
}

.service-param-select {
    flex: 0 1 200px;
}

.app-service-row__app {
    flex: 0 1 200px;
}

.app-service-row__service {
    flex: 0 1 300px;
}

.app-service-row__select {
    width: 100%;
}

.app-service-row__app :deep(.el-form-item),
.app-service-row__service :deep(.el-form-item) {
    margin-bottom: 0;
}

.app-service-row__app :deep(.el-form-item__content),
.app-service-row__service :deep(.el-form-item__content) {
    width: 100%;
}

.service-install-link {
    flex: 0 0 auto;
}

.service-install-link__text {
    white-space: nowrap;
}
</style>
