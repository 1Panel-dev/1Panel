<template>
    <div v-for="(p, index) in paramObjs" :key="index">
        <el-form-item :label="p.labelZh" :prop="p.prop">
            <el-input
                v-model.trim="form[p.envKey]"
                v-if="p.type == 'text'"
                :type="p.type"
                @change="updateParam"
            ></el-input>
            <el-input
                v-model.number="form[p.envKey]"
                v-if="p.type == 'number'"
                :type="p.type"
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
            <span v-if="p.type === 'service' && !p.services" style="margin-left: 5px">
                <el-link type="primary" :underline="false" @click="toPage()">{{ $t('app.toInstall') }}</el-link>
            </span>
        </el-form-item>
    </div>
</template>
<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { getRandomStr } from '@/utils/util';
import { GetAppService } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import { App } from '@/api/interface/app';
import { useRouter } from 'vue-router';
const router = useRouter();

interface ParamObj extends App.FromField {
    services: App.AppService[];
    prop: string;
    disabled: false;
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
            if (p.default == 'random') {
                form[p.envKey] = getRandomStr(6);
            } else {
                form[p.envKey] = p.default;
            }
            if (p.required) {
                if (p.type === 'service') {
                    rules[p.envKey] = [Rules.requiredSelect];
                } else {
                    rules[p.envKey] = [Rules.requiredInput];
                    if (p.envKey === 'PANEL_DB_NAME') {
                        rules[p.envKey].push(Rules.dbName);
                    }
                }
            }
            if (p.key) {
                form[p.envKey] = '';
                getServices(p.envKey, p.key, pObj);
            }
            emit('update:rules', rules);
            updateParam();
        }
    }
};

const getServices = async (envKey: string, key: string | undefined, pObj: ParamObj) => {
    await GetAppService(key).then((res) => {
        pObj.services = res.data;
        if (res.data && res.data.length > 0) {
            form[envKey] = res.data[0].value;
            if (res.data[0].config) {
                Object.entries(res.data[0].config).forEach(([k, v]) => {
                    params.value.formFields.forEach((field) => {
                        if (field.envKey === k) {
                            form[k] = v;
                        }
                    });
                });
            }
            updateParam();
        }
    });
};

const changeService = (value: string, services: App.AppService[]) => {
    services.forEach((item) => {
        if (item.value === value) {
            Object.entries(item.config).forEach(([k, v]) => {
                if (form.hasOwnProperty(k)) {
                    form[k] = v;
                }
            });
        }
    });
    updateParam();
};

const toPage = () => {
    router.push({ name: 'App' });
};

onMounted(() => {
    handleParams();
});
</script>
