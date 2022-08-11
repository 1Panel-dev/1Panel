<template>
    <div class="content-box">
        <el-form :model="formData" label-width="140px">
            <el-form-item label="Activity name :">
                <el-input v-model="formData.name" />
            </el-form-item>
            <el-form-item label="Activity zone :">
                <el-select v-model="formData.region" placeholder="please select your zone">
                    <el-option label="Zone one" value="shanghai" />
                    <el-option label="Zone two" value="beijing" />
                </el-select>
            </el-form-item>
            <el-form-item label="Activity time :">
                <el-date-picker v-model="formData.date1" type="date" placeholder="Pick a date" />
                <el-col :span="1" class="text-center">
                    <span class="text-gray-500">-</span>
                </el-col>
                <el-time-picker v-model="formData.date2" placeholder="Pick a time" />
            </el-form-item>
            <el-form-item label="Instant delivery :">
                <el-switch v-model="formData.delivery" />
            </el-form-item>
            <el-form-item label="Activity type :">
                <el-checkbox-group v-model="formData.type">
                    <el-checkbox label="Online activities" name="type" />
                    <el-checkbox label="Promotion activities" name="type" />
                    <el-checkbox label="Offline activities" name="type" />
                    <el-checkbox label="Simple brand exposure" name="type" />
                </el-checkbox-group>
            </el-form-item>
            <el-form-item label="Resources :">
                <el-radio-group v-model="formData.resource">
                    <el-radio label="Sponsor" />
                    <el-radio label="Venue" />
                </el-radio-group>
            </el-form-item>
            <el-form-item label="Activity form :">
                <el-input v-model="formData.desc" type="textarea" />
            </el-form-item>
            <el-form-item>
                <el-button type="primary" @click="onSubmit">Create</el-button>
                <el-button>Cancel</el-button>
            </el-form-item>
        </el-form>
    </div>
</template>

<script setup lang="ts" name="basicForm">
import { ElMessage } from 'element-plus';
import { reactive } from 'vue';

// do not use same name with ref
const formData = reactive({
    name: '',
    region: '',
    date1: '',
    date2: '',
    delivery: false,
    type: [],
    resource: '',
    desc: '',
});

const onSubmit = () => {
    ElMessage.success('提交的数据为 : ' + JSON.stringify(formData));
};
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
