<template>
    <LayoutContent :header="$t('commons.button.' + op)" :back-name="'Table'">
        <template #form>
            <el-form ref="ruleFormRef" label-position="left" :model="demoForm" :rules="rules" label-width="140px">
                <el-form-item :label="$t('business.user.username')" prop="username">
                    <el-input v-model="demoForm.username" />
                </el-form-item>
                <el-form-item :label="$t('business.user.email')" prop="email">
                    <el-input v-model="demoForm.email" />
                </el-form-item>
            </el-form>
            <div style="float: right">
                <el-button>{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary">{{ $t('commons.button.create') }}</el-button>
            </div>
        </template>
    </LayoutContent>
</template>
<script setup lang="ts">
import { User } from '@/api/interface/user';
import LayoutContent from '@/layout/layout-content.vue';
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rues';
const ruleFormRef = ref<FormInstance>();
interface OperateProps {
    op: string;
    id?: number;
}

withDefaults(defineProps<OperateProps>(), {
    op: 'create',
});

const demoForm = reactive<User.User>({
    id: 0,
    username: '',
    email: '',
});
const rules = reactive<FormRules>({
    username: [Rules.required, Rules.name],
    email: [Rules.required],
});
</script>
