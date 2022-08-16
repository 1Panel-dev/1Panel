<template>
    <LayoutContent :header="$t('commons.button.' + op)" :back-name="'Table'">
        <template #form>
            <el-form ref="ruleFormRef" label-position="left" :model="demoForm" :rules="rules" label-width="140px">
                <el-form-item :label="$t('business.user.username')" prop="name">
                    <el-input v-model="demoForm.name" />
                </el-form-item>
                <el-form-item :label="$t('business.user.email')" prop="email">
                    <el-input v-model="demoForm.email" />
                </el-form-item>
                <el-form-item :label="$t('business.user.password')" prop="password">
                    <el-input type="password" v-model="demoForm.password" />
                </el-form-item>
            </el-form>
            <div class="form-button">
                <el-button @click="router.back()">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submitForm(ruleFormRef)">{{
                    $t('commons.button.confirm')
                }}</el-button>
            </div>
        </template>
    </LayoutContent>
</template>
<script setup lang="ts">
import LayoutContent from '@/layout/layout-content.vue';
import { ElMessage, FormInstance, FormRules } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { Rules } from '@/global/form-rues';
import { addUser, editUser, getUserById } from '@/api/modules/user';
import i18n from '@/lang';
import { useRouter } from 'vue-router';
import { User } from '@/api/interface/user';
const router = useRouter();
const ruleFormRef = ref<FormInstance>();
let demoForm = ref<User.User>({
    id: 0,
    name: '',
    email: '',
    password: '',
});

interface OperateProps {
    op: string;
    id: string;
}

const props = withDefaults(defineProps<OperateProps>(), {
    op: 'create',
});

const rules = reactive<FormRules>({
    name: [Rules.required, Rules.name],
    email: [Rules.required, Rules.email],
    password: [Rules.required],
});

const submitForm = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        if (props.op === 'create') {
            addUser(demoForm.value).then(() => {
                ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
                router.back();
            });
        } else {
            editUser(demoForm.value).then(() => {
                ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
                router.back();
            });
        }
    });
};

const getUser = async (id: number) => {
    const res = await getUserById(id);
    demoForm.value = res.data;
};
onMounted(() => {
    if (props.op == 'edit') {
        console.log(props);
        getUser(Number(props.id)).catch(() => {
            router.back();
        });
    }
});
</script>
