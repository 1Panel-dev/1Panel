<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        size="600px"
        :title="`${drawerData.title}用户`"
    >
        <el-form
            ref="ruleFormRef"
            :rules="rules"
            :disabled="drawerData.isView"
            :model="drawerData.rowData"
            label-width="100px"
            label-suffix=" :"
            :hide-required-asterisk="drawerData.isView"
        >
            <el-form-item label="用户头像" prop="avatar">
                <UploadImg
                    v-model:imageUrl="drawerData.rowData!.avatar"
                    :disabled="drawerData.isView"
                    :upload-style="{ width: '120px', height: '120px' }"
                    @check-validate="checkValidate('avatar')"
                >
                    <template #tip> 头像大小不能超过 3M </template>
                </UploadImg>
            </el-form-item>
            <el-form-item label="用户姓名" prop="username">
                <el-input
                    v-model="drawerData.rowData!.username"
                    placeholder="请填写用户姓名"
                    clearable
                ></el-input>
            </el-form-item>
            <el-form-item label="性别" prop="gender">
                <el-select
                    v-model="drawerData.rowData!.gender"
                    placeholder="请选择性别"
                    clearable
                >
                    <el-option
                        v-for="item in genderType"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value"
                    />
                </el-select>
            </el-form-item>
            <el-form-item label="身份证号" prop="idCard">
                <el-input
                    v-model="drawerData.rowData!.idCard"
                    placeholder="请填写身份证号"
                    clearable
                ></el-input>
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
                <el-input
                    v-model="drawerData.rowData!.email"
                    placeholder="请填写邮箱"
                    clearable
                ></el-input>
            </el-form-item>
            <el-form-item label="居住地址" prop="address">
                <el-input
                    v-model="drawerData.rowData!.address"
                    placeholder="请填写居住地址"
                    clearable
                ></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false">取消</el-button>
            <el-button
                type="primary"
                v-show="!drawerData.isView"
                @click="handleSubmit"
                >确定</el-button
            >
        </template>
    </el-drawer>
</template>

<script setup lang="ts" name="UserDrawer">
import { User } from '@/api/interface';
import { ref, reactive } from 'vue';
import { genderType } from '@/utils/serviceDict';
import { ElMessage, FormInstance } from 'element-plus';
import UploadImg from '@/components/UploadImg/index.vue';

const rules = reactive({
    avatar: [{ required: true, message: '请上传用户头像', trigger: 'change' }],
    username: [
        { required: true, message: '请填写用户姓名', trigger: 'change' },
    ],
    gender: [{ required: true, message: '请选择性别', trigger: 'change' }],
    idCard: [{ required: true, message: '请填写身份证号', trigger: 'change' }],
    email: [{ required: true, message: '请填写邮箱', trigger: 'change' }],
    address: [{ required: true, message: '请填写居住地址', trigger: 'change' }],
});

interface DrawerProps {
    title: string;
    isView: boolean;
    rowData?: User.ResUserList;
    apiUrl?: (params: any) => Promise<any>;
    getTableList?: () => Promise<any>;
}

// drawer框状态
const drawerVisible = ref(false);
const drawerData = ref<DrawerProps>({
    isView: false,
    title: '',
});

// 接收父组件传过来的参数
const acceptParams = (params: DrawerProps): void => {
    drawerData.value = params;
    drawerVisible.value = true;
};

const ruleFormRef = ref<FormInstance>();
// 提交数据（新增/编辑）
const handleSubmit = () => {
    ruleFormRef.value!.validate(async (valid) => {
        if (!valid) return;
        try {
            await drawerData.value.apiUrl!(drawerData.value.rowData);
            ElMessage.success({
                message: `${drawerData.value.title}用户成功！`,
            });
            drawerData.value.getTableList!();
            drawerVisible.value = false;
        } catch (error) {
            console.log(error);
        }
    });
};

// 公共校验方法（图片上传成功触发重新校验）
const checkValidate = (val: string) => {
    ruleFormRef.value!.validateField(val, () => {});
};

defineExpose({
    acceptParams,
});
</script>
