<template>
    <div v-for="(p, index) in paramObjs" :key="index">
        <el-form-item :label="getLabel(p)" :prop="p.prop">
            <el-select
                v-model="form[p.key]"
                v-if="p.type == 'select'"
                :multiple="p.multiple"
                filterable
                allow-create
                default-first-option
                @change="updateParam"
                clearable
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

interface ParamObj extends App.InstallParams {
    prop: string;
}
const props = defineProps({
    form: {
        type: Object,
        default: function () {
            return {};
        },
    },
    params: {
        type: Array<App.InstallParams>,
        default: function () {
            return [];
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
    if (params.value != undefined) {
        for (const p of params.value) {
            form[p.key] = p.value;
            paramObjs.value.push({
                prop: p.key,
                labelEn: p.labelEn,
                labelZh: p.labelZh,
                values: p.values,
                value: p.value,
                required: p.required,
                edit: p.edit,
                key: p.key,
                rule: p.rule,
                type: p.type,
                multiple: p.multiple,
            });
            if (p.required) {
                if (p.type === 'select') {
                    rules[p.key] = [Rules.requiredSelect];
                } else {
                    rules[p.key] = [Rules.requiredInput];
                }
            }
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
