<template>
    <div>
        <el-form ref="ruleForm" label-position="left" :model="form" label-width="100px" :rules="rules">
            <el-form-item :label="$t('file.owner')">
                <el-checkbox v-model="form.owner.r" :label="$t('file.rRole')" />
                <el-checkbox v-model="form.owner.w" :label="$t('file.wRole')" />
                <el-checkbox v-model="form.owner.x" :label="$t('file.xRole')" />
            </el-form-item>
            <el-form-item :label="$t('file.group')">
                <el-checkbox v-model="form.group.r" :label="$t('file.rRole')" />
                <el-checkbox v-model="form.group.w" :label="$t('file.wRole')" />
                <el-checkbox v-model="form.group.x" :label="$t('file.xRole')" />
            </el-form-item>
            <el-form-item :label="$t('file.public')">
                <el-checkbox v-model="form.public.r" :label="$t('file.rRole')" />
                <el-checkbox v-model="form.public.w" :label="$t('file.wRole')" />
                <el-checkbox v-model="form.public.x" :label="$t('file.xRole')" />
            </el-form-item>
            <el-form-item :label="$t('file.role')" required prop="mode">
                <el-input v-model="form.mode" maxlength="4" @input="changeMode"></el-input>
            </el-form-item>
        </el-form>
    </div>
</template>
<script setup lang="ts">
import { FormInstance, FormRules } from 'element-plus';
import { computed, ref, toRefs, watch, onUpdated, onMounted, reactive } from 'vue';
import { Rules } from '@/global/form-rules';

interface Role {
    r: boolean;
    w: boolean;
    x: boolean;
}
interface RoleForm {
    owner: Role;
    group: Role;
    public: Role;
    mode: string;
}
interface Props {
    mode: string;
}

const roles = ref<string[]>(['0', '1', '2', '3', '4', '5', '6', '7']);

const props = withDefaults(defineProps<Props>(), {
    mode: '0755',
});
const rules = reactive<FormRules>({
    mode: [Rules.requiredInput, Rules.filePermission],
});

const { mode } = toRefs(props);
const ruleForm = ref<FormInstance>();
let form = ref<RoleForm>({
    owner: { r: true, w: true, x: true },
    group: { r: true, w: true, x: true },
    public: { r: true, w: false, x: true },
    mode: '0755',
});
const em = defineEmits(['getMode']);

const calculate = (role: Role) => {
    let num = 0;
    if (role.r) {
        num = num + 4;
    }
    if (role.w) {
        num = num + 2;
    }
    if (role.x) {
        num = num + 1;
    }

    return num;
};

const getRole = computed(() => {
    const value =
        '0' +
        String(calculate(form.value.owner)) +
        String(calculate(form.value.group)) +
        String(calculate(form.value.public));
    return value;
});

watch(
    () => getRole.value,
    (newVal) => {
        form.value.mode = newVal;
    },
);

watch(
    () => form.value.mode,
    (newVal) => {
        em('getMode', Number.parseInt(newVal, 8));
    },
);

const getRoleNum = (roleStr: string, role: Role) => {
    if (roles.value.indexOf(roleStr) < 0) {
        return;
    }
    switch (roleStr) {
        case '0':
            role.x = false;
            role.w = false;
            role.r = false;
            break;
        case '1':
            role.x = true;
            role.w = false;
            role.r = false;
            break;
        case '2':
            role.x = false;
            role.w = true;
            role.r = false;
            break;
        case '3':
            role.x = true;
            role.w = true;
            role.r = false;
            break;
        case '4':
            role.x = false;
            role.w = false;
            role.r = true;
            break;
        case '5':
            role.x = true;
            role.w = false;
            role.r = true;
            break;
        case '6':
            role.x = false;
            role.w = true;
            role.r = true;
            break;
        case '7':
            role.x = true;
            role.w = true;
            role.r = true;
            break;
    }
};

const changeMode = (val: String) => {
    if (val === '' || val.length !== 4) {
        return;
    }
    getRoleNum(val[1], form.value.owner);
    getRoleNum(val[2], form.value.group);
    getRoleNum(val[3], form.value.public);
};

const updateMode = () => {
    form.value.mode = mode.value;
    changeMode(form.value.mode);
};

onUpdated(() => {
    updateMode();
});

onMounted(() => {
    updateMode();
});
</script>
