<template>
    <div v-for="(p, index) in paramObjs" :key="index">
        <el-form-item :label="getLabel(p)" :prop="p.prop">
            <el-select
                v-model="form[p.envKey]"
                v-if="p.type == 'select'"
                :multiple="p.multiple"
                filterable
                allow-create
                default-first-option
                @change="updateParam"
            >
                <el-option
                    v-for="service in p.values"
                    :key="service.label"
                    :value="service.value"
                    :label="service.label"
                ></el-option>
            </el-select>
        </el-form-item>
    </div>
</template>

<script setup lang="ts">
import { App } from '@/api/interface/app';
import { Rules } from '@/global/form-rules';
import { getLanguage } from '@/utils/util';
import { computed, onMounted, reactive, ref } from 'vue';

interface ParamObj extends App.FromField {
    services: App.AppService[];
    prop: string;
    disabled: false;
    childProp: string;
}

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
});

let form = reactive({});
let rules = reactive({});
const params = computed({
    get() {
        return props.params;
    },
    set() {},
});
const emit = defineEmits(['update:form', 'update:rules']);
const updateParam = () => {
    emit('update:form', form);
};
const paramObjs = ref<ParamObj[]>([]);

const handleParams = () => {
    rules = props.rules;
    if (params.value != undefined && params.value.formFields != undefined) {
        for (const p of params.value.formFields) {
            const pObj = p;
            pObj.prop = p.envKey;
            pObj.disabled = p.disabled;
            form[p.envKey] = '';
            paramObjs.value.push(pObj);
            if (p.required) {
                if (p.type === 'select') {
                    rules[p.envKey] = [Rules.requiredSelect];
                } else {
                    rules[p.envKey] = [Rules.requiredInput];
                }
            }
            form[p.envKey] = p.default;
        }
        emit('update:rules', rules);
        updateParam();
    }
};

const getLabel = (row: ParamObj): string => {
    const language = getLanguage();
    if (language == 'zh' || language == 'tw') {
        return row.labelZh;
    } else {
        return row.labelEn;
    }
};

onMounted(() => {
    handleParams();
});
</script>
