<template>
    <div class="content-box">
        <el-form ref="ruleFormRef" :model="ruleForm" :rules="rules" label-width="140px">
            <el-form-item label="Activity name" prop="name">
                <el-input v-model="ruleForm.name" />
            </el-form-item>
            <el-form-item label="Activity phone" prop="phone">
                <el-input v-model="ruleForm.phone" placeholder="Activity phone" />
            </el-form-item>
            <el-form-item label="Activity zone" prop="region">
                <el-select v-model="ruleForm.region" placeholder="Activity zone">
                    <el-option label="Zone one" value="shanghai" />
                    <el-option label="Zone two" value="beijing" />
                </el-select>
            </el-form-item>
            <el-form-item label="Activity time" required>
                <el-form-item prop="date1">
                    <el-date-picker
                        v-model="ruleForm.date1"
                        type="date"
                        placeholder="Pick a date"
                        style="width: 100%"
                    />
                </el-form-item>
                <el-col class="text-center" :span="1">
                    <span class="text-gray-500">-</span>
                </el-col>
                <el-form-item prop="date2">
                    <el-time-picker v-model="ruleForm.date2" placeholder="Pick a time" style="width: 100%" />
                </el-form-item>
            </el-form-item>
            <el-form-item label="Instant delivery" prop="delivery">
                <el-switch v-model="ruleForm.delivery" />
            </el-form-item>
            <el-form-item label="Resources" prop="resource">
                <el-radio-group v-model="ruleForm.resource">
                    <el-radio label="Sponsorship" />
                    <el-radio label="Venue" />
                </el-radio-group>
            </el-form-item>
            <el-form-item label="Activity form" prop="desc">
                <el-input v-model="ruleForm.desc" type="textarea" />
            </el-form-item>
            <el-form-item>
                <el-button type="primary" @click="submitForm(ruleFormRef)">Create</el-button>
                <el-button @click="resetForm(ruleFormRef)">Reset</el-button>
            </el-form-item>
        </el-form>
    </div>
</template>

<script setup lang="ts" name="dynamicForm">
import { reactive, ref } from 'vue';
import { checkPhoneNumber } from '@/utils/eleValidate';
import type { FormInstance, FormRules } from 'element-plus';
import { ElMessage } from 'element-plus';

const ruleFormRef = ref<FormInstance>();
const ruleForm = reactive({
    name: '1Panel',
    phone: '',
    region: '',
    date1: '',
    date2: '',
    delivery: false,
    resource: '',
    desc: '',
});

const rules = reactive<FormRules>({
    name: [
        {
            required: true,
            message: 'Please input Activity name',
            trigger: 'blur',
        },
        { min: 3, max: 5, message: 'Length should be 3 to 5', trigger: 'blur' },
    ],
    phone: [{ required: true, validator: checkPhoneNumber, trigger: 'blur' }],
    region: [
        {
            required: true,
            message: 'Please select Activity zone',
            trigger: 'change',
        },
    ],
    date1: [
        {
            type: 'date',
            required: true,
            message: 'Please pick a date',
            trigger: 'change',
        },
    ],
    date2: [
        {
            type: 'date',
            required: true,
            message: 'Please pick a time',
            trigger: 'change',
        },
    ],
    resource: [
        {
            required: true,
            message: 'Please select activity resource',
            trigger: 'change',
        },
    ],
    desc: [
        {
            required: true,
            message: 'Please input activity form',
            trigger: 'blur',
        },
    ],
});

const submitForm = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid, fields) => {
        if (valid) {
            ElMessage.success('提交的数据为 : ' + JSON.stringify(ruleForm));
        } else {
            console.log('error submit!', fields);
        }
    });
};

const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.resetFields();
};
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
